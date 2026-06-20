import { error, fail, redirect } from '@sveltejs/kit';
import postgres from 'postgres';
import type { Actions, PageServerLoad } from './$types';

const sql = postgres();
const GO_SERVER = process.env.GO_SERVER_URL ?? 'http://localhost:8080';

export const load: PageServerLoad = async ({ params, locals }) => {
    if (!locals.user) redirect(303, '/login');

    const id = parseInt(params.id ?? '');
    if (isNaN(id)) error(400, 'Invalid game ID');

    const [game] = await sql`
        SELECT id, turns, map, map_size, game_speed, victory_type, category, tmp, has_map
        FROM games WHERE id = ${id}
    `;
    if (!game) error(404, 'Game not found');
    // finalized games are immutable
    if (!game.tmp) redirect(303, `/matches/view/${id}`);

    const rows = await sql`
        SELECT gp.id, gp.team, gp.eliminated, gp.leader, gp.pseudo_name, gp.score,
               gp.population, gp.science, gp.culture, gp.food, gp.production,
               gp.gold, gp.faith, gp.tourism, gp.favor,
               psi.player_id AS matched_player_id,
               mp.name       AS matched_player_name
        FROM game_players gp
        LEFT JOIN player_steam_ids psi ON psi.steam_id = gp.steam_id
        LEFT JOIN players mp ON mp.id = psi.player_id
        WHERE gp.game_id = ${id}
        ORDER BY gp.team, gp.id
    `;

    // Candidates for manual assignment: players who haven't linked a Steam
    // account. Anyone with a linked Steam ID is matched automatically from the
    // save, so listing them here would only bloat the picker.
    const players = await sql`
        SELECT id, name FROM players
        WHERE id NOT IN (SELECT player_id FROM player_steam_ids)
        ORDER BY name
    `;

    return { game, rows, players };
};

// ── Shared: build + validate player assignments ────────────────────────────────
async function parseAssignments(data: FormData, id: number) {
    const assignments: { rowId: number; playerId: number; winner: boolean }[] = [];
    const seenPlayers = new Set<number>();

    for (const [key, value] of data.entries()) {
        if (!key.startsWith('row_')) continue;
        const rowId = parseInt(key.slice(4));
        const playerId = parseInt(value as string);
        if (isNaN(rowId)) continue;
        if (isNaN(playerId) || playerId === 0)
            return { error: 'All players must be assigned', assignments: null };
        if (seenPlayers.has(playerId))
            return { error: 'The same player cannot be assigned twice', assignments: null };
        seenPlayers.add(playerId);
        assignments.push({ rowId, playerId, winner: data.get(`winner_${rowId}`) === 'on' });
    }

    if (assignments.length === 0)
        return { error: 'No players found', assignments: null };

    return { error: null, assignments };
}

