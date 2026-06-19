<script lang="ts">
    // Renders a player's avatar: an uploaded image, a chosen leader portrait,
    // or a letter fallback (first character of the name). Optionally overlays a
    // "denounced" badge in the bottom-left corner.
    import denouncedIcon from '$lib/assets/icons/relationships/denounced.png';

    let {
        id,
        name,
        avatar = null,
        denounced = false,
        wrapClass = '',
        letterClass = ''
    }: {
        id: number;
        name: string;
        avatar?: string | null;
        denounced?: boolean;
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

<!-- relative (not overflow-hidden) so the denounced badge can sit in the corner;
     the image clips itself to the circle via rounded-[inherit]. -->
<div class="relative flex items-center justify-center {finalWrapClass}">
    {#if src}
        <img {src} alt="" class="h-full w-full rounded-[inherit] object-cover" />
    {:else}
        <span class="leading-none {letterClass}">{name?.charAt(0).toUpperCase() ?? '?'}</span>
    {/if}
    {#if denounced}
        <img src={denouncedIcon} alt="" title="Denounced"
            class="absolute bottom-0 left-0 w-[45%] h-[45%] object-contain pointer-events-none drop-shadow-[0_1px_2px_rgba(0,0,0,0.55)]" />
    {/if}
</div>
