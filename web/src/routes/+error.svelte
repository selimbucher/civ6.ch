<script lang="ts">
    import { page } from '$app/state';
    import { Home, Swords } from '@lucide/svelte';

    const status = $derived(page.status);

    // A little flavour keyed to the status code, in the league's voice.
    const flavour = $derived.by(() => {
        if (status === 404)
            return {
                title: 'Lost to the fog of war',
                body: "This page isn't on the map. The scouts found nothing but barbarian camps out here."
            };
        if (status === 403)
            return {
                title: 'The gates are barred',
                body: "You don't have the diplomatic standing to enter this hall."
            };
        if (status >= 500)
            return {
                title: 'The empire has collapsed',
                body: 'Something broke on our side. The advisors have been notified and are panicking accordingly.'
            };
        return {
            title: 'A turn of misfortune',
            body: page.error?.message ?? 'An unexpected error interrupted your reign.'
        };
    });
</script>

<div class="pointer-events-none fixed inset-0 -z-10 overflow-hidden">
    <div
        class="absolute left-1/2 top-1/2 h-[34rem] w-[34rem] -translate-x-1/2 -translate-y-1/2 rounded-full"
        style="background: radial-gradient(circle, hsl(49.88deg 73% 50% / 9%) 0%, transparent 70%)"
    ></div>
</div>

<div class="flex flex-1 flex-col items-center justify-center px-6 py-16 text-center">
    <span
        class="select-none font-fancy font-bold leading-none text-primary"
        style="font-size: clamp(5rem, 18vw, 11rem); text-shadow: 3px 3px 0px var(--color-primary-shadow);"
    >
        {status}
    </span>

    <div class="mt-2 flex items-center gap-3 opacity-40">
        <div class="h-px w-20 bg-gradient-to-r from-transparent to-primary"></div>
        <div class="h-1.5 w-1.5 rotate-45 bg-primary"></div>
        <div class="h-px w-20 bg-gradient-to-l from-transparent to-primary"></div>
    </div>

    <h1 class="mt-6 font-fancy text-2xl font-semibold text-font-clear sm:text-3xl">
        {flavour.title}
    </h1>
    <p class="mt-3 max-w-md text-sm leading-relaxed text-font-dimer">
        {flavour.body}
    </p>

    <div class="mt-9 flex flex-wrap items-center justify-center gap-3">
        <a
            href="/"
            class="flex items-center gap-2 rounded-lg bg-gradient-primary px-6 py-2.5 font-fancy text-sm font-bold tracking-wider text-black transition-all duration-150 hover:brightness-125"
        >
            <Home class="h-4 w-4" />
            Back to safety
        </a>
        <a
            href="/leaderboard"
            class="flex items-center gap-2 rounded-lg border border-card-edge bg-card px-6 py-2.5 font-fancy text-sm tracking-wider text-font-dim transition-colors duration-150 hover:text-font-clear"
        >
            <Swords class="h-4 w-4" />
            Leaderboard
        </a>
    </div>
</div>
