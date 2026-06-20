// Resolves a player's stored avatar value (players.avatar) to an image URL,
// mirroring the logic in Avatar.svelte. Returns null when there's no picture
// (callers fall back to a letter/initial). Avatars are stored as:
//   'leader:<slug>'   — a built-in leader portrait
//   'upload' | 'upload:<version>' — an uploaded image (version busts the cache)
//   null              — no avatar
const leaderAssets = import.meta.glob<{ default: string }>(
    '$lib/assets/icons/leaders/*.webp',
    { eager: true }
);

function leaderPortrait(slug: string): string | null {
    const norm = (s: string) => s.normalize('NFD').replace(/[̀-ͯ]/g, '').toLowerCase();
    const keys = Object.keys(leaderAssets);
    const key =
        keys.find((k) => k.includes(`/${slug}_(Civ6).`)) ??
        keys.find((k) => norm(k).includes(`/${norm(slug)}_(civ6).`));
    return key ? leaderAssets[key].default : null;
}

export function avatarUrl(id: number, avatar: string | null | undefined): string | null {
    if (!avatar) return null;
    if (avatar.startsWith('leader:')) return leaderPortrait(avatar.slice(7));
    if (avatar === 'upload' || avatar.startsWith('upload:')) {
        const v = avatar.startsWith('upload:') ? avatar.slice(7) : '';
        return `/files/avatars/${id}${v ? `?v=${v}` : ''}`;
    }
    return null;
}
