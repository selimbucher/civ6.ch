<script lang="ts">
    import { ChevronDown, Search } from '@lucide/svelte';

    // `img` is an avatar URL; `fallback` is text whose first letter is shown in
    // a circle when there's no image (mirrors Avatar.svelte's initial fallback).
    type Item = { value: string; label: string; img?: string; fallback?: string };
    let {
        value = $bindable(''),
        items,
        placeholder = 'Select…',
        searchable = true,
        onChange
    }: {
        value?: string;
        items: Item[];
        placeholder?: string;
        searchable?: boolean;
        onChange?: (value: string) => void;
    } = $props();

    let open = $state(false);
    let search = $state('');
    let trigger = $state<HTMLButtonElement>();
    let menuStyle = $state('');

    const selected = $derived(items.find((i) => i.value === value));
    const filtered = $derived(
        search ? items.filter((i) => i.label.toLowerCase().includes(search.toLowerCase())) : items
    );

    // The menu is position:fixed so it escapes any overflow-hidden card,
    // positioned under the trigger. Close on scroll/resize so it can't drift.
    function toggle() {
        if (!open && trigger) {
            const r = trigger.getBoundingClientRect();
            menuStyle = `left:${r.left}px; top:${r.bottom + 4}px; width:${r.width}px;`;
        }
        open = !open;
        if (!open) search = '';
    }
    function pick(v: string) {
        value = v;
        onChange?.(v);
        open = false;
        search = '';
    }
</script>

<svelte:window onscroll={() => (open = false)} onresize={() => (open = false)} />

<div class="relative">
    <button bind:this={trigger} type="button" onclick={toggle}
        class="w-full flex items-center justify-between gap-2 rounded-lg border bg-card-2 px-3 py-2 text-sm transition-colors duration-150 cursor-pointer
               {open ? 'border-primary/40' : 'border-card-edge hover:border-card-edge-2'}">
        <span class="flex items-center gap-2 min-w-0">
            {#if selected?.img}
                <img src={selected.img} alt="" class="h-6 w-6 rounded-full object-cover shrink-0" />
            {:else if selected?.fallback}
                <span class="flex h-6 w-6 items-center justify-center rounded-full bg-card-2 text-[9px] font-bold font-fancy text-primary select-none shrink-0">
                    {selected.fallback.charAt(0).toUpperCase()}
                </span>
            {/if}
            <span class="truncate {selected ? 'text-font-clear' : 'text-font-dimest'}">
                {selected?.label ?? placeholder}
            </span>
        </span>
        <ChevronDown class="h-4 w-4 text-font-dimer shrink-0 transition-transform duration-150 {open ? 'rotate-180' : ''}" />
    </button>

    {#if open}
        <button type="button" class="fixed inset-0 z-40 cursor-default" tabindex="-1" aria-hidden="true"
            onclick={() => (open = false)}></button>
        <div style="position:fixed; {menuStyle}"
            class="z-50 rounded-lg border border-card-edge bg-card shadow-lg shadow-darken overflow-hidden">
            {#if searchable}
                <div class="flex items-center gap-2 px-3 py-2">
                    <Search class="h-3.5 w-3.5 text-font-dimest shrink-0" />
                    <!-- svelte-ignore a11y_autofocus -->
                    <input bind:value={search} autofocus placeholder="Search…"
                        class="w-full border-none bg-transparent text-sm text-font-clear outline-none placeholder:text-font-dimest" />
                </div>
            {/if}
            <div class="scroll-area max-h-60 overflow-y-auto py-1">
                {#each filtered as it}
                    <button type="button" onclick={() => pick(it.value)}
                        class="w-full flex items-center gap-2 text-left px-3 py-1.5 text-sm transition-colors duration-100 cursor-pointer
                               {it.value === value ? 'bg-primary/15 text-primary' : 'text-font-dim hover:bg-select hover:text-font-clear'}">
                        {#if it.img}
                            <img src={it.img} alt="" class="h-7 w-7 rounded-full object-cover shrink-0" />
                        {:else if it.fallback}
                            <span class="flex h-7 w-7 items-center justify-center rounded-full bg-card-2 text-[10px] font-bold font-fancy text-primary select-none shrink-0">
                                {it.fallback.charAt(0).toUpperCase()}
                            </span>
                        {/if}
                        <span class="truncate">{it.label}</span>
                    </button>
                {:else}
                    <div class="px-3 py-2 text-sm text-font-dimest italic">No matches.</div>
                {/each}
            </div>
        </div>
    {/if}
</div>

<style>
    .scroll-area {
        scrollbar-width: thin;
        scrollbar-color: var(--color-card-edge-2) transparent;
    }
    .scroll-area::-webkit-scrollbar {
        width: 6px;
    }
    .scroll-area::-webkit-scrollbar-track {
        background: transparent;
    }
    .scroll-area::-webkit-scrollbar-thumb {
        background-color: var(--color-card-edge-2);
        border-radius: 9999px;
    }
    .scroll-area::-webkit-scrollbar-thumb:hover {
        background-color: var(--color-font-dimest);
    }
</style>
