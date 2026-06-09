import postgres from 'postgres';
import type { PageServerLoad } from './$types';

const sql = postgres();

export const load: PageServerLoad = async () => {
    const achievements = await sql`
        SELECT
            a.id,
            a.name,
            a.description,
            a.difficulty,
            a.exclusive,
            a.disabled,
            COALESCE(
                json_agg(
                    json_build_object(
                        'player_id',   p.id,
                        'player_name', p.name,
                        'game_id',     pa.game_id,
                        'earned_at',   pa.earned_at
                    ) ORDER BY pa.earned_at ASC
                ) FILTER (WHERE p.id IS NOT NULL),
                '[]'::json
            ) AS earners
        FROM achievements a
        LEFT JOIN player_achievements pa ON pa.achievement_id = a.id
        LEFT JOIN players p ON p.id = pa.player_id AND p.active = true
        GROUP BY a.id
        ORDER BY a.disabled ASC, a.difficulty DESC, a.id
    `;

    const players = await sql`
        SELECT id, name, achievement_points, achievement_bitstring
        FROM players
        WHERE active = true
        ORDER BY achievement_points DESC, name
    `;

    return { achievements, players };
};
