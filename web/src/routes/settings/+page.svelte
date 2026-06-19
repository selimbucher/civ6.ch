<script lang="ts">
    import { enhance } from '$app/forms';
    import { page } from '$app/stores';
    import { onMount } from 'svelte';
    import { browser } from '$app/environment';
    import {
        Unlink, ExternalLink, ShieldCheck, Link, User, KeyRound, LogOut,
        Sparkles, PartyPopper, WandSparkles, Rows3, EyeOff, RotateCcw
    } from '@lucide/svelte';
    import type { PageData } from './$types';

    let { data, form }: { data: PageData; form: any } = $props();
    const { steamAccounts } = $derived(data);

    const steamStatus = $derived($page.url.searchParams.get('steam'));

    // Uncontrolled inputs: value derives from server data, so a successful save
    // (which reloads data + resets the form) shows the saved values rather than
    // going blank. The name is stored as "First Last".
    const fullName  = $derived((data.profile?.name ?? '') as string);
    const firstName = $derived(fullName.split(' ')[0] ?? '');
    const lastName  = $derived(fullName.split(' ').slice(1).join(' '));
    const email     = $derived((data.profile?.email ?? '') as string);

    function fmtDate(d: string | Date) {
        return new Date(d).toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' });
    }

    // ── Cosmetic, site-side extras (persisted locally) ─────────────────────────
    type Toggle = { id: string; icon: any; label: string; desc: string };
    const extras: Toggle[] = [
        { id: 'confetti', icon: PartyPopper, label: 'Confetti on your victories', desc: 'Rain confetti on match pages you won. You earned it.' },
        { id: 'dramatic', icon: WandSparkles,       label: 'Dramatic rating reveals',     desc: 'Animate rating changes with entirely unnecessary flourish.' },
        { id: 'compact',  icon: Rows3,       label: 'Compact match feed',          desc: 'Tighter spacing in the matches list for the data-hungry.' },
        { id: 'amnesia',  icon: EyeOff,      label: 'Selective memory',            desc: 'Dim your defeats so your profile is all glorious wins.' }
    ];
    const extraDefaults: Record<string, boolean> = { confetti: true, dramatic: false, compact: false, amnesia: false };

    let prefs = $state<Record<string, boolean>>({ ...extraDefaults });
    let toast = $state<string | null>(null);

    onMount(() => {
        try { prefs = { ...prefs, ...JSON.parse(localStorage.getItem('civ6_prefs') ?? '{}') }; }
        catch { /* keep defaults */ }
    });
    $effect(() => { if (browser) localStorage.setItem('civ6_prefs', JSON.stringify(prefs)); });

    function flash(msg: string) { toast = msg; setTimeout(() => (toast = null), 2400); }
</script>

