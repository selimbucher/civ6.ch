ALTER TABLE games ADD COLUMN IF NOT EXISTS save_hash TEXT;
-- Unique index on non-null hashes prevents duplicate saves from being inserted twice.
CREATE UNIQUE INDEX IF NOT EXISTS games_save_hash_unique
    ON games (save_hash)
    WHERE save_hash IS NOT NULL;
