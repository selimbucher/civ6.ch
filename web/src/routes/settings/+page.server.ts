import { redirect, fail } from '@sveltejs/kit';
import postgres from 'postgres';
import type { Actions, PageServerLoad } from './$types';

const sql = postgres();

export const load: PageServerLoad = async ({ locals }) => {
    if (!locals.user) redirect(303, '/login');

    const steamAccounts = await sql`
        SELECT steam_id, persona, linked_at
        FROM player_steam_ids
        WHERE player_id = ${locals.user.id}
        ORDER BY linked_at DESC
    `;

    return { steamAccounts };
};

export const actions: Actions = {
    unlink: async ({ request, locals }) => {
        if (!locals.user) return fail(401, { error: 'Not logged in' });
        const data = await request.formData();
        const steamId = data.get('steam_id') as string;
        if (!steamId) return fail(400, { error: 'Missing account' });

        await sql`
            DELETE FROM player_steam_ids
            WHERE steam_id = ${steamId} AND player_id = ${locals.user.id}
        `;
        return { unlinked: true };
    }
};
