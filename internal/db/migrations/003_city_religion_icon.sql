-- Store the religion's icon key (e.g. "Custom10", "Islam") alongside its
-- player-chosen name. The key comes from the religion's RELIGION_* type in the
-- save, so each distinct religion renders its real, distinct symbol instead of
-- a single generic glyph. Nullable: cities with no majority religion stay NULL.

ALTER TABLE game_player_cities
    ADD COLUMN IF NOT EXISTS religion_icon text;
