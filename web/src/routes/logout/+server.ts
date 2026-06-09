import { redirect } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import postgres from 'postgres';

const sql = postgres();

export const GET: RequestHandler = async ({ cookies }) => {
    const token = cookies.get('session');
    if (token) {
        await sql`DELETE FROM sessions WHERE token = ${token}`;
    }
    cookies.delete('session', { path: '/' });
    redirect(303, '/');
};