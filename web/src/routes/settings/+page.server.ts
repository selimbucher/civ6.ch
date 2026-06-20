import { redirect, fail } from '@sveltejs/kit';
import postgres from 'postgres';
import { compare as bcryptCompare } from '@node-rs/bcrypt';
import { verify, hash } from '@node-rs/argon2';
import type { Actions, PageServerLoad } from './$types';

const sql = postgres();
const GO_SERVER = process.env.GO_SERVER_URL ?? 'http://localhost:8080';

// Notification preference keys persisted in users.settings.notify.
const NOTIFY_KEYS = ['new_game', 'denounced', 'weekly', 'achievement'] as const;

export const load: PageServerLoad = async ({ locals }) => {
    if (!locals.user) redirect(303, '/login');
    const me = locals.user.id;

    const [profile] = await sql`
        SELECT u.username, u.email, u.settings, p.name, p.active, p.avatar
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
    // ── Profile (email), confirmed by password ─────────────────────────────────
    profile: async ({ request, locals }) => {
        if (!locals.user) return fail(401, { error: 'Not logged in' });
        const data = await request.formData();

        const emailRaw = (data.get('email') as string ?? '').trim();
        const email = emailRaw === '' ? null : emailRaw;
        const password = data.get('password') as string;

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

    // ── Profile picture ────────────────────────────────────────────────────────
    avatar_upload: async ({ request, locals }) => {
        if (!locals.user) return fail(401, { error: 'Not logged in' });
        const data = await request.formData();
        const file = data.get('avatar');
        if (!(file instanceof File) || file.size === 0)
            return fail(400, { avatarError: 'Choose an image to upload' });
        if (!file.type.startsWith('image/'))
            return fail(400, { avatarError: 'That file is not an image' });
        if (file.size > 5 * 1024 * 1024)
            return fail(400, { avatarError: 'Image must be under 5 MB' });

        let res: Response;
        try {
            res = await fetch(`${GO_SERVER}/players/${locals.user.id}/avatar`, {
                method: 'POST',
                headers: { 'Content-Type': file.type },
                body: await file.arrayBuffer()
            });
        } catch {
            return fail(502, { avatarError: 'Upload service unavailable' });
        }
        if (!res.ok) return fail(500, { avatarError: 'Upload failed' });

        await sql`UPDATE players SET avatar = 'upload' WHERE id = ${locals.user.id}`;
        return { avatarOk: true };
    },
    avatar_leader: async ({ request, locals }) => {
        if (!locals.user) return fail(401, { error: 'Not logged in' });
        const data = await request.formData();
        const slug = (data.get('leader') as string ?? '').trim();
        if (!slug) return fail(400, { avatarError: 'Choose a leader' });
        await sql`UPDATE players SET avatar = ${'leader:' + slug} WHERE id = ${locals.user.id}`;
        return { avatarOk: true };
    },
    avatar_clear: async ({ locals }) => {
        if (!locals.user) return fail(401, { error: 'Not logged in' });
        await sql`UPDATE players SET avatar = NULL WHERE id = ${locals.user.id}`;
        return { avatarOk: true };
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
        // denouncements holds the current state (for badges); denouncement_events
        // is an append-only log so rating amplification stays reproducible.
        const inserted = await sql`
            INSERT INTO denouncements (denouncer_id, denounced_id)
            VALUES (${locals.user.id}, ${target})
            ON CONFLICT DO NOTHING
            RETURNING denouncer_id
        `;
        if (inserted.length) {
            await sql`
                INSERT INTO denouncement_events (denouncer_id, denounced_id, action)
                VALUES (${locals.user.id}, ${target}, 'denounce')
            `;
        }
        return { denounceOk: true };
    },
    // Only the denouncer can forgive their own denouncement (scoped by
    // denouncer_id) — otherwise the target could simply clear it themselves.
    forgive: async ({ request, locals }) => {
        if (!locals.user) return fail(401, { error: 'Not logged in' });
        const data = await request.formData();
        const target = parseInt(data.get('player_id') as string);
        if (isNaN(target)) return fail(400, { diploError: 'Missing player' });
        const removed = await sql`
            DELETE FROM denouncements
            WHERE denouncer_id = ${locals.user.id} AND denounced_id = ${target}
            RETURNING denounced_id
        `;
        if (removed.length) {
            await sql`
                INSERT INTO denouncement_events (denouncer_id, denounced_id, action)
                VALUES (${locals.user.id}, ${target}, 'forgive')
            `;
        }
        return { forgiveOk: true };
    },

    // ── Danger zone: real, reversible actions ──────────────────────────────────
    // Denounce every other active player at once.
    denounce_all: async ({ locals }) => {
        if (!locals.user) return fail(401, { error: 'Not logged in' });
        const me = locals.user.id;
        const inserted = await sql`
            INSERT INTO denouncements (denouncer_id, denounced_id)
            SELECT ${me}, p.id FROM players p
            WHERE p.active = true AND p.id <> ${me}
            ON CONFLICT DO NOTHING
            RETURNING denounced_id
        `;
        for (const r of inserted) {
            await sql`
                INSERT INTO denouncement_events (denouncer_id, denounced_id, action)
                VALUES (${me}, ${r.denounced_id}, 'denounce')
            `;
        }
        return { denounceAllOk: true, denounceAllCount: inserted.length };
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
