import { redirect, fail } from '@sveltejs/kit';
import postgres from 'postgres';
import { compare as bcryptCompare } from '@node-rs/bcrypt';
import { verify, hash } from '@node-rs/argon2';
import type { Actions, PageServerLoad } from './$types';

const sql = postgres();

// Notification preference keys persisted in users.settings.notify.
const NOTIFY_KEYS = ['new_game', 'denounced', 'weekly', 'achievement'] as const;

export const load: PageServerLoad = async ({ locals }) => {
    if (!locals.user) redirect(303, '/login');
    const me = locals.user.id;

    const [profile] = await sql`
        SELECT u.username, u.email, u.settings, p.name, p.active
        FROM users u
        LEFT JOIN players p ON p.id = u.id
        WHERE u.id = ${me}
    `;

    const steamAccounts = await sql`
        SELECT steam_id, persona, linked_at
        FROM player_steam_ids
        WHERE player_id = ${me}
        ORDER BY linked_at DESC
    `;

    const denounced = await sql`
        SELECT d.denounced_id AS id, p.name, d.created_at
        FROM denouncements d
        JOIN players p ON p.id = d.denounced_id
        WHERE d.denouncer_id = ${me}
        ORDER BY p.name
    `;

    const players = await sql`
        SELECT id, name FROM players
        WHERE active = true AND id <> ${me}
        ORDER BY name
    `;

    return { profile, steamAccounts, denounced, players, notify: profile?.settings?.notify ?? {} };
};

export const actions: Actions = {
    // ── Profile (display name + email), confirmed by password ──────────────────
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

    // ── Notification preferences (the Town Crier) ──────────────────────────────
    notifications: async ({ request, locals }) => {
        if (!locals.user) return fail(401, { error: 'Not logged in' });
        const data = await request.formData();
        const notify: Record<string, boolean> = {};
        for (const k of NOTIFY_KEYS) notify[k] = data.get(`notify_${k}`) === 'on';
        await sql`
            UPDATE users
            SET settings = jsonb_set(COALESCE(settings, '{}'::jsonb), '{notify}', ${JSON.stringify(notify)}::jsonb)
            WHERE id = ${locals.user.id}
        `;
        return { notifyOk: true };
    },

    // ── Diplomacy: denounce / forgive ──────────────────────────────────────────
    denounce: async ({ request, locals }) => {
        if (!locals.user) return fail(401, { error: 'Not logged in' });
        const data = await request.formData();
        const target = parseInt(data.get('player_id') as string);
        if (isNaN(target)) return fail(400, { diploError: 'Choose a player to denounce' });
        if (target === locals.user.id) return fail(400, { diploError: 'You cannot denounce yourself. Probably.' });
        await sql`
            INSERT INTO denouncements (denouncer_id, denounced_id)
            VALUES (${locals.user.id}, ${target})
            ON CONFLICT DO NOTHING
        `;
        return { denounceOk: true };
    },
    forgive: async ({ request, locals }) => {
        if (!locals.user) return fail(401, { error: 'Not logged in' });
        const data = await request.formData();
        const target = parseInt(data.get('player_id') as string);
        if (isNaN(target)) return fail(400, { diploError: 'Missing player' });
        await sql`
            DELETE FROM denouncements
            WHERE denouncer_id = ${locals.user.id} AND denounced_id = ${target}
        `;
        return { forgiveOk: true };
    },

    // ── Danger zone: real, reversible actions ──────────────────────────────────
    // Withdraw every denouncement you have issued.
    sue_for_peace: async ({ locals }) => {
        if (!locals.user) return fail(401, { error: 'Not logged in' });
        const res = await sql`DELETE FROM denouncements WHERE denouncer_id = ${locals.user.id}`;
        return { peaceOk: true, peaceCount: res.count };
    },
    // Unlink every Steam account.
    sever_steam: async ({ locals }) => {
        if (!locals.user) return fail(401, { error: 'Not logged in' });
        const res = await sql`DELETE FROM player_steam_ids WHERE player_id = ${locals.user.id}`;
        return { steamSevered: true, steamCount: res.count };
    },
    // Retire from / return to the competitive ladder (leaderboards + rankings).
    toggle_active: async ({ locals }) => {
        if (!locals.user) return fail(401, { error: 'Not logged in' });
        const [row] = await sql`
            UPDATE players SET active = NOT active WHERE id = ${locals.user.id} RETURNING active
        `;
        return { activeToggled: true, active: row?.active };
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
