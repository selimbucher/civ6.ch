<script lang="ts">
	import './layout.css';
	import favicon from '$lib/assets/favicon-64.png';
	import favicon_apple from '$lib/assets/apple-touch-icon.png'
	import { BookPlus, BookUp, BookUp2, ChevronDown, CircleUser, LogIn, LogOut, NotebookPen, Settings, User } from '@lucide/svelte'
	import { page } from '$app/state';
	import { onMount } from 'svelte';
	import { mottos, daily } from '$lib/flavor';

	// Centralised, per-route document title. Kept in the layout (rather than each
	// page's <svelte:head>) so it works uniformly — Svelte's title de-duplication
	// is finicky when a page also defines top-level snippets.
	const SUFFIX = 'civ6.ch';
	const staticTitles: Record<string, string> = {
		'/': 'civ6.ch • Swiss Civilization VI League',
		'/leaderboard': 'Leaderboard',
		'/matches': 'Matches',
		'/matches/add': 'Upload a Game',
		'/matches/confirm/[id]': 'Confirm Game',
		'/stats': 'Hall of Records',
		'/achievements': 'Achievements',
		'/login': 'Login',
		'/settings': 'Preferences'
	};
	const pageTitle = $derived.by(() => {
		if (page.status >= 400) return `${page.status} • ${SUFFIX}`;
		const id = page.route.id ?? '/';
		if (id === '/profile/[id]') {
			const name = (page.data as any)?.player?.name;
			return name ? `${name} • ${SUFFIX}` : `Player • ${SUFFIX}`;
		}
		if (id === '/matches/view/[id]') {
			const v = (page.data as any)?.game?.victory_type;
			const label = v && v !== 'Unknown' ? `${v} Victory` : 'Match';
			return `${label} • ${SUFFIX}`;
		}
		const t = staticTitles[id];
		if (!t) return `civ6.ch • Leaderboard & Statistics`;
		return t === staticTitles['/'] ? t : `${t} • ${SUFFIX}`;
	});

	const links = [
		{ href: '/leaderboard', label: 'Leaderboard' },
		{ href: '/matches', label: 'Matches' },
		{ href: '/stats', label: 'Stats' },
		{ href: '/achievements', label: 'Achievements' },
	];

	let { children, data } = $props();
	const user = $derived(data.user);
	let dropdownOpen = $state(false);

	const motto = daily(mottos);
	// The home page is a full-screen splash; a footer would overlay the hero.
	const showFooter = $derived(page.route.id !== '/');
	const isActive = (href: string) => page.url.pathname === href || page.url.pathname.startsWith(href + '/');

	// Social embeds: match pages share their rendered map instead of the
	// generic banner, so a pasted link shows the actual game.
	const ogUrl = $derived(`https://civ6.ch${page.url.pathname}`);
	const ogImage = $derived.by(() => {
		if (page.route.id === '/matches/view/[id]' && (page.data as any)?.game?.has_map) {
			return `https://civ6.ch/files/maps/${(page.data as any).game.id}`;
		}
		return 'https://civ6.ch/og.png';
	});

	onMount(() => {
		console.log(
			'%c☢  Our words are backed with NUCLEAR WEAPONS.',
			'color:#e8c34a; font-size:14px; font-weight:bold; font-family:Georgia,serif;'
		);
		console.log(
			'%c— Gandhi, reviewing your DevTools usage. The league\'s data is public; scrape responsibly.',
			'color:#888; font-style:italic;'
		);
	});
</script>

