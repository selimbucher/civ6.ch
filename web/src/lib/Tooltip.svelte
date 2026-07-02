<script lang="ts">
    import type { Snippet } from 'svelte';

    let {
        children,
        content,
        label = '',
        placement = 'top',
        align = 'center',
        class: cls = ''
    }: {
        /** The trigger element(s). */
        children: Snippet;
        /** Rich tooltip body. Takes precedence over `label`. */
        content?: Snippet;
        /** Simple text body. */
        label?: string;
        /** Which side of the trigger the bubble appears on. */
        placement?: 'top' | 'bottom';
        /** Horizontal anchoring — useful near viewport edges. */
        align?: 'center' | 'left' | 'right';
        class?: string;
    } = $props();

    const alignClass = $derived(
        align === 'left'
            ? 'left-0'
            : align === 'right'
              ? 'right-0'
              : 'left-1/2 -translate-x-1/2'
    );

    const arrowAlignClass = $derived(
        align === 'left' ? 'left-3' : align === 'right' ? 'right-3' : 'left-1/2 -translate-x-1/2'
    );
</script>

<span class="group/tt relative inline-flex {cls}">
    {@render children()}
    <span
        role="tooltip"
        class="pointer-events-none absolute z-50 w-max max-w-[16rem] rounded-lg border border-card-edge-2 bg-card-2 px-2.5 py-1.5 text-xs leading-snug text-font-dim opacity-0 shadow-lg shadow-darken transition-[opacity,transform] duration-100 ease-out group-hover/tt:opacity-100
               {alignClass}
               {placement === 'top'
            ? 'bottom-full mb-2 origin-bottom translate-y-1 group-hover/tt:translate-y-0'
            : 'top-full mt-2 origin-top -translate-y-1 group-hover/tt:translate-y-0'}"
    >
        {#if content}{@render content()}{:else}{label}{/if}
        <span
            class="absolute h-2 w-2 rotate-45 border-card-edge-2 bg-card-2 {arrowAlignClass}
                   {placement === 'top' ? 'top-full -mt-1 border-b border-r' : 'bottom-full -mb-1 border-l border-t'}"
        ></span>
    </span>
</span>
