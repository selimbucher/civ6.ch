import { redirect } from '@sveltejs/kit';
import postgres from 'postgres';
import type { RequestHandler } from './$types';

const sql = postgres();

// Steam OpenID callback: verify the assertion with Steam, extract the
// SteamID64 from the claimed id, and link it to the logged-in player.
export const GET: RequestHandler = async ({ url, locals }) => {
    if (!locals.user) redirect(303, '/login');

    // Re-send all openid.* params back to Steam with mode=check_authentication
    // so Steam confirms the response was genuinely issued by it.
    const verify = new URLSearchParams(url.searchParams);
    verify.set('openid.mode', 'check_authentication');

    let valid = false;
    try {
        const res = await fetch('https://steamcommunity.com/openid/login', {
            method: 'POST',
            headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
            body: verify.toString()
        });
        valid = (await res.text()).includes('is_valid:true');
    } catch {
        redirect(303, '/settings?steam=error');
    }
    if (!valid) redirect(303, '/settings?steam=error');

    const claimedId = url.searchParams.get('openid.claimed_id') ?? '';
    const match = claimedId.match(/\/openid\/id\/(\d{17})$/);
    if (!match) redirect(303, '/settings?steam=error');
    const steamId = match[1];

    // One Steam account maps to one player. Steam has just proven ownership,
    // so the authenticated user always wins any prior (mistaken) link.
    await sql`
        INSERT INTO player_steam_ids (steam_id, player_id, persona)
        VALUES (
            ${steamId},
            ${locals.user.id},
            (SELECT pseudo_name FROM game_players
              WHERE steam_id = ${steamId} AND pseudo_name IS NOT NULL
              ORDER BY id DESC LIMIT 1)
        )
        ON CONFLICT (steam_id)
        DO UPDATE SET player_id = ${locals.user.id}, linked_at = now()
    `;

    redirect(303, '/settings?steam=linked');
};
