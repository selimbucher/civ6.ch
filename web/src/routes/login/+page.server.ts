import { fail, redirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ locals }) => {
    if (locals.user) redirect(303, '/');
};
import postgres from 'postgres';
import { compare as bcryptCompare } from '@node-rs/bcrypt';
import { verify, hash } from '@node-rs/argon2';

const sql = postgres();

export const actions: Actions = {
    default: async ({ request, cookies }) => {
        const data = await request.formData();
        const username = data.get('username') as string;
        const password = data.get('password') as string;

        if (!username || !password) {
            return fail(400, { error: 'Username and password are required' });
        }

        const [user] = await sql`
            SELECT id, pw_hash, pw_attempts FROM users
            WHERE username = ${username} OR email = ${username}
        `;

        const isBcrypt = user.pw_hash.startsWith('$2y$') || user.pw_hash.startsWith('$2b$');

        // Node's bcrypt doesn't recognize the php shit $2y$, replace with $2b$
        const hashForCompare = user.pw_hash.replace(/^\$2y\$/, '$2b$');

        const valid = isBcrypt
            ? await bcryptCompare(password, hashForCompare)
            : await verify(user.pw_hash, password);

        if (!valid) {
            await sql`
                UPDATE users SET pw_attempts = pw_attempts + 1 WHERE id = ${user.id}
            `;
            return fail(401, { error: 'Invalid username or password' });
        }

        // Migrate bcrypt → argon2 on successful login
        if (isBcrypt) {
            const newHash = await hash(password);
            await sql`
                UPDATE users SET pw_hash = ${newHash}, pw_attempts = 0 WHERE id = ${user.id}
            `;
        } else {
            await sql`UPDATE users SET pw_attempts = 0 WHERE id = ${user.id}`;
        }

        // Create session
        const [session] = await sql`
            INSERT INTO sessions (player_id, token, expires_at)
            VALUES (${user.id}, gen_random_uuid(), now() + interval '30 days')
            RETURNING token
        `;

        cookies.set('session', session.token, {
            httpOnly: true,
            secure: true,
            // 'lax' (not 'strict') so the session cookie is still sent on the
            // top-level GET redirect back from external sign-ins (Steam OpenID).
            sameSite: 'lax',
            maxAge: 60 * 60 * 24 * 30,
            path: '/'
        });

        redirect(303, '/');
    }
};