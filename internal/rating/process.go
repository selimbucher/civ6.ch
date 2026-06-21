package rating

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/selimbucher/civ6.ch/internal/glicko"
)

const (
	defaultRating     = 1500.0
	defaultRD         = 150.0
	defaultVolatility = 0.06
	maxRD             = 150.0
	tau               = 0.25

	// denounceAmplify scales the rating change of a matchup between two players
	// with an active denouncement (in either direction) between them.
	denounceAmplify = 1.5

	// teamSizeBonus is the effective rating advantage credited to a team that
	// outnumbers the smallest team, scaled by the size ratio (len/smallest − 1)
	// rather than the raw player difference: outnumbering 2:1 (2v1) is a far
	// bigger edge than 3:2, even though both add one player. A team with twice
	// the players gets the full bonus; 1.5× gets half. At 100, a 2:1 advantage
	// ≈ a 64% expected win. This replaces the old sum-of-ratings handicap, which
	// inflated a larger team's rating by ~1500 per extra player (a near-certain
	// win) and so punished bigger teams far too harshly.
	teamSizeBonus = 100.0
)

type gameRow struct {
	id       int
	category string
	date     time.Time
	rated    bool
	tmp      bool
	draw     bool
	weight   float64
}

type gamePlayerRow struct {
	playerID int
	team     int
	winner   bool
	score    int
	leader   string
}

type ratingRow struct {
	rating     float64
	rd         float64
	volatility float64
	lastPlayed *time.Time
}

func fetchGame(ctx context.Context, pool *pgxpool.Pool, gameID int) (gameRow, error) {
	var g gameRow
	err := pool.QueryRow(ctx, `
		SELECT id, category, date, rated, tmp, draw, weight
		FROM games WHERE id = $1
	`, gameID).Scan(&g.id, &g.category, &g.date, &g.rated, &g.tmp, &g.draw, &g.weight)
	return g, err
}

