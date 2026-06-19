<script lang="ts">
    import { enhance } from '$app/forms';
    import { page } from '$app/stores';
    import { onMount } from 'svelte';
    import { browser } from '$app/environment';
    import {
        Unlink, ExternalLink, ShieldCheck, Link,
        Bell, Gamepad2, Palette, Swords, Scroll, Sparkles, Crown, Flame, RotateCcw
    } from '@lucide/svelte';
    import type { PageData } from './$types';

    let { data }: { data: PageData } = $props();
    const { steamAccounts } = $derived(data);

    const steamStatus = $derived($page.url.searchParams.get('steam'));

    function fmtDate(d: string | Date) {
        return new Date(d).toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' });
    }

    // ── Preferences (purely cosmetic, persisted locally) ───────────────────────
    type Toggle = { id: string; label: string; desc: string };
    type Select = { id: string; label: string; desc: string; options: string[] };

    const toggleSections: { icon: any; title: string; items: Toggle[] }[] = [
        {
            icon: Bell, title: 'Notifications', items: [
                { id: 'barb_alerts',  label: 'Barbarian proximity alerts', desc: 'Warn me when barbarians camp within three tiles of my comfort zone.' },
                { id: 'denounce',     label: 'Denouncement digest',        desc: 'A weekly summary of every leader who publicly denounced you.' },
                { id: 'eureka',       label: 'Eureka fanfare',             desc: 'Play a triumphant chord each time a boost is earned.' },
                { id: 'one_more',     label: 'One More Turn™ reminder',     desc: 'Gently remind you that it is, in fact, 3 AM.' }
            ]
        },
        {
            icon: Gamepad2, title: 'Gameplay', items: [
                { id: 'grievances',   label: 'Eternal grievance ledger',   desc: 'Never forget a single slight. Ever. Bonus spite at war declarations.' },
                { id: 'auto_renounce',label: 'Pragmatic friendships',      desc: 'End friendships the moment they stop being strategically useful.' },
                { id: 'loyalty',      label: 'Emotional loyalty pressure',  desc: 'Apply loyalty pressure to nearby coworkers and roommates.' },
                { id: 'warmonger',    label: 'Warmonger guilt',            desc: 'Feel a faint pang of remorse after each surprise war.' }
            ]
        }
    ];

    const selects: Select[] = [
        { id: 'voice',   label: 'Leader voice',     desc: 'How the game addresses you in defeat.',         options: ['Default', 'Aggressive', 'Passive-aggressive', 'Suspiciously friendly', 'Trajan (Latin only)'] },
        { id: 'era',     label: 'Lobby soundtrack', desc: 'Background music while you wait for the host.',   options: ['Ancient', 'Medieval', 'Atomic', 'Elevator'] },
        { id: 'mapTint', label: 'Map preview tint', desc: 'Aesthetic filter applied to match maps.',         options: ['None', 'Sepia (Old Map)', 'Apocalypse Red', 'Blueprint'] }
    ];

    function defaults(): Record<string, boolean | string> {
        const d: Record<string, boolean | string> = {
            barb_alerts: true, denounce: false, eureka: true, one_more: false,
            grievances: true, auto_renounce: false, loyalty: false, warmonger: false
        };
        for (const s of selects) d[s.id] = s.options[0];
        return d;
    }

    let prefs = $state<Record<string, boolean | string>>(defaults());
    let toast = $state<string | null>(null);

    onMount(() => {
        try {
            const saved = JSON.parse(localStorage.getItem('civ6_prefs') ?? '{}');
            prefs = { ...prefs, ...saved };
        } catch { /* corrupted prefs, keep defaults */ }
    });

    $effect(() => {
        if (browser) localStorage.setItem('civ6_prefs', JSON.stringify(prefs));
    });

    function flash(msg: string) {
        toast = msg;
        setTimeout(() => (toast = null), 2600);
    }
</script>

