// Package achievement evaluates and awards achievements to players.
// Each achievement is defined in its own file (a<id>_<slug>.go) via init().
// Achievements are NEVER revoked once earned — only new ones are added per run.
// player_achievements is the authoritative table; achievement_bitstring is a cache.
package achievement

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ── Public types used by individual definition files ─────────────────────────

// G holds per-game data for game-based achievement checks.
type G struct {
	GameID           int
	VictoryType      string
	Leader           string
	Turns            int
	Score            int
	Favor            int
	Science          int
	Culture          int
	Winner           bool
	PlayerCount      int
	TeamCount        int
	PreRatingRank    int  // 1 = weakest entering
	PostRatingRank   int  // 1 = strongest rating after game
	PostRatingOverall float64
	PostRDOverall     float64
	WinStreakAfter    int
	ScoreRank        int  // 1 = highest score
	MiningResearched *bool
	EnemyLeaders     []string
	LosingStreakBefore int // consecutive losses immediately before this game
}

// S holds cumulative stats for stats-based achievement checks.
type S struct {
	Rating           float64
	RD               float64
	Rank             int
	HighestWinstreak int
}

// HasEnemy returns true if any opponent played the given leader (case-insensitive).
func (g G) HasEnemy(leader string) bool {
	for _, l := range g.EnemyLeaders {
		if strings.EqualFold(l, leader) {
			return true
		}
	}
	return false
}

// ── Registry ──────────────────────────────────────────────────────────────────

type def struct {
	id         int
	checkGame  func(G) bool // nil = not game-based
	checkStats func(S) bool // nil = not stats-based
}

var registry = map[int]*def{}

// RegisterGame registers a per-game achievement. Called from init() in definition files.
func RegisterGame(id int, fn func(G) bool) {
	registry[id] = &def{id: id, checkGame: fn}
}

// RegisterStats registers a cumulative-stats achievement.
func RegisterStats(id int, fn func(S) bool) {
	registry[id] = &def{id: id, checkStats: fn}
}

// ── Evaluator ─────────────────────────────────────────────────────────────────

var pointsFor = map[int]int{-1: 200, 0: 50, 1: 100, 2: 150, 3: 200}

type achMeta struct{ id, difficulty int }

func RecalculateAll(ctx context.Context, pool *pgxpool.Pool) error {
	// Clear all legacy/unverified achievements to allow clean chronological reconstruction from history.
	if _, err := pool.Exec(ctx, `DELETE FROM player_achievements`); err != nil {
		return fmt.Errorf("clear achievements: %w", err)
	}

	rows, err := pool.Query(ctx, `SELECT id, difficulty FROM achievements WHERE disabled = false ORDER BY id`)
	if err != nil {
		return fmt.Errorf("fetch achievements: %w", err)
	}
	defer rows.Close()
	var metas []achMeta
	for rows.Next() {
		var m achMeta
		rows.Scan(&m.id, &m.difficulty)
		metas = append(metas, m)
	}

	pRows, err := pool.Query(ctx, `SELECT id FROM players`)
	if err != nil {
		return fmt.Errorf("fetch players: %w", err)
	}
	defer pRows.Close()
	var pids []int
	for pRows.Next() {
		var id int
		pRows.Scan(&id)
		pids = append(pids, id)
	}

	for _, pid := range pids {
		if err := evaluate(ctx, pool, pid, metas); err != nil {
			return fmt.Errorf("player %d: %w", pid, err)
		}
	}
	return nil
}

func evaluate(ctx context.Context, pool *pgxpool.Pool, playerID int, metas []achMeta) error {
	// Game history.
	games, err := loadGames(ctx, pool, playerID)
	if err != nil {
		return err
	}

	// Stats.
	stats, err := loadStats(ctx, pool, playerID)
	if err != nil {
		return err
	}

	for _, m := range metas {
		d := registry[m.id]
		if d == nil {
			continue // no definition registered for this id
		}

		// Evaluate stats-based (only if they don't have a game-based check).
		if d.checkStats != nil && d.checkStats(stats) {
			award(ctx, pool, playerID, m.id, nil)
			continue
		}

		// Evaluate game-based (chronologically).
		if d.checkGame != nil {
			for _, g := range games {
				if d.checkGame(g) {
					gid := g.GameID
					award(ctx, pool, playerID, m.id, &gid)
					break // award first matching game and stop
				}
			}
		}
	}

	// Recompute bitstring + points from authoritative table.
	_, err = pool.Exec(ctx, `
		UPDATE players p SET
			achievement_bitstring = COALESCE((
				SELECT bit_or(1::bigint << (pa.achievement_id - 1))
				FROM player_achievements pa WHERE pa.player_id = p.id
			), 0),
			achievement_points = COALESCE((
				SELECT SUM(CASE a.difficulty WHEN -1 THEN 200 WHEN 0 THEN 50 WHEN 1 THEN 100 WHEN 2 THEN 150 WHEN 3 THEN 200 ELSE 50 END)
				FROM player_achievements pa
				JOIN achievements a ON a.id = pa.achievement_id
				WHERE pa.player_id = p.id AND a.disabled = false
			), 0)
		WHERE p.id = $1`, playerID)
	return err
}