<svelte:head>
	<title>{pageTitle}</title>
	<meta name="description" content="The Swiss Civilization VI League — live ratings, match history, statistics and dubious achievements." />
	<meta name="theme-color" content="#0a0a0a" />
	<link rel="icon" type="image/png" href={favicon} />
	<link rel="apple-touch-icon" href={favicon_apple} />

	<!-- Open Graph / social sharing -->
	<meta property="og:type" content="website" />
	<meta property="og:site_name" content="civ6.ch" />
	<meta property="og:title" content={pageTitle} />
	<meta property="og:description" content="Live ratings, match history, statistics and dubious achievements." />
	<meta property="og:url" content={ogUrl} />
	<meta property="og:image" content={ogImage} />
	{#if ogImage.endsWith('/og.png')}
		<meta property="og:image:width" content="1200" />
		<meta property="og:image:height" content="630" />
	{/if}
	<meta name="twitter:card" content="summary_large_image" />
	<meta name="twitter:title" content={pageTitle} />
	<meta name="twitter:description" content="Live ratings, match history, statistics and dubious achievements." />
	<meta name="twitter:image" content={ogImage} />
</svelte:head>

<svelte:window onclick={() => dropdownOpen = false} />

<div class="flex flex-col min-h-dvh">
	<header class="w-full h-16 bg-background/60 z-50 backdrop-blur-sm">
	<div class="relative mx-auto h-full w-full max-w-page flex justify-between px-4 md:px-12 items-center">
		<a href="/" class="text-3xl font-bold z-10 hover:text-font-clear transition-colors duration-120 ease-out font-fancy" style="text-shadow: 1px 1px 0px var(--color-primary-shadow);">civ6.ch</a>
		<nav class="absolute left-0 hidden md:flex w-full gap-6 justify-center p-1 z-9">
			{#each links as link}
				{@const active = isActive(link.href)}
				<a href={link.href} aria-current={active ? 'page' : undefined}
				   class="relative text-m font-semibold tracking-wider hover:text-font-clear transition-colors duration-150 ease-in-out group font-fancy {active ? 'text-font-clear' : ''}">
					{link.label}
					<span class="absolute bottom-0 left-0 w-full h-0.5 rounded-full transition-transform duration-150 ease-in-out {active ? 'bg-primary scale-x-100' : 'bg-font-clear scale-x-0 group-hover:scale-x-100'}"></span>
				</a>
			{/each}
		</nav>
		<div class="flex gap-1 items-center z-10">
			{#if user}
				<div class="relative ml-2">
					<button
						onclick={(e) => { e.stopPropagation(); dropdownOpen = !dropdownOpen; }}
						class="flex items-center gap-1.5 text-lg font-fancy hover:text-font-clear font-semibold transition-colors duration-150 ease-out"
					>
						{user.name ?? user.username}
						<ChevronDown class="h-4 w-4 transition-transform duration-150 {dropdownOpen ? 'rotate-180' : ''}" />
					</button>

					{#if dropdownOpen}
						<div class="absolute right-0 top-full mt-2 w-44 rounded-xl border border-card-edge bg-card shadow-md shadow-darken overflow-hidden">
							<a
								href="/profile/{user.id}"
								onclick={() => dropdownOpen = false}
								class="flex items-center gap-2.5 px-4 py-2.5 text-sm text-font-dim hover:bg-select hover:text-font-clear transition-colors duration-150"
							>
								<User class="h-4 w-4" />
								Profile
							</a>
							<a
								href="/settings"
								type="submit"
								class="flex w-full items-center gap-2.5 px-4 py-2.5 text-sm text-font-dim hover:bg-select hover:text-font-clear transition-colors duration-150"
							>
								<Settings class="h-4 w-4" />
								Preferences
							</a>
							<a
								href="/logout"
								type="submit"
								class="flex w-full items-center gap-2.5 px-4 py-2.5 text-sm text-font-dim hover:bg-select hover:text-font-clear transition-colors duration-150"
							>
								<LogOut class="h-4 w-4" />
								Logout
							</a>
							<a
								href="/matches/add"
								type="submit"
								class="flex w-full items-center gap-2.5 px-4 py-2.5 text-sm text-font-dim hover:bg-select hover:text-font-clear transition-colors duration-150"
							>
								<NotebookPen class="h-4 w-4" />
								Upload Game
							</a>
						</div>
					{/if}
				</div>
			{:else}
				<a href="/login" class="tracking-wide flex items-center justify-center ml-2 px-4 py-1 rounded-lg font-fancy hover:text-font-clear font-semibold transition-colors duration-150 ease-out">
					<CircleUser strokeWidth={2} class="inline-block mr-1.5 h-5 w-5 mb-0.5"/><span> Login</span>
				</a>
			{/if}
		</div>
	</div>
	</header>

	<!-- Mobile nav strip -->
	<nav class="md:hidden flex border-b border-card-edge">
		{#each links as link}
			{@const active = isActive(link.href)}
			<a href={link.href} aria-current={active ? 'page' : undefined}
			   class="flex-1 text-center font-fancy text-xs py-3 transition-colors duration-150 {active ? 'text-primary border-b-2 border-primary' : 'text-font-dimest hover:text-font-clear'}">
				{link.label}
			</a>
		{/each}
	</nav>

	<main class="flex-1 flex flex-col mt-4 md:mt-4 w-full max-w-page mx-auto">
		{@render children()}
	</main>

	{#if showFooter}
		<footer class="mt-20 border-t border-card-edge bg-background/60">
			<div class="mx-auto w-full max-w-page px-4 md:px-12 py-10">
				<div class="flex flex-col gap-8 sm:flex-row sm:justify-between">
					<div class="max-w-xs">
						<a href="/" class="font-fancy text-2xl font-bold hover:text-font-clear transition-colors duration-150" style="text-shadow: 1px 1px 0px var(--color-primary-shadow);">civ6.ch</a>
						<p class="mt-2 text-sm leading-relaxed text-font-dimest">
							The Swiss Civilization VI League. Proudly neutral since 4000 BC.
						</p>
					</div>
					<nav class="flex flex-col gap-2">
						<span class="font-fancy text-xs uppercase tracking-widest text-font-dimest">Navigate</span>
						{#each links as link}
							<a href={link.href} class="text-sm text-font-dim hover:text-font-clear transition-colors duration-150">{link.label}</a>
						{/each}
					</nav>
					<div class="flex flex-col gap-2 sm:text-right">
						<span class="font-fancy text-xs uppercase tracking-widest text-font-dimest">State of the League</span>
						{#if data.league}
							<span class="text-sm text-font-dim">{data.league.games.toLocaleString('de-CH')} games recorded</span>
							<span class="text-sm text-font-dim">{data.league.turns.toLocaleString('de-CH')} turns endured</span>
						{/if}
						<span class="text-sm text-font-dim">0 wars ongoing&nbsp;<span class="text-font-dimest">(unverified)</span></span>
					</div>
				</div>
				<div class="mt-10 flex flex-col gap-1.5 border-t border-card-edge pt-5 sm:flex-row sm:items-baseline sm:justify-between">
					<span class="text-xs text-font-dimest">© {new Date().getFullYear()} civ6.ch — all rights reserved, all wonders reserved.</span>
					<span class="text-xs italic text-font-dimest">{motto}</span>
				</div>
			</div>
		</footer>
	{/if}
</div>