{#snippet steamIcon(cls: string)}
    <svg class={cls} viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
        <path d="M11.979 0C5.678 0 .511 4.86.022 11.037l6.432 2.658c.545-.371 1.203-.59 1.912-.59.063 0 .125.004.188.006l2.861-4.142V8.91c0-2.495 2.028-4.524 4.524-4.524 2.494 0 4.524 2.031 4.524 4.527s-2.03 4.525-4.524 4.525h-.105l-4.076 2.911c0 .052.004.105.004.159 0 1.875-1.515 3.396-3.39 3.396-1.635 0-3.016-1.173-3.331-2.727L.436 15.27C1.862 20.307 6.486 24 11.979 24c6.627 0 11.999-5.373 11.999-12S18.605 0 11.979 0zM7.54 18.21l-1.473-.61c.262.543.714.999 1.314 1.25 1.297.539 2.793-.076 3.332-1.375.263-.63.264-1.319.005-1.949s-.75-1.121-1.377-1.383c-.624-.26-1.29-.249-1.878-.03l1.523.63c.956.4 1.409 1.5 1.009 2.455-.397.957-1.497 1.41-2.454 1.012H7.54zm11.415-9.303c0-1.662-1.353-3.015-3.015-3.015-1.665 0-3.015 1.353-3.015 3.015 0 1.665 1.35 3.015 3.015 3.015 1.663 0 3.015-1.35 3.015-3.015zm-5.273-.005c0-1.252 1.013-2.266 2.265-2.266 1.249 0 2.266 1.014 2.266 2.266 0 1.251-1.017 2.265-2.266 2.265-1.253 0-2.265-1.014-2.265-2.265z" />
    </svg>
{/snippet}

{#snippet sectionHead(Icon: any, title: string)}
    <div class="flex items-center gap-2">
        <Icon class="h-5 w-5 text-primary" strokeWidth={1.75} />
        <span class="font-fancy text-lg font-semibold text-font-clear">{title}</span>
    </div>
{/snippet}

{#snippet toggle(id: string)}
    {@const on = prefs[id] === true}
    <button type="button" role="switch" aria-checked={on} aria-label="toggle"
        onclick={() => (prefs[id] = !on)}
        class="relative h-5.5 w-10 rounded-full transition-colors duration-150 shrink-0 cursor-pointer
               {on ? 'bg-primary' : 'bg-card-edge-2'}">
        <span class="absolute top-0.5 left-0.5 h-4.5 w-4.5 rounded-full bg-font-clear shadow-sm transition-transform duration-150
                     {on ? 'translate-x-4.5' : ''}"></span>
    </button>
{/snippet}

<div class="mx-3 md:mx-12 mb-12 flex flex-col gap-4 max-w-3xl">

    <div class="flex flex-col gap-1 mt-2">
        <h1 class="font-fancy text-2xl font-semibold text-font-clear">Preferences</h1>
        <p class="text-sm text-font-dimer">Tune your experience. None of this affects your rating — we checked.</p>
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

    <!-- Steam accounts card -->
    <div class="rounded-2xl border border-card-edge bg-card shadow-md shadow-darken overflow-hidden">
        <div class="h-[3px] bg-gradient-primary"></div>
        <div class="p-6 flex flex-col gap-5">

            <div class="flex flex-col gap-1">
                <div class="flex items-center gap-2">
                    {@render steamIcon('h-5 w-5 text-primary')}
                    <span class="font-fancy text-lg font-semibold text-font-clear">Steam Account</span>
                </div>
                <p class="text-sm text-font-dimer leading-relaxed">
                    Link your Steam account so uploaded games recognise you automatically.
                </p>
            </div>

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
                <p class="text-sm text-font-dimest italic">No account linked yet.</p>
            {/if}

            <div>
                <a href="/auth/steam"
                   class="inline-flex items-center gap-2 rounded-lg px-4 py-2.5 text-sm font-semibold
                          bg-gradient-primary text-black hover:brightness-125 transition-all duration-150">
                    <Link class="h-4 w-4" strokeWidth={2} />
                    {steamAccounts.length > 0 ? 'Link another account' : 'Link Steam account'}
                </a>
            </div>
        </div>
    </div>

    <!-- Toggle sections -->
    {#each toggleSections as section}
        <div class="rounded-2xl border border-card-edge bg-card shadow-md shadow-darken overflow-hidden">
            <div class="h-[3px] bg-gradient-primary"></div>
            <div class="p-6 flex flex-col gap-5">
                {@render sectionHead(section.icon, section.title)}
                <div class="flex flex-col divide-y divide-card-edge">
                    {#each section.items as item}
                        <div class="flex items-center gap-4 py-3 first:pt-0 last:pb-0">
                            <div class="flex flex-col leading-tight flex-1 min-w-0">
                                <span class="text-sm text-font-clear">{item.label}</span>
                                <span class="text-xs text-font-dimest mt-0.5">{item.desc}</span>
                            </div>
                            {@render toggle(item.id)}
                        </div>
                    {/each}
                </div>
            </div>
        </div>
    {/each}

    <!-- Appearance / selects -->
    <div class="rounded-2xl border border-card-edge bg-card shadow-md shadow-darken overflow-hidden">
        <div class="h-[3px] bg-gradient-primary"></div>
        <div class="p-6 flex flex-col gap-5">
            {@render sectionHead(Palette, 'Flavor')}
            <div class="flex flex-col divide-y divide-card-edge">
                {#each selects as s}
                    <div class="flex items-center gap-4 py-3 first:pt-0 last:pb-0">
                        <div class="flex flex-col leading-tight flex-1 min-w-0">
                            <span class="text-sm text-font-clear">{s.label}</span>
                            <span class="text-xs text-font-dimest mt-0.5">{s.desc}</span>
                        </div>
                        <select bind:value={prefs[s.id]}
                            class="shrink-0 rounded-lg border border-card-edge bg-card-2 px-3 py-1.5 text-sm text-font-dim
                                   outline-none focus:border-primary/40 cursor-pointer">
                            {#each s.options as opt}
                                <option value={opt}>{opt}</option>
                            {/each}
                        </select>
                    </div>
                {/each}
            </div>
        </div>
    </div>

    <!-- Danger zone -->
    <div class="rounded-2xl border border-font-bad/30 bg-card shadow-md shadow-darken overflow-hidden">
        <div class="h-[3px] bg-font-bad/70"></div>
        <div class="p-6 flex flex-col gap-5">
            <div class="flex items-center gap-2">
                <Flame class="h-5 w-5 text-font-bad" strokeWidth={1.75} />
                <span class="font-fancy text-lg font-semibold text-font-clear">Danger Zone</span>
            </div>

            <div class="flex flex-col gap-3">
                <div class="flex items-center gap-4">
                    <div class="flex flex-col leading-tight flex-1 min-w-0">
                        <span class="text-sm text-font-clear">Sue for peace with everyone</span>
                        <span class="text-xs text-font-dimest mt-0.5">Immediately end all wars. Peace is, of course, temporary.</span>
                    </div>
                    <button type="button" onclick={() => flash('☮ A fragile peace settles over the realm.')}
                        class="shrink-0 flex items-center gap-1.5 rounded-lg border border-card-edge px-3 py-1.5 text-sm text-font-dimer
                               hover:border-primary/40 hover:text-primary transition-colors duration-150 cursor-pointer">
                        <Scroll class="h-3.5 w-3.5" strokeWidth={1.5} /> Sue for peace
                    </button>
                </div>

                <div class="flex items-center gap-4">
                    <div class="flex flex-col leading-tight flex-1 min-w-0">
                        <span class="text-sm text-font-clear">Forgive all grievances</span>
                        <span class="text-xs text-font-dimest mt-0.5">Wipe the ledger clean. You will absolutely remember anyway.</span>
                    </div>
                    <button type="button" onclick={() => flash('You feel lighter. The ledger remembers.')}
                        class="shrink-0 flex items-center gap-1.5 rounded-lg border border-card-edge px-3 py-1.5 text-sm text-font-dimer
                               hover:border-primary/40 hover:text-primary transition-colors duration-150 cursor-pointer">
                        <Swords class="h-3.5 w-3.5" strokeWidth={1.5} /> Forgive
                    </button>
                </div>

                <div class="flex items-center gap-4">
                    <div class="flex flex-col leading-tight flex-1 min-w-0">
                        <span class="text-sm text-font-bad font-medium">Abdicate the throne</span>
                        <span class="text-xs text-font-dimest mt-0.5">Permanently dissolve your dynasty and forfeit all glory.</span>
                    </div>
                    <button type="button"
                        onclick={() => flash('A successor was found within the hour. Long live the monarch.')}
                        class="shrink-0 flex items-center gap-1.5 rounded-lg px-3 py-1.5 text-sm font-semibold
                               bg-font-bad/15 text-font-bad border border-font-bad/30 hover:bg-font-bad/25 transition-colors duration-150 cursor-pointer">
                        <Crown class="h-3.5 w-3.5" strokeWidth={1.75} /> Abdicate
                    </button>
                </div>
            </div>
        </div>
    </div>

    <div class="flex items-center justify-between text-xs text-font-dimest px-1">
        <span class="flex items-center gap-1.5">
            <Sparkles class="h-3.5 w-3.5" /> Preferences save automatically to this device.
        </span>
        <button type="button" onclick={() => { prefs = defaults(); flash('Preferences reset to factory settings.'); }}
            class="flex items-center gap-1.5 hover:text-font-dim transition-colors duration-150 cursor-pointer">
            <RotateCcw class="h-3.5 w-3.5" /> Reset to defaults
        </button>
    </div>
</div>

<!-- Toast -->
{#if toast}
    <div class="fixed bottom-6 left-1/2 -translate-x-1/2 z-50 rounded-xl border border-card-edge bg-card px-4 py-2.5
                text-sm text-font-clear shadow-lg shadow-darken">
        {toast}
    </div>
{/if}