func fetchGamePlayers(ctx context.Context, pool *pgxpool.Pool, gameID int) ([]gamePlayerRow, error) {
	rows, err := pool.Query(ctx, `
		SELECT player_id, team, winner, COALESCE(score, 0), COALESCE(leader, '')
		FROM game_players WHERE game_id = $1
	`, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var players []gamePlayerRow
	for rows.Next() {
		var p gamePlayerRow
		if err := rows.Scan(&p.playerID, &p.team, &p.winner, &p.score, &p.leader); err != nil {
			return nil, err
		}
		players = append(players, p)
	}
	return players, rows.Err()
}

func fetchRating(ctx context.Context, pool *pgxpool.Pool, playerID int, category string) (ratingRow, error) {
	var r ratingRow
	err := pool.QueryRow(ctx, `
		SELECT rating, rd, volatility, last_played
		FROM player_ratings
		WHERE player_id = $1 AND category = $2
	`, playerID, category).Scan(&r.rating, &r.rd, &r.volatility, &r.lastPlayed)
	if err != nil {
		// no row yet — return defaults
		return ratingRow{
			rating:     defaultRating,
			rd:         defaultRD,
			volatility: defaultVolatility,
		}, nil
	}
	return r, nil
}

func upsertRating(ctx context.Context, pool *pgxpool.Pool, playerID int, category string, p glicko.Player, lastPlayed time.Time) error {
	_, err := pool.Exec(ctx, `
		INSERT INTO player_ratings (player_id, category, rating, rd, volatility, last_played)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (player_id, category) DO UPDATE
		SET rating = $3, rd = $4, volatility = $5, last_played = $6
	`, playerID, category, p.Rating, p.RD, p.Volatility, lastPlayed)
	return err
}

func fetchStats(ctx context.Context, pool *pgxpool.Pool, playerID int, category string) (games, wins, streak, highestWinstreak int, err error) {
	err = pool.QueryRow(ctx, `
		SELECT games, wins, streak, highest_winstreak
		FROM player_stats
		WHERE player_id = $1 AND category = $2
	`, playerID, category).Scan(&games, &wins, &streak, &highestWinstreak)
	if err != nil {
		return 0, 0, 0, 0, nil // no row yet, defaults
	}
	return
}

func upsertStats(ctx context.Context, pool *pgxpool.Pool, playerID int, category string, games, wins, streak, highestWinstreak int) error {
	_, err := pool.Exec(ctx, `
		INSERT INTO player_stats (player_id, category, games, wins, streak, highest_winstreak)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (player_id, category) DO UPDATE
		SET games = $3, wins = $4, streak = $5, highest_winstreak = $6
	`, playerID, category, games, wins, streak, highestWinstreak)
	return err
}

func updateGamePlayer(ctx context.Context, pool *pgxpool.Pool, gameID, playerID int,
	pre, post, preOverall, postOverall glicko.Player) error {
	_, err := pool.Exec(ctx, `
		UPDATE game_players SET
			pre_rating = $3, pre_rd = $4, pre_volatility = $5,
			post_rating = $6, post_rd = $7, post_volatility = $8,
			pre_rating_overall = $9, pre_rd_overall = $10, pre_volatility_overall = $11,
			post_rating_overall = $12, post_rd_overall = $13, post_volatility_overall = $14
		WHERE game_id = $1 AND player_id = $2
	`, gameID, playerID,
		pre.Rating, pre.RD, pre.Volatility,
		post.Rating, post.RD, post.Volatility,
		preOverall.Rating, preOverall.RD, preOverall.Volatility,
		postOverall.Rating, postOverall.RD, postOverall.Volatility,
	)
	return err
}

// orderedPair returns the two ids in ascending order, so a denouncement in
// either direction maps to the same key.
func orderedPair(a, b int) [2]int {
	if a > b {
		a, b = b, a
	}
	return [2]int{a, b}
}

// activeDenouncements returns the set of unordered player pairs (among the given
// players) that had an active denouncement — in either direction — as of asOf.
// State is reconstructed from the append-only denouncement_events log so that
// rating recalculation is reproducible regardless of later forgive/denounce
// actions: a pair is active when the most recent event at or before asOf for
// some direction is a 'denounce'.
func activeDenouncements(ctx context.Context, pool *pgxpool.Pool, playerIDs []int, asOf time.Time) (map[[2]int]bool, error) {
	pairs := map[[2]int]bool{}
	rows, err := pool.Query(ctx, `
		SELECT denouncer_id, denounced_id FROM (
			SELECT DISTINCT ON (denouncer_id, denounced_id)
			       denouncer_id, denounced_id, action
			FROM denouncement_events
			WHERE created_at <= $1
			  AND denouncer_id = ANY($2)
			  AND denounced_id = ANY($2)
			ORDER BY denouncer_id, denounced_id, created_at DESC, id DESC
		) latest
		WHERE action = 'denounce'
	`, asOf, playerIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var a, b int
		if err := rows.Scan(&a, &b); err != nil {
			return nil, err
		}
		pairs[orderedPair(a, b)] = true
	}
	return pairs, rows.Err()
}

func teamScore(players []gamePlayerRow, team int) int {
	s := 0
	for _, p := range players {
		if p.team == team {
			s += p.score
		}
	}
	return s
}

func teamWon(players []gamePlayerRow, team int) bool {
	for _, p := range players {
		if p.team == team {
			return p.winner
		}
	}
	return false
}

// aggregateRatings combines the members' ratings into one team aggregate.
// The team's skill is the members' average rating, plus a teamSizeBonus scaled
// by how far it outnumbers the smallest team (len/smallest − 1), so a 2:1 edge
// counts double a 3:2 one. RD is averaged over the members. get selects which
// rating row (category or overall) to aggregate.
//
// Averaging rather than summing matters: Glicko ratings sit on a scale whose
// origin (1500) is arbitrary, so summing them treats "two average players" as a
// single 3000-rated entity — a near-certain favourite — which made larger teams
// gain almost nothing for winning and bleed rating for losing.
func aggregateRatings(members []gamePlayerRow, smallestTeam int, get func(playerID int) ratingRow) (rating, rd float64) {
	for _, m := range members {
		rating += get(m.playerID).rating
	}
	rating /= float64(len(members))
	rating += teamSizeBonus * (float64(len(members))/float64(smallestTeam) - 1)
	for _, m := range members {
		rd += get(m.playerID).rd
	}
	rd /= float64(len(members))
	return rating, rd
}

func ProcessGame(ctx context.Context, pool *pgxpool.Pool, gameID int, decayRD bool) error {
	game, err := fetchGame(ctx, pool, gameID)
	if err != nil {
		return fmt.Errorf("fetch game: %w", err)
	}
	if !game.rated || game.tmp {
		return nil
	}

	players, err := fetchGamePlayers(ctx, pool, gameID)
	if err != nil {
		return fmt.Errorf("fetch game players: %w", err)
	}

	// group players by team
	teams := map[int][]gamePlayerRow{}
	playerIDs := make([]int, 0, len(players))
	for _, p := range players {
		teams[p.team] = append(teams[p.team], p)
		playerIDs = append(playerIDs, p.playerID)
	}

	// active grudges as of this game's date, for matchup amplification
	denouncePairs, err := activeDenouncements(ctx, pool, playerIDs, game.date)
	if err != nil {
		return fmt.Errorf("fetch denouncements: %w", err)
	}

	// find smallest team size for normalization
	smallestTeam := -1
	for _, members := range teams {
		if smallestTeam == -1 || len(members) < smallestTeam {
			smallestTeam = len(members)
		}
	}

	// fetch current ratings for all players
	type playerState struct {
		cat     ratingRow
		overall ratingRow
	}
	states := map[int]playerState{}
	for _, p := range players {
		cat, err := fetchRating(ctx, pool, p.playerID, game.category)
		if err != nil {
			return fmt.Errorf("fetch rating: %w", err)
		}
		overall, err := fetchRating(ctx, pool, p.playerID, "overall")
		if err != nil {
			return fmt.Errorf("fetch overall rating: %w", err)
		}
		states[p.playerID] = playerState{cat: cat, overall: overall}
	}

	catOf := func(playerID int) ratingRow { return states[playerID].cat }
	overallOf := func(playerID int) ratingRow { return states[playerID].overall }

	// apply RD decay if requested
	if decayRD {
		for pid, state := range states {
			if state.cat.lastPlayed != nil {
				days := game.date.Sub(*state.cat.lastPlayed).Hours() / 24
				state.cat.rd = glicko.DecayRD(state.cat.rd, state.cat.volatility, days, maxRD)
			}
			if state.overall.lastPlayed != nil {
				days := game.date.Sub(*state.overall.lastPlayed).Hours() / 24
				state.overall.rd = glicko.DecayRD(state.overall.rd, state.overall.volatility, days, maxRD)
			}
			states[pid] = state
		}
	}

	// compute new ratings for each team
	newCat := map[int]glicko.Player{}
	newOverall := map[int]glicko.Player{}

	for teamID, members := range teams {
		// aggregate this team's rating
		teamRating, teamRD := aggregateRatings(members, smallestTeam, catOf)
		teamRatingOverall, teamRDOverall := aggregateRatings(members, smallestTeam, overallOf)

		// aggregate opponents (oppTeamIDs is kept parallel to the opp/result
		// slices so per-member grudge weights can be computed below)
		var oppsCat []glicko.Opponent
		var oppsOverall []glicko.Opponent
		var results []float64
		var oppTeamIDs []int

		myScore := teamScore(players, teamID)
		myWon := teamWon(players, teamID)

		for oppID, oppMembers := range teams {
			if oppID == teamID {
				continue
			}
			oppRating, oppRD := aggregateRatings(oppMembers, smallestTeam, catOf)
			oppRatingOverall, oppRDOverall := aggregateRatings(oppMembers, smallestTeam, overallOf)

			oppScore := teamScore(players, oppID)
			result := glicko.Result(myScore, oppScore, myWon, teamWon(players, oppID))

			oppsCat = append(oppsCat, glicko.Opponent{Rating: oppRating, RD: oppRD})
			oppsOverall = append(oppsOverall, glicko.Opponent{Rating: oppRatingOverall, RD: oppRDOverall})
			results = append(results, result)
			oppTeamIDs = append(oppTeamIDs, oppID)
		}

		ts := len(members)
		for _, m := range members {
			pre := glicko.Player{
				Rating:     states[m.playerID].cat.rating,
				RD:         states[m.playerID].cat.rd,
				Volatility: states[m.playerID].cat.volatility,
			}
			preOverall := glicko.Player{
				Rating:     states[m.playerID].overall.rating,
				RD:         states[m.playerID].overall.rd,
				Volatility: states[m.playerID].overall.volatility,
			}

			// amplify matchups against teams holding a denounced rival
			weights := make([]float64, len(oppTeamIDs))
			for i, oppID := range oppTeamIDs {
				weights[i] = 1.0
				for _, o := range teams[oppID] {
					if denouncePairs[orderedPair(m.playerID, o.playerID)] {
						weights[i] = denounceAmplify
						break
					}
				}
			}

			newCat[m.playerID] = glicko.Update(pre, teamRating, teamRD, oppsCat, results, weights, ts, tau)
			newOverall[m.playerID] = glicko.Update(preOverall, teamRatingOverall, teamRDOverall, oppsOverall, results, weights, ts, tau)
		}
	}

	// write results
	for _, p := range players {
		pre := glicko.Player{
			Rating:     states[p.playerID].cat.rating,
			RD:         states[p.playerID].cat.rd,
			Volatility: states[p.playerID].cat.volatility,
		}
		preOverall := glicko.Player{
			Rating:     states[p.playerID].overall.rating,
			RD:         states[p.playerID].overall.rd,
			Volatility: states[p.playerID].overall.volatility,
		}
		if err := updateGamePlayer(ctx, pool, gameID, p.playerID, pre, newCat[p.playerID], preOverall, newOverall[p.playerID]); err != nil {
			return fmt.Errorf("update game player: %w", err)
		}
		if err := upsertRating(ctx, pool, p.playerID, game.category, newCat[p.playerID], game.date); err != nil {
			return fmt.Errorf("upsert category rating: %w", err)
		}
		if err := upsertRating(ctx, pool, p.playerID, "overall", newOverall[p.playerID], game.date); err != nil {
			return fmt.Errorf("upsert overall rating: %w", err)
		}

		won := false
		for _, gp := range players {
			if gp.playerID == p.playerID && gp.winner {
				won = true
				break
			}
		}

		catGames, catWins, catStreak, catHighest, err := fetchStats(ctx, pool, p.playerID, game.category)
		if err != nil {
			return fmt.Errorf("fetch stats: %w", err)
		}
		overallGames, overallWins, overallStreak, overallHighest, err := fetchStats(ctx, pool, p.playerID, "overall")
		if err != nil {
			return fmt.Errorf("fetch overall stats: %w", err)
		}

		catGames++
		overallGames++
		if won {
			catWins++
			overallWins++
			catStreak++
			overallStreak++
		} else {
			catStreak = 0
			overallStreak = 0
		}
		if catStreak > catHighest {
			catHighest = catStreak
		}
		if overallStreak > overallHighest {
			overallHighest = overallStreak
		}

		if err := upsertStats(ctx, pool, p.playerID, game.category, catGames, catWins, catStreak, catHighest); err != nil {
			return fmt.Errorf("upsert category stats: %w", err)
		}
		if err := upsertStats(ctx, pool, p.playerID, "overall", overallGames, overallWins, overallStreak, overallHighest); err != nil {
			return fmt.Errorf("upsert overall stats: %w", err)
		}
	}

	return nil
}
