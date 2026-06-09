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
)

type gameRow struct {
	id         int
	category   string
	date       time.Time
	rated      bool
	tmp        bool
	draw       bool
	weight     float64
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
	for _, p := range players {
		teams[p.team] = append(teams[p.team], p)
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
		teamRating := 0.0
		teamRD := 0.0
		for _, m := range members {
			teamRating += states[m.playerID].cat.rating
		}
		teamRating /= float64(smallestTeam)
		for _, m := range members {
			teamRD += states[m.playerID].cat.rd
		}
		teamRD /= float64(len(members))

		teamRatingOverall := 0.0
		teamRDOverall := 0.0
		for _, m := range members {
			teamRatingOverall += states[m.playerID].overall.rating
		}
		teamRatingOverall /= float64(smallestTeam)
		for _, m := range members {
			teamRDOverall += states[m.playerID].overall.rd
		}
		teamRDOverall /= float64(len(members))

		// aggregate opponents
		var oppsCat []glicko.Opponent
		var oppsOverall []glicko.Opponent
		var results []float64

		myScore := teamScore(players, teamID)
		myWon := teamWon(players, teamID)

		for oppID, oppMembers := range teams {
			if oppID == teamID {
				continue
			}
			oppRating := 0.0
			oppRD := 0.0
			for _, m := range oppMembers {
				oppRating += states[m.playerID].cat.rating
			}
			oppRating /= float64(smallestTeam)
			for _, m := range oppMembers {
				oppRD += states[m.playerID].cat.rd
			}
			oppRD /= float64(len(oppMembers))

			oppRatingOverall := 0.0
			oppRDOverall := 0.0
			for _, m := range oppMembers {
				oppRatingOverall += states[m.playerID].overall.rating
			}
			oppRatingOverall /= float64(smallestTeam)
			for _, m := range oppMembers {
				oppRDOverall += states[m.playerID].overall.rd
			}
			oppRDOverall /= float64(len(oppMembers))

			oppScore := teamScore(players, oppID)
			result := glicko.Result(myScore, oppScore, myWon, teamWon(players, oppID))

			oppsCat = append(oppsCat, glicko.Opponent{Rating: oppRating, RD: oppRD})
			oppsOverall = append(oppsOverall, glicko.Opponent{Rating: oppRatingOverall, RD: oppRDOverall})
			results = append(results, result)
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
			newCat[m.playerID] = glicko.Update(pre, teamRating, teamRD, oppsCat, results, ts, tau)
			newOverall[m.playerID] = glicko.Update(preOverall, teamRatingOverall, teamRDOverall, oppsOverall, results, ts, tau)
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