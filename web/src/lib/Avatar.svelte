<script lang="ts">
    // Renders a player's avatar: an uploaded image, a chosen leader portrait,
    // or a letter fallback (first character of the name).
    let {
        id,
        name,
        avatar = null,
        wrapClass = '',
        letterClass = ''
    }: {
        id: number;
        name: string;
        avatar?: string | null;
        wrapClass?: string;
        letterClass?: string;
    } = $props();

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

    const isLeader = $derived(avatar?.startsWith('leader:'));
    const src = $derived(
        !avatar
            ? null
            : isLeader
              ? leaderPortrait(avatar!.slice(7))
              : avatar === 'upload'
                ? `/files/avatars/${id}`
                : null
    );

    // Leader icons have built-in borders and transparency. Removing bg and border
    // classes prevents double-borders and inner-fill artifacts.
    const finalWrapClass = $derived(
        isLeader
            ? wrapClass.replace(/\bborder(?:-[a-z0-9-]+)?\b/g, '').replace(/\bbg-[a-z0-9-]+\b/g, '') + ' scale-[1.05]'
            : wrapClass
    );
</script>

<div class="overflow-hidden flex items-center justify-center {finalWrapClass}">
    {#if src}
        <img {src} alt="" class="h-full w-full object-cover" />
    {:else}
        <span class={letterClass}>{name?.charAt(0).toUpperCase() ?? '?'}</span>
    {/if}
</div>
