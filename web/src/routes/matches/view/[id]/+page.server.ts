import postgres from 'postgres';
import { error } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';

const sql = postgres();


export const load: PageServerLoad = async ({ params }) => {
    const id = parseInt(params.id);
    if (isNaN(id)) error(400, 'Invalid game ID');

    const [game] = await sql`
        SELECT
            g.id,
            g.victory_type,
            g.category,
            g.map,
            g.map_size,
            g.game_speed,
            g.turns,
            g.date,
            g.draw,
            g.has_map,
            g.has_save,
            g.shuffle_techs,
            g.allow_conquest,
            g.allow_score,
            g.allow_science,
            g.allow_religious,
            g.allow_culture,
            g.allow_diplomatic,
            g.secret_societies,
            g.heroes_and_legends,
            g.apocalypse_mode,
            g.monopolies,
            g.barbarian_clans,
            g.zombie_defense,
            g.era,
            g.ruleset,
            g.difficulty,
            g.save_filename,
            json_agg(
                json_build_object(
                    'player_id',           gp.player_id,
                    'name',                p.name,
                    'pseudo_name',         gp.pseudo_name,
                    'leader',              gp.leader,
                    'team',                gp.team,
                    'winner',              gp.winner,
                    'eliminated',          gp.eliminated,
                    'left_game',           gp.left_game,
                    'score',               gp.score,
                    'pre_rating',          gp.pre_rating,
                    'post_rating',         gp.post_rating,
                    'pre_rating_overall',  gp.pre_rating_overall,
                    'post_rating_overall', gp.post_rating_overall,
                    'population',          gp.population,
                    'science',             gp.science,
                    'culture',             gp.culture,
                    'food',                gp.food,
                    'production',          gp.production,
                    'gold',                gp.gold,
                    'faith',               gp.faith,
                    'tourism',             gp.tourism,
                    'favor',               gp.favor,
                    'cities',              (
                        SELECT COALESCE(json_agg(
                            json_build_object(
                                'name', c.name,
                                'population', c.population,
                                'religion', c.religion,
                                'wonders', c.wonders,
                                'food', c.food,
                                'production', c.production,
                                'gold', c.gold,
                                'science', c.science,
                                'culture', c.culture,
                                'faith', c.faith
                            ) ORDER BY c.population DESC, c.name ASC
                        ), '[]'::json)
                        FROM game_player_cities c
                        WHERE c.game_player_id = gp.id
                    )
                ) ORDER BY
                    gp.winner DESC,
                    gp.team_avg_score DESC,
                    gp.score DESC
            ) AS players
        FROM games g
        JOIN (
            SELECT
                gp.*,
                AVG(gp.score) OVER (PARTITION BY gp.team) AS team_avg_score
            FROM game_players gp
            WHERE gp.game_id = ${id}
        ) gp ON gp.game_id = g.id
        JOIN players p ON p.id = gp.player_id
        WHERE g.id = ${id}
        GROUP BY g.id
    `;

    if (!game) error(404, 'Game not found');

    return { game, hasMap: !!game.has_map, hasSave: !!game.has_save };
};