export const actions: Actions = {

    // ── Delete (cancel) ────────────────────────────────────────────────────────
    cancel: async ({ params, locals }) => {
        if (!locals.user) return fail(401, { error: 'Not logged in' });
        const id = parseInt(params.id ?? '');
        if (isNaN(id)) return fail(400, { error: 'Invalid game ID' });

        const [game] = await sql`SELECT id, tmp FROM games WHERE id = ${id}`;
        if (!game) return fail(404, { error: 'Game not found' });
        if (!game.tmp) return fail(400, { error: 'Cannot delete a finalized game' });

        try { await fetch(`${GO_SERVER}/games/${id}`, { method: 'DELETE' }); } catch { /* ok */ }

        await sql`DELETE FROM game_player_cities WHERE game_player_id IN (SELECT id FROM game_players WHERE game_id = ${id})`;
        await sql`DELETE FROM game_players WHERE game_id = ${id}`;
        await sql`DELETE FROM games WHERE id = ${id}`;

        redirect(303, '/matches');
    },

    // ── Save progress (ongoing) ────────────────────────────────────────────────
    // Save progress — UI now uses <a href="/matches"> instead.
    save: async () => { redirect(303, '/matches'); },


    // ── Finalize ───────────────────────────────────────────────────────────────
    confirm: async ({ request, params, locals }) => {
        if (!locals.user) return fail(401, { error: 'Not logged in' });
        const id = parseInt(params.id ?? '');
        if (isNaN(id)) return fail(400, { error: 'Invalid game ID' });

        const data = await request.formData();
        const victoryType = data.get('victory_type') as string;
        if (!victoryType) return fail(400, { error: 'Victory type is required to finalize' });

        const { error: assignError, assignments } = await parseAssignments(data, id);
        if (assignError || !assignments) return fail(400, { error: assignError });

        const [gameCheck] = await sql`SELECT id FROM games WHERE id = ${id}`;
        if (!gameCheck) return fail(404, { error: 'Game not found' });

        // Team structure comes from the parsed save (shared-victory team id).
        const teamRows = await sql`SELECT id, team FROM game_players WHERE game_id = ${id}`;
        const teamByRow = new Map<number, number>(teamRows.map((r: any) => [r.id, r.team]));

        // Winners are decided per team: the winning team is whichever team an
        // admin marked, and the whole team wins together.
        const winningTeams = new Set(
            assignments.filter(a => a.winner).map(a => teamByRow.get(a.rowId))
        );
        if (winningTeams.size === 0) return fail(400, { error: 'Select the winning team' });
        if (winningTeams.size > 1) return fail(400, { error: 'Only one team can win' });
        const winningTeam = [...winningTeams][0];

        // Category from team sizes: any team with >1 member ⇒ teams game.
        const teamSizes = new Map<number, number>();
        for (const a of assignments) {
            const t = teamByRow.get(a.rowId)!;
            teamSizes.set(t, (teamSizes.get(t) ?? 0) + 1);
        }
        const maxTeam = Math.max(...teamSizes.values());
        const gameCategory = maxTeam > 1 ? 'teams' : assignments.length === 2 ? '1v1' : 'ffa';
        await sql`UPDATE games SET category = ${gameCategory} WHERE id = ${id}`;

        const playerIds = assignments.map(a => a.playerId);
        const ratings = await sql`
            SELECT p.id,
                COALESCE(pr_cat.rating,    1500) AS rating,
                COALESCE(pr_cat.rd,         350) AS rd,
                COALESCE(pr_cat.volatility, 0.06) AS volatility,
                COALESCE(pr_all.rating,    1500) AS rating_overall,
                COALESCE(pr_all.rd,         350) AS rd_overall,
                COALESCE(pr_all.volatility, 0.06) AS volatility_overall
            FROM players p
            LEFT JOIN player_ratings pr_cat ON pr_cat.player_id = p.id AND pr_cat.category = ${gameCategory}
            LEFT JOIN player_ratings pr_all ON pr_all.player_id = p.id AND pr_all.category = 'overall'
            WHERE p.id = ANY(${playerIds})
        `;
        const ratingMap = Object.fromEntries(ratings.map((r: any) => [r.id, r]));

        await sql.begin(async sql => {
            for (const { rowId, playerId } of assignments) {
                const r = ratingMap[playerId];
                const winner = teamByRow.get(rowId) === winningTeam;
                await sql`
                    UPDATE game_players SET
                        player_id               = ${playerId},
                        winner                  = ${winner},
                        pre_rating              = ${r.rating},
                        pre_rd                  = ${r.rd},
                        pre_volatility          = ${r.volatility},
                        post_rating             = ${r.rating},
                        post_rd                 = ${r.rd},
                        post_volatility         = ${r.volatility},
                        pre_rating_overall      = ${r.rating_overall},
                        pre_rd_overall          = ${r.rd_overall},
                        pre_volatility_overall  = ${r.volatility_overall},
                        post_rating_overall     = ${r.rating_overall},
                        post_rd_overall         = ${r.rd_overall},
                        post_volatility_overall = ${r.volatility_overall}
                    WHERE id = ${rowId} AND game_id = ${id}
                `;
            }
            await sql`
                UPDATE games SET
                    victory_type = ${victoryType},
                    tmp          = false
                WHERE id = ${id}
            `;
        });

        try {
            await fetch(`${GO_SERVER}/recalculate`, { method: 'POST' });
        } catch { /* non-fatal */ }

        redirect(303, `/matches/view/${id}`);
    },

    // ── Update save (ongoing games only) ──────────────────────────────────────
    update: async ({ request, params, locals }) => {
        if (!locals.user) return fail(401, { error: 'Not logged in' });
        const id = parseInt(params.id ?? '');
        if (isNaN(id)) return fail(400, { error: 'Invalid game ID' });

        const data = await request.formData();
        const entry = data.get('save');
        if (!(entry instanceof File) || entry.size === 0)
            return fail(400, { updateError: 'No file selected' });

        let res: Response;
        try {
            res = await fetch(`${GO_SERVER}/games/${id}/update`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/octet-stream' },
                body: await entry.arrayBuffer(),
            });
        } catch {
            return fail(502, { updateError: 'Parser service unavailable' });
        }

        const json: { ok: boolean; turns?: number; message?: string; error?: string } = await res.json();
        if (!json.ok) return fail(400, { updateError: json.error ?? 'Update failed' });

        return { updated: true, updateMessage: json.message };
    },
};
