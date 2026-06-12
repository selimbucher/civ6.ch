package rating

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/selimbucher/civ6.ch/internal/achievement"
)

func RecalculateAll(ctx context.Context, pool *pgxpool.Pool) error {
	// Reset ratings and stats (not achievements — those are never revoked).
	var err error
	_, err = pool.Exec(ctx, `
		UPDATE player_ratings SET rating=$1, rd=$2, volatility=$3, last_played=NULL
	`, defaultRating, defaultRD, defaultVolatility)
	if err != nil {
		return fmt.Errorf("reset ratings: %w", err)
	}
	_, err = pool.Exec(ctx, `UPDATE player_stats SET games=0, wins=0, streak=0, highest_winstreak=0`)
	if err != nil {
		return fmt.Errorf("reset stats: %w", err)
	}
	_, err = pool.Exec(ctx, `
		UPDATE game_players SET
			pre_rating=0, pre_rd=0, pre_volatility=0,
			post_rating=0, post_rd=0, post_volatility=0,
			pre_rating_overall=0, pre_rd_overall=0, pre_volatility_overall=0,
			post_rating_overall=0, post_rd_overall=0, post_volatility_overall=0
	`)
	if err != nil {
		return fmt.Errorf("reset game players: %w", err)
	}

	// Fetch all rated and confirmed games ordered by date
	rows, err := pool.Query(ctx, `
		SELECT id FROM games
		WHERE rated = TRUE AND tmp = FALSE
		ORDER BY date ASC
	`)
	if err != nil {
		return fmt.Errorf("fetch games: %w", err)
	}
	defer rows.Close()

	var gameIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return fmt.Errorf("scan game id: %w", err)
		}
		gameIDs = append(gameIDs, id)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate games: %w", err)
	}

	// Process each game in order
	for _, id := range gameIDs {
		if err := ProcessGame(ctx, pool, id, true); err != nil {
			return fmt.Errorf("process game %d: %w", id, err)
		}
	}

	// Re-award achievements based on final ratings + full game history.
	if err := achievement.RecalculateAll(ctx, pool); err != nil {
		return fmt.Errorf("recalculate achievements: %w", err)
	}

	return nil
}
