import postgres from 'postgres';
import { error } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';

const sql = postgres();

export const load: PageServerLoad = async ({ params }) => {
    const id = parseInt(params.id);
    if (isNaN(id)) error(400, 'Invalid player ID');

    const [player] = await sql`
        SELECT id, name, achievement_points, achievement_bitstring, streak, active
        FROM players
        WHERE id = ${id}
    `;
    if (!player) error(404, 'Player not found');

    const ratings = await sql`
        SELECT category, rating, rd, last_played
        FROM player_ratings
        WHERE player_id = ${id}
    `;

    const stats = await sql`
        SELECT category, games, wins, streak, highest_winstreak
        FROM player_stats
        WHERE player_id = ${id}
    `;

    const ranks = await sql`
        SELECT category, rank::int
        FROM (
            SELECT
                pr.player_id,
                pr.category,
                ROW_NUMBER() OVER (PARTITION BY pr.category ORDER BY pr.rating DESC)::int AS rank
            FROM player_ratings pr
            JOIN players p ON p.id = pr.player_id
            WHERE p.active = true
        ) ranked
        WHERE player_id = ${id}
    `;

    const leaderStats = await sql`
        SELECT
            gp.leader,
            COUNT(*)::int AS games,
            SUM(gp.winner::int)::int AS wins
        FROM game_players gp
        JOIN games g ON g.id = gp.game_id AND g.tmp = false
        WHERE gp.player_id = ${id}
          AND gp.leader IS NOT NULL
        GROUP BY gp.leader
        ORDER BY games DESC
        LIMIT 8
    `;

    const victoryStats = await sql`
        SELECT
            g.victory_type,
            COUNT(*)::int AS total,
            SUM(gp.winner::int)::int AS wins
        FROM game_players gp
        JOIN games g ON g.id = gp.game_id AND g.tmp = false
        WHERE gp.player_id = ${id}
        GROUP BY g.victory_type
        ORDER BY total DESC
    `;

    const ratingHistory = await sql`
        SELECT
            g.date,
            gp.pre_rating_overall,
            gp.post_rating_overall,
            gp.winner
        FROM game_players gp
        JOIN games g ON g.id = gp.game_id AND g.tmp = false
        WHERE gp.player_id = ${id}
          AND gp.pre_rating_overall > 0
        ORDER BY g.date ASC
    `;

    const recentGames = await sql`
        SELECT
            g.id,
            g.victory_type,
            g.category,
            g.map,
            g.turns,
            g.date,
            gp.winner,
            gp.leader,
            gp.pre_rating_overall,
            gp.post_rating_overall,
            (
                SELECT json_agg(
                    json_build_object(
                        'player_id', gp2.player_id,
                        'name', p2.name,
                        'winner', gp2.winner
                    ) ORDER BY gp2.winner DESC, p2.name
                )
                FROM game_players gp2
                JOIN players p2 ON p2.id = gp2.player_id
                WHERE gp2.game_id = g.id
            ) AS players
        FROM game_players gp
        JOIN games g ON g.id = gp.game_id AND g.tmp = false
        WHERE gp.player_id = ${id}
        ORDER BY g.date DESC
        LIMIT 15
    `;

    const achievements = await sql`
        SELECT id, name, description, difficulty
        FROM achievements
        WHERE disabled = false
        ORDER BY difficulty DESC, id
    `;

    const highestRating = await sql`
        SELECT MAX(post_rating_overall)::int as max_rating
        FROM game_players
        WHERE player_id = ${id}
    `;

    const personalStats = await sql`
        SELECT
            COALESCE(ps.wins, 0) as wins,
            COALESCE(ps.games, 0) as games,
            (SELECT COUNT(*) FROM game_players gp JOIN games g ON g.id=gp.game_id WHERE gp.player_id=${id} AND gp.winner=true AND g.victory_type='Diplomatic') as diplomatic_wins,
            (SELECT COUNT(*) FROM game_player_cities c JOIN game_players gp ON gp.id=c.game_player_id WHERE gp.player_id=${id})::int as cities_founded,
            (SELECT MAX(g.date) FROM game_players gp JOIN games g ON g.id=gp.game_id WHERE gp.player_id=${id})::date as last_game_date,
            (SELECT MIN(g.date) FROM game_players gp JOIN games g ON g.id=gp.game_id WHERE gp.player_id=${id})::date as first_game_date,
            (SELECT COUNT(DISTINCT g.id) FROM games g JOIN game_players gp_self ON gp_self.game_id=g.id AND gp_self.player_id=${id} JOIN game_players gp_other ON gp_other.game_id=g.id AND gp_other.player_id != ${id} JOIN game_player_cities c ON c.game_player_id=gp_other.id WHERE c.wonders IS NOT NULL AND 'Colosseum' = ANY(c.wonders))::int as colosseum_robbed
        FROM player_stats ps
        WHERE ps.player_id = ${id} AND ps.category = 'overall'
    `;

    return {
        player,
        ratings,
        stats,
        ranks,
        leaderStats,
        victoryStats,
        ratingHistory,
        recentGames,
        achievements,
        highestRating: highestRating[0],
        personalStats: personalStats[0]
    };
};
