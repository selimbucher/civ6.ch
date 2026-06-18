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

<div class="mx-3 md:mx-12 mb-12 flex flex-col gap-4 max-w-3xl">

    <!-- Steam accounts card -->
    <div class="rounded-2xl border border-card-edge bg-card shadow-md shadow-darken overflow-hidden">
        <div class="h-[3px] bg-gradient-primary"></div>
        <div class="p-6 flex flex-col gap-5">

            <div class="flex flex-col gap-1">
                <div class="flex items-center gap-2">
                    <ShieldCheck class="h-5 w-5 text-primary" strokeWidth={1.75} />
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