{#snippet steamIcon(cls: string)}
    <svg class={cls} viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
        <path d="M11.979 0C5.678 0 .511 4.86.022 11.037l6.432 2.658c.545-.371 1.203-.59 1.912-.59.063 0 .125.004.188.006l2.861-4.142V8.91c0-2.495 2.028-4.524 4.524-4.524 2.494 0 4.524 2.031 4.524 4.527s-2.03 4.525-4.524 4.525h-.105l-4.076 2.911c0 .052.004.105.004.159 0 1.875-1.515 3.396-3.39 3.396-1.635 0-3.016-1.173-3.331-2.727L.436 15.27C1.862 20.307 6.486 24 11.979 24c6.627 0 11.999-5.373 11.999-12S18.605 0 11.979 0zM7.54 18.21l-1.473-.61c.262.543.714.999 1.314 1.25 1.297.539 2.793-.076 3.332-1.375.263-.63.264-1.319.005-1.949s-.75-1.121-1.377-1.383c-.624-.26-1.29-.249-1.878-.03l1.523.63c.956.4 1.409 1.5 1.009 2.455-.397.957-1.497 1.41-2.454 1.012H7.54zm11.415-9.303c0-1.662-1.353-3.015-3.015-3.015-1.665 0-3.015 1.353-3.015 3.015 0 1.665 1.35 3.015 3.015 3.015 1.663 0 3.015-1.35 3.015-3.015zm-5.273-.005c0-1.252 1.013-2.266 2.265-2.266 1.249 0 2.266 1.014 2.266 2.266 0 1.251-1.017 2.265-2.266 2.265-1.253 0-2.265-1.014-2.265-2.265z" />
    </svg>
{/snippet}

{#snippet sectionHead(Icon: any, title: string, subtitle: string)}
    <div class="flex flex-col gap-1">
        <div class="flex items-center gap-2">
            <Icon class="h-5 w-5 text-primary" strokeWidth={1.75} />
            <span class="font-fancy text-lg font-semibold text-font-clear">{title}</span>
        </div>
        <p class="text-sm text-font-dimer leading-relaxed">{subtitle}</p>
    </div>
{/snippet}

{#snippet banner(ok: boolean, msg: string)}
    <div class="rounded-lg border px-3 py-2 text-sm
                {ok ? 'border-font-good/30 bg-font-good/10 text-font-good'
                    : 'border-font-bad/30 bg-font-bad/10 text-font-bad'}">
        {ok ? '✓ ' : ''}{msg}
    </div>
{/snippet}

{#snippet field(label: string, name: string, type: string, value: string, placeholder: string)}
    <label class="flex flex-col gap-1.5 flex-1">
        <span class="text-xs font-fancy tracking-wide uppercase text-font-dimest">{label}</span>
        <input {name} {type} {value} {placeholder} autocomplete="off"
            class="rounded-lg border border-card-edge bg-card-2 px-3 py-2 text-sm text-font-clear
                   outline-none focus:border-primary/40 placeholder:text-font-dimest" />
    </label>
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

<div class="mx-auto w-full max-w-3xl px-3 md:px-6 mb-12 flex flex-col gap-4">

    <div class="flex flex-col gap-1 mt-2">
        <h1 class="font-fancy text-2xl font-semibold text-font-clear">Preferences</h1>
        <p class="text-sm text-font-dimer">Manage your account and how you appear on civ6.ch.</p>
    </div>

    <!-- Profile -->
    <div class="rounded-2xl border border-card-edge bg-card shadow-md shadow-darken overflow-hidden">
        <div class="h-[3px] bg-gradient-primary"></div>
        <form method="POST" action="?/profile" use:enhance class="p-6 flex flex-col gap-5">
            {@render sectionHead(User, 'Profile', 'Your name as shown across civ6.ch, and your contact email.')}

            {#if form?.profileOk}{@render banner(true, 'Profile updated.')}{/if}
            {#if form?.profileError}{@render banner(false, form.profileError)}{/if}

            <div class="flex flex-col gap-4">
                <div class="flex flex-col sm:flex-row gap-4">
                    {@render field('First name', 'first', 'text', firstName, '')}
                    {@render field('Last name', 'last', 'text', lastName, '')}
                </div>
                <div class="flex flex-col gap-1">
                    {@render field('Email', 'email', 'email', email, 'you@example.com')}
                    <span class="text-xs text-font-dimest">Used to sign in and recover your account. Never shown publicly.</span>
                </div>
                <div class="flex flex-col gap-1">
                    {@render field('Current password', 'password', 'password', '', '')}
                    <span class="text-xs text-font-dimest">Required to confirm changes to your account.</span>
                </div>
            </div>

            <div>
                <button type="submit"
                    class="inline-flex items-center gap-2 rounded-lg px-4 py-2 text-sm font-semibold
                           bg-gradient-primary text-black hover:brightness-125 transition-all duration-150">
                    Save changes
                </button>
            </div>
        </form>
    </div>

    <!-- Security -->
    <div class="rounded-2xl border border-card-edge bg-card shadow-md shadow-darken overflow-hidden">
        <div class="h-[3px] bg-gradient-primary"></div>
        <div class="p-6 flex flex-col gap-5">
            {@render sectionHead(KeyRound, 'Security', 'Change your password and manage active sessions.')}

            {#if form?.passwordOk}{@render banner(true, 'Password changed.')}{/if}
            {#if form?.passwordError}{@render banner(false, form.passwordError)}{/if}
            {#if form?.signedOut}{@render banner(true, 'Signed out of all other devices.')}{/if}

            <form method="POST" action="?/password" use:enhance class="flex flex-col gap-4">
                {@render field('Current password', 'current', 'password', '', '')}
                <div class="flex flex-col sm:flex-row gap-4">
                    {@render field('New password', 'new', 'password', '', '')}
                    {@render field('Confirm new password', 'confirm', 'password', '', '')}
                </div>
                <div>
                    <button type="submit"
                        class="inline-flex items-center gap-2 rounded-lg px-4 py-2 text-sm font-semibold
                               bg-gradient-primary text-black hover:brightness-125 transition-all duration-150">
                        Change password
                    </button>
                </div>
            </form>

            <div class="border-t border-card-edge pt-4 flex items-center gap-4">
                <div class="flex flex-col leading-tight flex-1 min-w-0">
                    <span class="text-sm text-font-clear">Sign out everywhere else</span>
                    <span class="text-xs text-font-dimest mt-0.5">Ends every other session. This device stays signed in.</span>
                </div>
                <form method="POST" action="?/signout_all" use:enhance class="shrink-0">
                    <button type="submit"
                        class="flex items-center gap-1.5 rounded-lg border border-card-edge px-3 py-1.5 text-sm text-font-dimer
                               hover:border-font-bad/40 hover:text-font-bad transition-colors duration-150 cursor-pointer">
                        <LogOut class="h-3.5 w-3.5" strokeWidth={1.5} /> Sign out others
                    </button>
                </form>
            </div>
        </div>
    </div>

    <!-- Steam account -->
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

            {#if steamStatus === 'linked'}{@render banner(true, 'Steam account linked.')}{/if}
            {#if steamStatus === 'error'}{@render banner(false, 'Could not verify your Steam account. Please try again.')}{/if}

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

    <!-- Extras (cosmetic) -->
    <div class="rounded-2xl border border-card-edge bg-card shadow-md shadow-darken overflow-hidden">
        <div class="h-[3px] bg-gradient-primary"></div>
        <div class="p-6 flex flex-col gap-5">
            {@render sectionHead(Sparkles, 'Extras', 'Purely cosmetic. Saved to this device, harmless to your rating.')}
            <div class="flex flex-col divide-y divide-card-edge">
                {#each extras as item}
                    <div class="flex items-center gap-4 py-3 first:pt-0 last:pb-0">
                        <item.icon class="h-4.5 w-4.5 text-font-dimer shrink-0" strokeWidth={1.5} />
                        <div class="flex flex-col leading-tight flex-1 min-w-0">
                            <span class="text-sm text-font-clear">{item.label}</span>
                            <span class="text-xs text-font-dimest mt-0.5">{item.desc}</span>
                        </div>
                        {@render toggle(item.id)}
                    </div>
                {/each}
            </div>
            <div class="flex justify-end">
                <button type="button" onclick={() => { prefs = { ...extraDefaults }; flash('Extras reset to defaults.'); }}
                    class="flex items-center gap-1.5 text-xs text-font-dimest hover:text-font-dim transition-colors duration-150 cursor-pointer">
                    <RotateCcw class="h-3.5 w-3.5" /> Reset extras
                </button>
            </div>
        </div>
    </div>
</div>

{#if toast}
    <div class="fixed bottom-6 left-1/2 -translate-x-1/2 z-50 rounded-xl border border-card-edge bg-card px-4 py-2.5
                text-sm text-font-clear shadow-lg shadow-darken">
        {toast}
    </div>
{/if}
