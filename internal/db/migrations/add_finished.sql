-- Distinguishes finalized games from ongoing ones.
-- All existing confirmed games (tmp=false) are finished by definition.
ALTER TABLE games ADD COLUMN IF NOT EXISTS finished BOOLEAN NOT NULL DEFAULT TRUE;
