-- Store each religion's display colour (the colour the founding player chose,
-- read from the save) so the religion symbol can be tinted with its real,
-- distinct colour instead of a flat monochrome glyph. Hex string like "#7bff61";
-- nullable for cities with no religion.

ALTER TABLE game_player_cities
    ADD COLUMN IF NOT EXISTS religion_color text;
