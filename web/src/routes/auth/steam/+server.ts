import { redirect } from '@sveltejs/kit';
import type { RequestHandler } from './$types';

// Kick off Steam OpenID 2.0 sign-in. Steam verifies the user owns the account
// and redirects back to /auth/steam/return with the claimed identity.
export const GET: RequestHandler = ({ url, locals }) => {
    if (!locals.user) redirect(303, '/login');

    const params = new URLSearchParams({
        'openid.ns': 'http://specs.openid.net/auth/2.0',
        'openid.mode': 'checkid_setup',
        'openid.return_to': `${url.origin}/auth/steam/return`,
        'openid.realm': url.origin,
        'openid.identity': 'http://specs.openid.net/auth/2.0/identifier_select',
        'openid.claimed_id': 'http://specs.openid.net/auth/2.0/identifier_select'
    });

    redirect(303, `https://steamcommunity.com/openid/login?${params}`);
};
