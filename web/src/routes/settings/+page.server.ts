import { redirect, fail } from '@sveltejs/kit';
import postgres from 'postgres';
import { compare as bcryptCompare } from '@node-rs/bcrypt';
import { verify, hash } from '@node-rs/argon2';
import type { Actions, PageServerLoad } from './$types';

const sql = postgres();

export const load: PageServerLoad = async ({ locals }) => {
    if (!locals.user) redirect(303, '/login');

    const [profile] = await sql`
        SELECT u.username, u.email, p.name
        FROM users u
        LEFT JOIN players p ON p.id = u.id
        WHERE u.id = ${locals.user.id}
    `;

    const steamAccounts = await sql`
        SELECT steam_id, persona, linked_at
        FROM player_steam_ids
        WHERE player_id = ${locals.user.id}
        ORDER BY linked_at DESC
    `;

    return { profile, steamAccounts };
};

export const actions: Actions = {
    // ── Profile (display name + email) ─────────────────────────────────────────
    profile: async ({ request, locals }) => {
        if (!locals.user) return fail(401, { error: 'Not logged in' });
        const data = await request.formData();

        const first = (data.get('first') as string ?? '').trim();
        const last = (data.get('last') as string ?? '').trim();
        const emailRaw = (data.get('email') as string ?? '').trim();
        const email = emailRaw === '' ? null : emailRaw;
        const password = data.get('password') as string;

        if (!first || !last)
            return fail(400, { profileError: 'First and last name are both required' });
        if (first.length > 30 || last.length > 30)
            return fail(400, { profileError: 'Names must be 30 characters or fewer' });
        if (email && !/^[^@\s]+@[^@\s]+\.[^@\s]+$/.test(email))
            return fail(400, { profileError: 'That email address looks invalid' });
        if (!password)
            return fail(400, { profileError: 'Enter your current password to save changes' });

        // Profile changes require confirming the current password.
        const [user] = await sql`SELECT pw_hash FROM users WHERE id = ${locals.user.id}`;
        if (!user) return fail(404, { profileError: 'Account not found' });
        const isBcrypt = user.pw_hash.startsWith('$2y$') || user.pw_hash.startsWith('$2b$');
        const valid = isBcrypt
            ? await bcryptCompare(password, user.pw_hash.replace(/^\$2y\$/, '$2b$'))
            : await verify(user.pw_hash, password);
        if (!valid) return fail(401, { profileError: 'Current password is incorrect' });

        if (email) {
            const [clash] = await sql`
                SELECT id FROM users WHERE email = ${email} AND id <> ${locals.user.id}
            `;
            if (clash) return fail(400, { profileError: 'That email is already in use' });
        }

        const name = `${first} ${last}`;
        await sql`UPDATE players SET name = ${name} WHERE id = ${locals.user.id}`;
        await sql`UPDATE users SET email = ${email} WHERE id = ${locals.user.id}`;
        return { profileOk: true };
    },

    // ── Change password ────────────────────────────────────────────────────────
    password: async ({ request, locals }) => {
        if (!locals.user) return fail(401, { error: 'Not logged in' });
        const data = await request.formData();

        const current = data.get('current') as string;
        const next = data.get('new') as string;
        const confirm = data.get('confirm') as string;

        if (!current || !next) return fail(400, { passwordError: 'All fields are required' });
        if (next.length < 8) return fail(400, { passwordError: 'New password must be at least 8 characters' });
        if (next !== confirm) return fail(400, { passwordError: 'New passwords do not match' });

        const [user] = await sql`SELECT pw_hash FROM users WHERE id = ${locals.user.id}`;
        if (!user) return fail(404, { passwordError: 'Account not found' });

        const isBcrypt = user.pw_hash.startsWith('$2y$') || user.pw_hash.startsWith('$2b$');
        const valid = isBcrypt
            ? await bcryptCompare(current, user.pw_hash.replace(/^\$2y\$/, '$2b$'))
            : await verify(user.pw_hash, current);
        if (!valid) return fail(401, { passwordError: 'Current password is incorrect' });

        const newHash = await hash(next);
        await sql`UPDATE users SET pw_hash = ${newHash}, pw_attempts = 0 WHERE id = ${locals.user.id}`;
        return { passwordOk: true };
    },

    // ── Sign out of all other devices ──────────────────────────────────────────
    signout_all: async ({ locals, cookies }) => {
        if (!locals.user) return fail(401, { error: 'Not logged in' });
        const current = cookies.get('session');
        await sql`
            DELETE FROM sessions
            WHERE player_id = ${locals.user.id} AND token <> ${current ?? ''}
        `;
        return { signedOut: true };
    },

    // ── Unlink a Steam account ─────────────────────────────────────────────────
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
