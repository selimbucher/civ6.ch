<script lang="ts">
    import { enhance } from '$app/forms';
    import { page } from '$app/stores';
    import { Link2, Unlink, ExternalLink, ShieldCheck } from '@lucide/svelte';
    import type { PageData } from './$types';

    let { data }: { data: PageData } = $props();
    const { steamAccounts } = $derived(data);

    const steamStatus = $derived($page.url.searchParams.get('steam'));

    function fmtDate(d: string | Date) {
        return new Date(d).toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' });
    }
</script>

{#snippet steamIcon(cls: string)}
    <svg class={cls} viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
        <path d="M11.979 0C5.678 0 .511 4.86.022 11.037l6.432 2.658c.545-.371 1.203-.59 1.912-.59.063 0 .125.004.188.006l2.861-4.142V8.91c0-2.495 2.028-4.524 4.524-4.524 2.494 0 4.524 2.031 4.524 4.527s-2.03 4.525-4.524 4.525h-.105l-4.076 2.911c0 .052.004.105.004.159 0 1.875-1.515 3.396-3.39 3.396-1.635 0-3.016-1.173-3.331-2.727L.436 15.27C1.862 20.307 6.486 24 11.979 24c6.627 0 11.999-5.373 11.999-12S18.605 0 11.979 0zM7.54 18.21l-1.473-.61c.262.543.714.999 1.314 1.25 1.297.539 2.793-.076 3.332-1.375.263-.63.264-1.319.005-1.949s-.75-1.121-1.377-1.383c-.624-.26-1.29-.249-1.878-.03l1.523.63c.956.4 1.409 1.5 1.009 2.455-.397.957-1.497 1.41-2.454 1.012H7.54zm11.415-9.303c0-1.662-1.353-3.015-3.015-3.015-1.665 0-3.015 1.353-3.015 3.015 0 1.665 1.35 3.015 3.015 3.015 1.663 0 3.015-1.35 3.015-3.015zm-5.273-.005c0-1.252 1.013-2.266 2.265-2.266 1.249 0 2.266 1.014 2.266 2.266 0 1.251-1.017 2.265-2.266 2.265-1.253 0-2.265-1.014-2.265-2.265z" />
    </svg>
{/snippet}

<div class="mx-3 md:mx-12 mb-12 flex flex-col gap-4 max-w-3xl">

    <!-- Steam accounts card -->
    <div class="rounded-2xl border border-card-edge bg-card shadow-md shadow-darken overflow-hidden">
        <div class="h-[3px] bg-gradient-primary"></div>
        <div class="p-6 flex flex-col gap-5">

            <div class="flex flex-col gap-1">
                <div class="flex items-center gap-2">
                    {@render steamIcon('h-5 w-5 text-primary')}
                    <span class="font-fancy text-lg font-semibold text-font-clear">Steam Accounts</span>
                </div>
                <p class="text-sm text-font-dimer leading-relaxed">
                    Link your Steam account so uploaded games recognise you automatically.
                </p>
            </div>

            <!-- Linked accounts -->
            {#if steamAccounts.length > 0}
                <div class="flex flex-col gap-2">
                    {#each steamAccounts as acc}
                        <div class="flex items-center gap-3 rounded-xl border border-card-edge bg-card-2 px-4 py-3">
                            <div class="h-9 w-9 rounded-full bg-primary/10 border border-primary/20 flex items-center justify-center shrink-0">
                                <ShieldCheck class="h-4.5 w-4.5 text-primary" strokeWidth={1.75} />
                            </div>
                            <div class="flex flex-col leading-tight min-w-0 flex-1">
                                <span class="text-sm text-font-clear font-medium truncate">
                                    {acc.persona ?? 'Steam account'}
                                </span>
                                <a href="https://steamcommunity.com/profiles/{acc.steam_id}"
                                   target="_blank" rel="noopener"
                                   class="text-xs text-font-dimest hover:text-primary transition-colors duration-150 flex items-center gap-1 w-fit">
                                    {acc.steam_id}<ExternalLink class="h-3 w-3" />
                                </a>
                            </div>
                            <span class="text-xs text-font-dimest hidden sm:block">linked {fmtDate(acc.linked_at)}</span>
                            <form method="POST" action="?/unlink" use:enhance class="shrink-0">
                                <input type="hidden" name="steam_id" value={acc.steam_id} />
                                <button type="submit"
                                    class="flex items-center gap-1.5 rounded-lg border border-card-edge px-3 py-1.5 text-xs text-font-dimer
                                           hover:border-font-bad/40 hover:text-font-bad transition-colors duration-150 cursor-pointer">
                                    <Unlink class="h-3.5 w-3.5" strokeWidth={1.5} /> Unlink
                                </button>
                            </form>
                        </div>
                    {/each}
                </div>
            {:else}
                <p class="text-sm text-font-dimest italic">No Steam account linked yet.</p>
            {/if}

            <!-- Link button -->
            <div>
                <a href="/auth/steam"
                   class="inline-flex items-center gap-2 rounded-lg px-4 py-2.5 text-sm font-semibold
                          bg-gradient-primary text-black hover:brightness-125 transition-all duration-150">
                    <Link2 class="h-4 w-4" strokeWidth={2} />
                    {steamAccounts.length > 0 ? 'Link another Steam account' : 'Link Steam account'}
                </a>
            </div>
        </div>
    </div>

    {#if steamStatus === 'linked'}
        <div class="rounded-xl border border-font-good/30 bg-font-good/10 px-4 py-2 text-sm text-font-good">
            ✓ Steam account linked.
        </div>
    {:else if steamStatus === 'error'}
        <div class="rounded-xl border border-font-bad/30 bg-font-bad/10 px-4 py-2 text-sm text-font-bad">
            Could not verify your Steam account. Please try again.
        </div>
    {/if}
    
</div>
