<script lang="ts">
    import { enhance } from '$app/forms';
    import { page } from '$app/stores';
    import {
        Unlink, ExternalLink, ShieldCheck, Link, User, KeyRound, LogOut
    } from '@lucide/svelte';
    import type { PageData } from './$types';

    let { data, form }: { data: PageData; form: any } = $props();
    const { steamAccounts } = $derived(data);

    const steamStatus = $derived($page.url.searchParams.get('steam'));

    let name = $state(data.profile?.name ?? '');
    let email = $state(data.profile?.email ?? '');

    function fmtDate(d: string | Date) {
        return new Date(d).toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' });
    }
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

<div class="mx-3 md:mx-12 mb-12 flex flex-col gap-4 max-w-3xl">

    <div class="flex flex-col gap-1 mt-2">
        <h1 class="font-fancy text-2xl font-semibold text-font-clear">Preferences</h1>
        <p class="text-sm text-font-dimer">Manage your account and how you appear on civ6.ch.</p>
    </div>

    <!-- Profile -->
    <div class="rounded-2xl border border-card-edge bg-card shadow-md shadow-darken overflow-hidden">
        <div class="h-[3px] bg-gradient-primary"></div>
        <form method="POST" action="?/profile" use:enhance class="p-6 flex flex-col gap-5">
            {@render sectionHead(User, 'Profile', 'Your public display name and contact email.')}

            {#if form?.profileOk}{@render banner(true, 'Profile updated.')}{/if}
            {#if form?.profileError}{@render banner(false, form.profileError)}{/if}

            <div class="flex flex-col gap-4">
                <label class="flex flex-col gap-1.5">
                    <span class="text-xs font-fancy tracking-wide uppercase text-font-dimest">Display name</span>
                    <input name="name" bind:value={name} maxlength="40" autocomplete="off"
                        class="rounded-lg border border-card-edge bg-card-2 px-3 py-2 text-sm text-font-clear
                               outline-none focus:border-primary/40 placeholder:text-font-dimest" />
                    <span class="text-xs text-font-dimest">Shown on the leaderboard, matches and your profile.</span>
                </label>
                <label class="flex flex-col gap-1.5">
                    <span class="text-xs font-fancy tracking-wide uppercase text-font-dimest">Email</span>
                    <input name="email" type="email" bind:value={email} autocomplete="off"
                        placeholder="you@example.com"
                        class="rounded-lg border border-card-edge bg-card-2 px-3 py-2 text-sm text-font-clear
                               outline-none focus:border-primary/40 placeholder:text-font-dimest" />
                    <span class="text-xs text-font-dimest">Used to sign in and recover your account. Never shown publicly.</span>
                </label>
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
                <label class="flex flex-col gap-1.5">
                    <span class="text-xs font-fancy tracking-wide uppercase text-font-dimest">Current password</span>
                    <input name="current" type="password" autocomplete="current-password"
                        class="rounded-lg border border-card-edge bg-card-2 px-3 py-2 text-sm text-font-clear
                               outline-none focus:border-primary/40" />
                </label>
                <div class="flex flex-col sm:flex-row gap-4">
                    <label class="flex flex-col gap-1.5 flex-1">
                        <span class="text-xs font-fancy tracking-wide uppercase text-font-dimest">New password</span>
                        <input name="new" type="password" autocomplete="new-password"
                            class="rounded-lg border border-card-edge bg-card-2 px-3 py-2 text-sm text-font-clear
                                   outline-none focus:border-primary/40" />
                    </label>
                    <label class="flex flex-col gap-1.5 flex-1">
                        <span class="text-xs font-fancy tracking-wide uppercase text-font-dimest">Confirm new password</span>
                        <input name="confirm" type="password" autocomplete="new-password"
                            class="rounded-lg border border-card-edge bg-card-2 px-3 py-2 text-sm text-font-clear
                                   outline-none focus:border-primary/40" />
                    </label>
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
</div>