func award(ctx context.Context, pool *pgxpool.Pool, playerID, achID int, gameID *int) {
	pool.Exec(ctx, `INSERT INTO player_achievements (player_id, achievement_id, game_id) VALUES ($1,$2,$3) ON CONFLICT DO NOTHING`, playerID, achID, gameID)
}

// ── Data loaders ──────────────────────────────────────────────────────────────

func loadGames(ctx context.Context, pool *pgxpool.Pool, playerID int) ([]G, error) {
	rows, err := pool.Query(ctx, `
		SELECT
			g.id, g.victory_type,
			COALESCE(gp.leader,''), COALESCE(g.turns,0),
			COALESCE(gp.score,0), COALESCE(gp.favor,0),
			COALESCE(gp.science,0), COALESCE(gp.culture,0),
			gp.winner,
			(SELECT COUNT(*) FROM game_players WHERE game_id=g.id)::int,
			(SELECT COUNT(DISTINCT team) FROM game_players WHERE game_id=g.id)::int,
			(SELECT COUNT(*) FROM game_players gp2 WHERE gp2.game_id=g.id AND gp2.pre_rating_overall < gp.pre_rating_overall)::int + 1,
			(SELECT COUNT(*) FROM game_players gp2 WHERE gp2.game_id=g.id AND gp2.post_rating_overall > gp.post_rating_overall)::int + 1,
			COALESCE(gp.post_rating_overall, 0.0),
			COALESCE(gp.post_rd_overall, 0.0),
			(SELECT COUNT(*) FROM game_players gp2 WHERE gp2.game_id=g.id AND COALESCE(gp2.score,0) > COALESCE(gp.score,0))::int + 1,
			gp.mining_researched
		FROM games g
		JOIN game_players gp ON gp.game_id=g.id AND gp.player_id=$1
		WHERE g.tmp=false
		ORDER BY g.date ASC
	`, playerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var games []G
	var losingStreak int
	var winStreak int
	for rows.Next() {
		var gm G
		rows.Scan(
			&gm.GameID, &gm.VictoryType, &gm.Leader, &gm.Turns,
			&gm.Score, &gm.Favor, &gm.Science, &gm.Culture, &gm.Winner,
			&gm.PlayerCount, &gm.TeamCount,
			&gm.PreRatingRank, &gm.PostRatingRank,
			&gm.PostRatingOverall, &gm.PostRDOverall,
			&gm.ScoreRank,
			&gm.MiningResearched,
		)
		gm.LosingStreakBefore = losingStreak
		if gm.Winner {
			losingStreak = 0
			winStreak++
		} else {
			losingStreak++
			winStreak = 0
		}
		gm.WinStreakAfter = winStreak

		// enemy leaders
		er, _ := pool.Query(ctx, `SELECT COALESCE(leader,'') FROM game_players WHERE game_id=$1 AND player_id!=$2`, gm.GameID, playerID)
		if er != nil {
			for er.Next() {
				var l string
				er.Scan(&l)
				gm.EnemyLeaders = append(gm.EnemyLeaders, l)
			}
			er.Close()
		}
		games = append(games, gm)
	}
	return games, rows.Err()
}

func loadStats(ctx context.Context, pool *pgxpool.Pool, playerID int) (S, error) {
	var s S
	pool.QueryRow(ctx, `SELECT COALESCE(rating,1500), COALESCE(rd,350) FROM player_ratings WHERE player_id=$1 AND category='overall'`, playerID).Scan(&s.Rating, &s.RD)
	pool.QueryRow(ctx, `SELECT COUNT(*)+1 FROM player_ratings pr JOIN players p ON p.id=pr.player_id AND p.active=true WHERE pr.category='overall' AND pr.rating>(SELECT COALESCE(rating,1500) FROM player_ratings WHERE player_id=$1 AND category='overall')`, playerID).Scan(&s.Rank)
	pool.QueryRow(ctx, `SELECT COALESCE(highest_winstreak,0) FROM player_stats WHERE player_id=$1 AND category='overall'`, playerID).Scan(&s.HighestWinstreak)
	return s, nil
}
