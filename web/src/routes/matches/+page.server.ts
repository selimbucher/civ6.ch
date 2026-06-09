import postgres from 'postgres';

const sql = postgres();

export async function load({ locals }) {
    const games = await sql`
        SELECT
            g.id,
            g.victory_type,
            g.category,
            g.map,
            g.turns,
            g.date,
            g.draw,
            g.has_map,
            json_agg(
                json_build_object(
                    'player_id', gp.player_id,
                    'name', p.name,
                    'team', gp.team,
                    'winner', gp.winner,
                    'score', gp.score,
                    'leader', gp.leader,
                    'pre_rating', gp.pre_rating_overall,
                    'post_rating', gp.post_rating_overall
                ) ORDER BY gp.team, gp.winner DESC
            ) as players
        FROM games g
        JOIN game_players gp ON gp.game_id = g.id
        JOIN players p ON p.id = gp.player_id
        WHERE g.tmp = false
        GROUP BY g.id
        ORDER BY g.date DESC
    `;

    let unconfirmed: any[] = [];
    if (locals.user) {
        unconfirmed = await sql`
            SELECT
                g.id,
                g.date,
                g.map,
                g.turns,
                g.has_map,
                json_agg(
                    json_build_object(
                        'leader',      gp.leader,
                        'pseudo_name', gp.pseudo_name
                    ) ORDER BY gp.id
                ) AS players
            FROM games g
            JOIN game_players gp ON gp.game_id = g.id
            WHERE g.tmp = true
            GROUP BY g.id
            ORDER BY g.date DESC
        `;
    }

    return { games, unconfirmed };
}