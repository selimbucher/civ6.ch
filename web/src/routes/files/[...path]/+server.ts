import type { RequestHandler } from './$types';

const GO_SERVER = process.env.GO_SERVER_URL ?? 'http://localhost:8080';

export const GET: RequestHandler = async ({ params }) => {
    let res: Response;
    try {
        res = await fetch(`${GO_SERVER}/files/${params.path}`);
    } catch {
        return new Response('upstream unavailable', { status: 503 });
    }

    if (!res.ok) {
        return new Response(null, { status: res.status });
    }

    // Read the full body as bytes — avoids issues with stream passthrough.
    const body = await res.arrayBuffer();

    return new Response(body, {
        status: 200,
        headers: {
            'Content-Type':        res.headers.get('Content-Type')        ?? 'application/octet-stream',
            'Content-Disposition': res.headers.get('Content-Disposition') ?? '',
            'Cache-Control':       res.headers.get('Cache-Control')       ?? 'no-cache',
        },
    });
};
