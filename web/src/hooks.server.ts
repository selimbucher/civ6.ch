import postgres from 'postgres';
import type { Handle } from '@sveltejs/kit';

const sql = postgres();

export const handle: Handle = async ({ event, resolve }) => {
    const token = event.cookies.get('session');

    if (token) {
        const [session] = await sql`
            SELECT s.player_id, u.username, u.privileges, p.name
            FROM sessions s
            JOIN users u ON u.id = s.player_id
            LEFT JOIN players p ON p.id = s.player_id
            WHERE s.token = ${token}
            AND s.expires_at > now()
        `;
        if (session) {
            event.locals.user = {
                id: session.player_id,
                username: session.username,
                name: session.name,
                privileges: session.privileges,
            };
            // Re-issue the cookie as SameSite=Lax so sessions created with the
            // old 'strict' setting are upgraded transparently (needed for the
            // Steam OpenID return to carry the session).
            event.cookies.set('session', token, {
                httpOnly: true,
                secure: true,
                sameSite: 'lax',
                maxAge: 60 * 60 * 24 * 30,
                path: '/'
            });
        }
    }

    return resolve(event);
};