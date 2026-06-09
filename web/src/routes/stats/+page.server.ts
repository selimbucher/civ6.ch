import postgres from 'postgres';
import type { PageServerLoad } from './$types';

const sql = postgres();

export const load: PageServerLoad = async () => {
    const [globalStats] = await sql`
        SELECT
            COUNT(DISTINCT id)::int AS total_games,
            COALESCE(SUM(turns) FILTER (WHERE turns < 1000), 0)::int AS total_turns,
            ROUND(AVG(turns) FILTER (WHERE turns < 1000))::int AS avg_turns
        FROM games WHERE tmp = false
    `;

    const [{ total_players }] = await sql`
        SELECT COUNT(DISTINCT player_id)::int AS total_players
        FROM game_players WHERE player_id IS NOT NULL
    `;

    const winsByType = await sql`
        SELECT p.name, g.victory_type, COUNT(*)::int AS wins
        FROM game_players gp
        JOIN players p ON p.id = gp.player_id
        JOIN games g ON g.id = gp.game_id
        WHERE g.tmp = false AND gp.winner = true AND gp.player_id IS NOT NULL
        GROUP BY p.name, g.victory_type
        ORDER BY wins DESC
    `;

    const [mostLosses] = await sql`
        SELECT p.name, COUNT(*)::int AS count
        FROM game_players gp JOIN players p ON p.id = gp.player_id JOIN games g ON g.id = gp.game_id
        WHERE g.tmp = false AND gp.winner = false AND gp.player_id IS NOT NULL
        GROUP BY p.name ORDER BY count DESC LIMIT 1
    `;

    const [mostGames] = await sql`
        SELECT p.name, COUNT(*)::int AS count
        FROM game_players gp JOIN players p ON p.id = gp.player_id JOIN games g ON g.id = gp.game_id
        WHERE g.tmp = false AND gp.player_id IS NOT NULL
        GROUP BY p.name ORDER BY count DESC LIMIT 1
    `;

    const [biggestGain] = await sql`
        SELECT p.name, ROUND(gp.post_rating_overall - gp.pre_rating_overall)::int AS delta
        FROM game_players gp JOIN players p ON p.id = gp.player_id JOIN games g ON g.id = gp.game_id
        WHERE g.tmp = false AND gp.player_id IS NOT NULL
        ORDER BY (gp.post_rating_overall - gp.pre_rating_overall) DESC LIMIT 1
    `;

    const [biggestLoss] = await sql`
        SELECT p.name, ROUND(gp.pre_rating_overall - gp.post_rating_overall)::int AS delta
        FROM game_players gp JOIN players p ON p.id = gp.player_id JOIN games g ON g.id = gp.game_id
        WHERE g.tmp = false AND gp.player_id IS NOT NULL
        ORDER BY (gp.pre_rating_overall - gp.post_rating_overall) DESC LIMIT 1
    `;

    const [mostTurns] = await sql`
        SELECT p.name, SUM(g.turns)::int AS count
        FROM game_players gp JOIN players p ON p.id = gp.player_id JOIN games g ON g.id = gp.game_id
        WHERE g.tmp = false AND gp.player_id IS NOT NULL AND g.turns < 1000
        GROUP BY p.name ORDER BY count DESC LIMIT 1
    `;

    const [shortestGame] = await sql`
        SELECT g.id, g.turns::int, p.name
        FROM games g
        JOIN game_players gp ON gp.game_id = g.id AND gp.winner = true
        JOIN players p ON p.id = gp.player_id
        WHERE g.tmp = false AND g.turns > 0 AND g.turns < 1000 AND g.victory_type <> 'Capitulation'
        ORDER BY g.turns ASC LIMIT 1
    `;

    const victoryDist = await sql`
        SELECT victory_type, COUNT(*)::int AS count
        FROM games WHERE tmp = false
        GROUP BY victory_type ORDER BY count DESC
    `;

    const topLeaders = await sql`
        SELECT leader, COUNT(*)::int AS picks, SUM(winner::int)::int AS wins
        FROM game_players gp JOIN games g ON g.id = gp.game_id
        WHERE g.tmp = false AND leader IS NOT NULL
        GROUP BY leader ORDER BY picks DESC LIMIT 12
    `;

    const byType: Record<string, { name: string; count: number }> = {};
    for (const row of winsByType) {
        const vt = row.victory_type as string;
        if (!byType[vt]) {
            byType[vt] = { name: row.name, count: row.wins };
        }
    }

    return {
        globalStats: {
            total_games: globalStats.total_games as number,
            total_turns: globalStats.total_turns as number,
            avg_turns: globalStats.avg_turns as number,
            total_players,
        },
        byType,
        mostLosses: mostLosses as { name: string; count: number } | undefined,
        mostGames: mostGames as { name: string; count: number } | undefined,
        biggestGain: biggestGain as { name: string; delta: number } | undefined,
        biggestLoss: biggestLoss as { name: string; delta: number } | undefined,
        mostTurns: mostTurns as { name: string; count: number } | undefined,
        shortestGame: shortestGame as { id: number; turns: number; name: string } | undefined,
        victoryDist,
        topLeaders,
    };
};
