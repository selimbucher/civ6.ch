<script lang="ts">
    import { enhance } from '$app/forms';
    import { page } from '$app/stores';
    import {
        Unlink, ExternalLink, ShieldCheck, Link, User, KeyRound, LogOut,
        Bell, Megaphone, Angry, HeartHandshake, Scroll, Crown, Flame, ChevronDown, Search
    } from '@lucide/svelte';
    import type { PageData } from './$types';

    let { data, form }: { data: PageData; form: any } = $props();
    const { steamAccounts, denounced, players } = $derived(data);

    const steamStatus = $derived($page.url.searchParams.get('steam'));

    // Uncontrolled profile inputs (value from server data) so saved values show
    // after submit. Name is stored as "First Last".
    const fullName  = $derived((data.profile?.name ?? '') as string);
    const firstName = $derived(fullName.split(' ')[0] ?? '');
    const lastName  = $derived(fullName.split(' ').slice(1).join(' '));
    const email     = $derived((data.profile?.email ?? '') as string);

    const notifyMeta = [
        { key: 'new_game',    label: 'A game you played is logged', desc: 'A toga-clad courier sprints to your inbox when a match you were in is uploaded.' },
        { key: 'denounced',   label: 'A rival denounces you',       desc: 'Hear the bad news first, by carrier pigeon (email).' },
        { key: 'weekly',      label: 'The weekly herald',           desc: 'A Sunday scroll summarising the standings and fresh grudges.' },
        { key: 'achievement', label: 'You unlock an achievement',   desc: 'Silent trumpets, delivered straight to your email.' }
    ];
    let notif = $state<Record<string, boolean>>({});
    $effect(() => { notif = { ...(data.notify ?? {}) }; });

    let denounceTarget = $state('');
    let denounceOpen = $state(false);
    let denounceSearch = $state('');
    const selectedName = $derived(players.find((p: any) => String(p.id) === denounceTarget)?.name ?? '');
    const filteredPlayers = $derived(
        players.filter((p: any) => p.name.toLowerCase().includes(denounceSearch.toLowerCase()))
    );
</script>

