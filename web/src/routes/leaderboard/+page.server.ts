import postgres from 'postgres';

const sql = postgres();

export async function load({ url }) {
    const category = url.searchParams.get('category') ?? 'ffa';

    const overall = await sql`
        SELECT
            p.id, p.name, p.avatar,
            COALESCE(pr.rating, 1500) as rating,
            COALESCE(pr.rd, 1500) as rd,
            COALESCE(ps.games, 0) as games,
            COALESCE(ps.wins, 0) as wins,
            COALESCE(ps.streak, 0) as streak,
            (
                SELECT gp2.leader
                FROM game_players gp2
                JOIN games g2 ON g2.id = gp2.game_id AND g2.tmp = false
                WHERE gp2.player_id = p.id AND gp2.leader IS NOT NULL
                GROUP BY gp2.leader
                ORDER BY COUNT(*) DESC
                LIMIT 1
            ) as top_leader
        FROM players p
        LEFT JOIN player_ratings pr ON pr.player_id = p.id AND pr.category = 'overall'
        LEFT JOIN player_stats ps ON ps.player_id = p.id AND ps.category = 'overall'
        WHERE p.active = true AND COALESCE(ps.games, 0) > 0
        ORDER BY pr.rating DESC NULLS LAST
    `;

    const categorical = await sql`
        SELECT
            p.id, p.name, p.avatar,
            COALESCE(pr.rating, 1500) as rating,
            COALESCE(pr.rd, 1500) as rd,
            COALESCE(ps.games, 0) as games,
            COALESCE(ps.wins, 0) as wins,
            COALESCE(ps.streak, 0) as streak
        FROM players p
        LEFT JOIN player_ratings pr ON pr.player_id = p.id AND pr.category = ${category}
        LEFT JOIN player_stats ps ON ps.player_id = p.id AND ps.category = ${category}
        WHERE p.active = true AND COALESCE(ps.games, 0) > 0
        ORDER BY pr.rating DESC NULLS LAST
    `;

    return { overall, categorical, category };
}
