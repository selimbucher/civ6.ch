-- Append-only log of denounce / forgive actions.
--
-- The `denouncements` table only holds *current* state (rows are deleted on
-- forgive), so it cannot say which grudges were active when a past game was
-- played. Rating amplification (matchups between denounced rivals count 1.5x)
-- must be reproducible across full recalculations, so we reconstruct the
-- historical state from this log instead.

CREATE TABLE IF NOT EXISTS denouncement_events (
    id           SERIAL PRIMARY KEY,
    denouncer_id INTEGER NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    denounced_id INTEGER NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    action       TEXT NOT NULL CHECK (action IN ('denounce', 'forgive')),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS denouncement_events_pair_idx
    ON denouncement_events (denouncer_id, denounced_id, created_at DESC);

-- Backfill: seed a 'denounce' event for every currently-active denouncement
-- that has no history yet, so existing grudges keep amplifying.
INSERT INTO denouncement_events (denouncer_id, denounced_id, action, created_at)
SELECT d.denouncer_id, d.denounced_id, 'denounce', d.created_at
FROM denouncements d
WHERE NOT EXISTS (
    SELECT 1 FROM denouncement_events e
    WHERE e.denouncer_id = d.denouncer_id
      AND e.denounced_id = d.denounced_id
);