{#snippet steamIcon(cls: string)}
    <svg class={cls} viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
        <path d="M11.979 0C5.678 0 .511 4.86.022 11.037l6.432 2.658c.545-.371 1.203-.59 1.912-.59.063 0 .125.004.188.006l2.861-4.142V8.91c0-2.495 2.028-4.524 4.524-4.524 2.494 0 4.524 2.031 4.524 4.527s-2.03 4.525-4.524 4.525h-.105l-4.076 2.911c0 .052.004.105.004.159 0 1.875-1.515 3.396-3.39 3.396-1.635 0-3.016-1.173-3.331-2.727L.436 15.27C1.862 20.307 6.486 24 11.979 24c6.627 0 11.999-5.373 11.999-12S18.605 0 11.979 0zM7.54 18.21l-1.473-.61c.262.543.714.999 1.314 1.25 1.297.539 2.793-.076 3.332-1.375.263-.63.264-1.319.005-1.949s-.75-1.121-1.377-1.383c-.624-.26-1.29-.249-1.878-.03l1.523.63c.956.4 1.409 1.5 1.009 2.455-.397.957-1.497 1.41-2.454 1.012H7.54zm11.415-9.303c0-1.662-1.353-3.015-3.015-3.015-1.665 0-3.015 1.353-3.015 3.015 0 1.665 1.35 3.015 3.015 3.015 1.663 0 3.015-1.35 3.015-3.015zm-5.273-.005c0-1.252 1.013-2.266 2.265-2.266 1.249 0 2.266 1.014 2.266 2.266 0 1.251-1.017 2.265-2.266 2.265-1.253 0-2.265-1.014-2.265-2.265z" />
    </svg>
{/snippet}

{#snippet head(Icon: any, title: string)}
    <div class="flex items-center gap-2">
        <Icon class="h-5 w-5 text-primary" strokeWidth={1.75} />
        <span class="font-fancy text-lg font-semibold text-font-clear">{title}</span>
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

{#snippet switchBtn(on: boolean, onToggle: () => void)}
    <button type="button" role="switch" aria-checked={on} aria-label="toggle" onclick={onToggle}
        class="relative h-5.5 w-10 rounded-full transition-colors duration-150 shrink-0 cursor-pointer
               {on ? 'bg-primary' : 'bg-card-edge-2'}">
        <span class="absolute top-0.5 left-0.5 h-4.5 w-4.5 rounded-full bg-font-clear shadow-sm transition-transform duration-150
                     {on ? 'translate-x-4.5' : ''}"></span>
    </button>
{/snippet}

{#snippet primaryBtn(label: string)}
    <button type="submit"
        class="inline-flex items-center gap-2 rounded-lg px-4 py-2 text-sm font-semibold
               bg-gradient-primary text-black hover:brightness-125 transition-all duration-150">
        {label}
    </button>
{/snippet}

<h1 class="font-fancy text-2xl font-semibold text-font-clear mx-auto w-full max-w-6xl px-4 mt-2 mb-4">Preferences</h1>

<div class="mx-auto w-full max-w-6xl px-4 mb-12 columns-1 lg:columns-2 gap-4 [&>*]:mb-4 [&>*]:break-inside-avoid">

    <!-- Profile -->
    <div class="rounded-2xl border border-card-edge bg-card shadow-md shadow-darken overflow-hidden">
        <div class="h-[3px] bg-gradient-primary"></div>
        <form method="POST" action="?/profile" use:enhance class="p-6 flex flex-col gap-5">
            {@render head(User, 'Profile')}
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
            <div>{@render primaryBtn('Save changes')}</div>
        </form>
    </div>

    <!-- Security -->
    <div class="rounded-2xl border border-card-edge bg-card shadow-md shadow-darken overflow-hidden">
        <div class="h-[3px] bg-gradient-primary"></div>
        <div class="p-6 flex flex-col gap-5">
            {@render head(KeyRound, 'Security')}
            {#if form?.passwordOk}{@render banner(true, 'Password changed.')}{/if}
            {#if form?.passwordError}{@render banner(false, form.passwordError)}{/if}
            {#if form?.signedOut}{@render banner(true, 'Signed out of all other devices.')}{/if}

            <form method="POST" action="?/password" use:enhance class="flex flex-col gap-4">
                {@render field('Current password', 'current', 'password', '', '')}
                <div class="flex flex-col sm:flex-row gap-4">
                    {@render field('New password', 'new', 'password', '', '')}
                    {@render field('Confirm', 'confirm', 'password', '', '')}
                </div>
                <div>{@render primaryBtn('Change password')}</div>
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

    <!-- Town Crier (notifications) -->
    <div class="rounded-2xl border border-card-edge bg-card shadow-md shadow-darken overflow-hidden">
        <div class="h-[3px] bg-gradient-primary"></div>
        <form method="POST" action="?/notifications" use:enhance class="p-6 flex flex-col gap-5">
            <div class="flex flex-col gap-1">
                {@render head(Bell, 'Town Crier')}
                <p class="text-sm text-font-dimer">Choose when a messenger should bring word to your inbox.</p>
            </div>
            {#if form?.notifyOk}{@render banner(true, 'Notification preferences saved.')}{/if}

            <div class="flex flex-col divide-y divide-card-edge">
                {#each notifyMeta as n}
                    <div class="flex items-center gap-4 py-3 first:pt-0 last:pb-0">
                        <input type="hidden" name="notify_{n.key}" value={notif[n.key] ? 'on' : ''} />
                        <div class="flex flex-col leading-tight flex-1 min-w-0">
                            <span class="text-sm text-font-clear">{n.label}</span>
                            <span class="text-xs text-font-dimest mt-0.5">{n.desc}</span>
                        </div>
                        {@render switchBtn(!!notif[n.key], () => (notif[n.key] = !notif[n.key]))}
                    </div>
                {/each}
            </div>
            <div>{@render primaryBtn('Save preferences')}</div>
        </form>
    </div>

    <!-- Diplomacy (denounce / forgive) -->
    <div class="rounded-2xl border border-card-edge bg-card shadow-md shadow-darken">
        <div class="h-[3px] bg-gradient-primary rounded-t-[15px]"></div>
        <div class="p-6 flex flex-col gap-5">
            <div class="flex flex-col gap-1">
                {@render head(Megaphone, 'Diplomacy')}
                <p class="text-sm text-font-dimer">
                    Publicly denounce a rival. They'll wear a frowny face on their profile — and hear about it.
                </p>
            </div>
            {#if form?.denounceOk}{@render banner(true, 'Denounced. The realm has been notified.')}{/if}
            {#if form?.forgiveOk}{@render banner(true, 'Forgiven. How magnanimous.')}{/if}
            {#if form?.diploError}{@render banner(false, form.diploError)}{/if}

            <form method="POST" action="?/denounce" use:enhance={() => { denounceOpen = false; }} class="flex items-end gap-2">
                <input type="hidden" name="player_id" value={denounceTarget} />
                <div class="flex flex-col gap-1.5 flex-1 min-w-0">
                    <span class="text-xs font-fancy tracking-wide uppercase text-font-dimest">Denounce a player</span>

                    <!-- Custom dropdown -->
                    <div class="relative">
                        <button type="button" onclick={() => (denounceOpen = !denounceOpen)}
                            class="w-full flex items-center justify-between gap-2 rounded-lg border bg-card-2 px-3 py-2 text-sm
                                   transition-colors duration-150 cursor-pointer
                                   {denounceOpen ? 'border-primary/40' : 'border-card-edge hover:border-card-edge-2'}">
                            <span class="truncate {selectedName ? 'text-font-clear' : 'text-font-dimest'}">
                                {selectedName || 'Choose a rival…'}
                            </span>
                            <ChevronDown class="h-4 w-4 text-font-dimer shrink-0 transition-transform duration-150 {denounceOpen ? 'rotate-180' : ''}" />
                        </button>

                        {#if denounceOpen}
                            <button type="button" class="fixed inset-0 z-10 cursor-default" tabindex="-1"
                                aria-hidden="true" onclick={() => (denounceOpen = false)}></button>
                            <div class="absolute z-20 mt-1 w-full rounded-lg border border-card-edge bg-card shadow-lg shadow-darken overflow-hidden">
                                <div class="flex items-center gap-2 px-3 py-2 border-b border-card-edge">
                                    <Search class="h-3.5 w-3.5 text-font-dimest shrink-0" />
                                    <!-- svelte-ignore a11y_autofocus -->
                                    <input bind:value={denounceSearch} autofocus placeholder="Search players…"
                                        class="w-full bg-transparent text-sm text-font-clear outline-none placeholder:text-font-dimest" />
                                </div>
                                <div class="max-h-60 overflow-y-auto py-1">
                                    {#each filteredPlayers as p}
                                        <button type="button"
                                            onclick={() => { denounceTarget = String(p.id); denounceOpen = false; denounceSearch = ''; }}
                                            class="w-full text-left px-3 py-1.5 text-sm transition-colors duration-100 cursor-pointer
                                                   {String(p.id) === denounceTarget
                                                     ? 'bg-primary/15 text-primary'
                                                     : 'text-font-dim hover:bg-select hover:text-font-clear'}">
                                            {p.name}
                                        </button>
                                    {:else}
                                        <div class="px-3 py-2 text-sm text-font-dimest italic">No players found.</div>
                                    {/each}
                                </div>
                            </div>
                        {/if}
                    </div>
                </div>

                <button type="submit" disabled={!denounceTarget}
                    class="shrink-0 flex items-center gap-1.5 rounded-lg px-4 py-2 text-sm font-semibold transition-all duration-150
                           {denounceTarget
                             ? 'bg-font-bad/15 text-font-bad border border-font-bad/30 hover:bg-font-bad/25 cursor-pointer'
                             : 'border border-card-edge text-font-dimest cursor-not-allowed opacity-50'}">
                    <Angry class="h-4 w-4" strokeWidth={1.75} /> Denounce
                </button>
            </form>

            {#if denounced.length > 0}
                <div class="flex flex-col gap-2">
                    <span class="text-xs font-fancy tracking-wide uppercase text-font-dimest">Currently denounced</span>
                    {#each denounced as d}
                        <div class="flex items-center gap-3 rounded-xl border border-card-edge bg-card-2 px-4 py-2.5">
                            <Angry class="h-4.5 w-4.5 text-font-bad shrink-0" strokeWidth={1.75} />
                            <a href="/profile/{d.id}" class="text-sm text-font-clear hover:text-primary transition-colors flex-1 min-w-0 truncate">{d.name}</a>
                            <form method="POST" action="?/forgive" use:enhance class="shrink-0">
                                <input type="hidden" name="player_id" value={d.id} />
                                <button type="submit"
                                    class="flex items-center gap-1.5 rounded-lg border border-card-edge px-3 py-1.5 text-xs text-font-dimer
                                           hover:border-font-good/40 hover:text-font-good transition-colors duration-150 cursor-pointer">
                                    <HeartHandshake class="h-3.5 w-3.5" strokeWidth={1.5} /> Forgive
                                </button>
                            </form>
                        </div>
                    {/each}
                </div>
            {:else}
                <p class="text-sm text-font-dimest italic">You hold no active grudges. Suspicious.</p>
            {/if}
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
                                <span class="text-sm text-font-clear font-medium truncate">{acc.persona ?? 'Steam account'}</span>
                                <a href="https://steamcommunity.com/profiles/{acc.steam_id}"
                                   target="_blank" rel="noopener"
                                   class="text-xs text-font-dimest hover:text-primary transition-colors duration-150 flex items-center gap-1 w-fit">
                                    {acc.steam_id}<ExternalLink class="h-3 w-3" />
                                </a>
                            </div>
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

    <!-- Danger Zone -->
    <div class="rounded-2xl border border-font-bad/30 bg-card shadow-md shadow-darken overflow-hidden">
        <div class="h-[3px] bg-font-bad/70"></div>
        <div class="p-6 flex flex-col gap-4">
            <div class="flex items-center gap-2">
                <Flame class="h-5 w-5 text-font-bad" strokeWidth={1.75} />
                <span class="font-fancy text-lg font-semibold text-font-clear">Danger Zone</span>
            </div>

            {#if form?.peaceOk}{@render banner(true, `Sued for peace — withdrew ${form.peaceCount} denouncement${form.peaceCount === 1 ? '' : 's'}.`)}{/if}
            {#if form?.steamSevered}{@render banner(true, `Severed ${form.steamCount} Steam link${form.steamCount === 1 ? '' : 's'}.`)}{/if}
            {#if form?.activeToggled}{@render banner(true, form.active ? 'Welcome back — you are ranked once more.' : 'You have retired from the ladder.')}{/if}

            <!-- Sue for peace: clears all your denouncements -->
            <div class="flex items-center gap-4">
                <div class="flex flex-col leading-tight flex-1 min-w-0">
                    <span class="text-sm text-font-clear">Sue for peace with all rivals</span>
                    <span class="text-xs text-font-dimest mt-0.5">Withdraw every denouncement you've issued. Grudges may, of course, be rekindled.</span>
                </div>
                <form method="POST" action="?/sue_for_peace" class="shrink-0"
                    use:enhance={({ cancel }) => { if (!confirm('Withdraw all your denouncements?')) cancel(); }}>
                    <button type="submit"
                        class="flex items-center gap-1.5 rounded-lg border border-card-edge px-3 py-1.5 text-sm text-font-dimer
                               hover:border-primary/40 hover:text-primary transition-colors duration-150 cursor-pointer">
                        <Scroll class="h-3.5 w-3.5" strokeWidth={1.5} /> Sue for peace
                    </button>
                </form>
            </div>

            <!-- Sever Steam ties: unlink all steam accounts -->
            <div class="flex items-center gap-4">
                <div class="flex flex-col leading-tight flex-1 min-w-0">
                    <span class="text-sm text-font-clear">Sever your Steam allegiances</span>
                    <span class="text-xs text-font-dimest mt-0.5">Unlink every Steam account. New uploads will no longer recognise you automatically.</span>
                </div>
                <form method="POST" action="?/sever_steam" class="shrink-0"
                    use:enhance={({ cancel }) => { if (!confirm('Unlink all your Steam accounts?')) cancel(); }}>
                    <button type="submit"
                        class="flex items-center gap-1.5 rounded-lg border border-card-edge px-3 py-1.5 text-sm text-font-dimer
                               hover:border-font-bad/40 hover:text-font-bad transition-colors duration-150 cursor-pointer">
                        <Unlink class="h-3.5 w-3.5" strokeWidth={1.5} /> Sever ties
                    </button>
                </form>
            </div>

            <!-- Abdicate / Reclaim: retire from or return to the ladder -->
            <div class="flex items-center gap-4">
                <div class="flex flex-col leading-tight flex-1 min-w-0">
                    <span class="text-sm text-font-bad font-medium">
                        {data.profile?.active ? 'Abdicate the throne' : 'Reclaim the throne'}
                    </span>
                    <span class="text-xs text-font-dimest mt-0.5">
                        {data.profile?.active
                            ? 'Retire from competition — you’ll be hidden from leaderboards and rankings until you return.'
                            : 'You are currently retired and hidden from the leaderboards. Return to the rankings.'}
                    </span>
                </div>
                <form method="POST" action="?/toggle_active" class="shrink-0"
                    use:enhance={({ cancel }) => {
                        const msg = data.profile?.active ? 'Retire from the ladder?' : 'Return to the ladder?';
                        if (!confirm(msg)) cancel();
                    }}>
                    <button type="submit"
                        class="flex items-center gap-1.5 rounded-lg px-3 py-1.5 text-sm font-semibold transition-colors duration-150 cursor-pointer
                               {data.profile?.active
                                 ? 'bg-font-bad/15 text-font-bad border border-font-bad/30 hover:bg-font-bad/25'
                                 : 'bg-gradient-primary text-black hover:brightness-125'}">
                        <Crown class="h-3.5 w-3.5" strokeWidth={1.75} />
                        {data.profile?.active ? 'Abdicate' : 'Reclaim'}
                    </button>
                </form>
            </div>
        </div>
    </div>
</div>
