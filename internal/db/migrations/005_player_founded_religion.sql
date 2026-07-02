-- Each player's founded religion (if any), so a match can show who founded what.
-- Name may be NULL for the pre-baked religions that can't be renamed (Islam, …);
-- the icon key then supplies the display name. Colour/icon mirror the per-city
-- columns so the founded religion renders with its real symbol and colour.

ALTER TABLE game_players
    ADD COLUMN IF NOT EXISTS founded_religion       text,
    ADD COLUMN IF NOT EXISTS founded_religion_icon  text,
    ADD COLUMN IF NOT EXISTS founded_religion_color text;
