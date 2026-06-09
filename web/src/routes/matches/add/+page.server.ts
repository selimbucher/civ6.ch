import { fail, redirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ locals }) => {
    if (!locals.user) redirect(303, '/login');
};

export const actions: Actions = {
    default: async ({ request, locals }) => {
        if (!locals.user) return fail(401, { error: 'Not logged in' });

        const data = await request.formData();
        const entry = data.get('save');
        if (!(entry instanceof File) || entry.size === 0) {
            return fail(400, { error: 'No file selected' });
        }
        const file = entry;

        const name = file.name.toLowerCase();
        if (!name.endsWith('.civ6save')) {
            return fail(400, { error: `Invalid file extension: ${file.name}` });
        }

        let res: Response;
        try {
            // Changed from 'localhost' to '127.0.0.1' to avoid loopback resolution overhead
            res = await fetch('http://127.0.0.1:8080/parse', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/octet-stream',
                    'X-Filename': file.name,
                },
                body: await file.arrayBuffer(),
            });
        } catch (err) {
            return fail(502, { error: 'Parser service unavailable' });
        }

        if (!res.ok) {
            const errText = await res.text();
            return fail(500, { error: `Failed to parse save file: ${errText}` });
        }

        const { game_id } = await res.json();
        redirect(303, `/matches/confirm/${game_id}`);
    }
};