<script lang="ts">
    import type { Snippet } from 'svelte';

    let {
        title,
        subtitle = '',
        actions,
        class: cls = ''
    }: {
        title: string;
        subtitle?: string;
        /** Right-hand content (e.g. a count or controls). */
        actions?: Snippet;
        class?: string;
    } = $props();
</script>

<div class="relative {cls}">
    <div class="flex items-end justify-between gap-4 border-b border-card-edge px-1 pb-5 pt-2">
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
        {/if}
    </div>
</div>
