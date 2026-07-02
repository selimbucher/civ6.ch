<script lang="ts">
    import type { Snippet } from 'svelte';

    let {
        title,
        subtitle = '',
        icon,
        actions,
        class: cls = ''
    }: {
        title: string;
        subtitle?: string;
        /** Decorative icon shown on the right (ignored if `actions` is set). */
        icon?: Snippet;
        /** Right-hand content (e.g. a count or controls). Overrides `icon`. */
        actions?: Snippet;
        class?: string;
    } = $props();
</script>

<div
    class="relative overflow-hidden rounded-2xl border border-card-edge bg-card shadow-md shadow-darken {cls}"
>
    <div class="absolute left-0 top-0 h-0.75 w-full bg-gradient-primary"></div>
    <div class="flex items-center justify-between gap-4 px-6 py-6 md:px-10 md:py-8">
        <div class="min-w-0">
            <h1
                class="font-fancy text-3xl font-bold tracking-wide text-font-clear md:text-4xl"
                style="text-shadow: 1px 1px 0px var(--color-primary-shadow);"
            >
                {title}
            </h1>
            {#if subtitle}
                <p class="mt-1 text-sm text-font-dimer">{subtitle}</p>
            {/if}
        </div>
        {#if actions}
            <div class="shrink-0">{@render actions()}</div>
        {:else if icon}
            <div class="shrink-0 text-primary">{@render icon()}</div>
        {/if}
    </div>
</div>
