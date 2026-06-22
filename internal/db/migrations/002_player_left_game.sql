-- Distinguish players who *left* a game (and were taken over by the AI) from
-- players who were genuinely *eliminated*.
--
-- The parser marks both the same way (the human slot flips to AI), but a player
-- who left still controls a living civ with cities, while an eliminated player
-- has none. game_players.eliminated now means "eliminated" only; left_game
-- carries the "left and became a bot" case so the UI can mark it differently and
-- so a left player keeps their real score instead of being zeroed.

ALTER TABLE game_players
    ADD COLUMN IF NOT EXISTS left_game boolean NOT NULL DEFAULT false;
