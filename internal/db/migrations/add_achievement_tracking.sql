-- Authoritative record of which player earned which achievement and in which game.
-- game_id is NULL for stats-based achievements (e.g. "reach rating 1600").
CREATE TABLE IF NOT EXISTS player_achievements (
    player_id    INTEGER  NOT NULL REFERENCES players(id),
    achievement_id INTEGER NOT NULL REFERENCES achievements(id),
    game_id      INTEGER  REFERENCES games(id),
    earned_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (player_id, achievement_id)
);

-- Backfill from existing bitstrings (game_id unknown for legacy data).
INSERT INTO player_achievements (player_id, achievement_id, earned_at)
SELECT p.id, a.id, NOW()
FROM players p
JOIN achievements a ON (p.achievement_bitstring & (1::bigint << (a.id - 1))) <> 0
ON CONFLICT DO NOTHING;

-- Mining research flag for the El Salvador achievement.
ALTER TABLE game_players ADD COLUMN IF NOT EXISTS mining_researched BOOLEAN NOT NULL DEFAULT FALSE;
