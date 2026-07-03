import postgres from 'postgres';
import type { LayoutServerLoad } from './$types';

const sql = postgres();

// Footer counters. Refreshed lazily at most once every 5 minutes so the
// layout load doesn't hit the database on every navigation.
let league: { games: number; turns: number; denouncements: number } | null = null;
let leagueFetchedAt = 0;

async function leagueCounters() {
    if (league && Date.now() - leagueFetchedAt < 5 * 60_000) return league;
    const [row] = await sql`
        SELECT
            COUNT(*)::int AS games,
            COALESCE(SUM(turns) FILTER (WHERE turns < 1000), 0)::int AS turns,
            (SELECT COUNT(*)::int FROM denouncements) AS denouncements
        FROM games WHERE tmp = false
    `;
    league = { games: row.games, turns: row.turns, denouncements: row.denouncements };
    leagueFetchedAt = Date.now();
    return league;
}

export const load: LayoutServerLoad = async ({ locals }) => {
    return { user: locals.user ?? null, league: await leagueCounters() };
};
