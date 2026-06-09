<script lang="ts">
	import './layout.css';
	import favicon from '$lib/assets/favicon-64.png';
	import favicon_apple from '$lib/assets/apple-touch-icon.png'
	import { BookPlus, BookUp, BookUp2, ChevronDown, CircleUser, LogIn, LogOut, NotebookPen, Settings, User } from '@lucide/svelte'

	const links = [
		{ href: '/leaderboard', label: 'Leaderboard' },
		{ href: '/matches', label: 'Matches' },
		{ href: '/stats', label: 'Stats' },
		{ href: '/achievements', label: 'Achievements' },
	];

	let { children, data } = $props();
	const user = $derived(data.user);
	let dropdownOpen = $state(false);
</script>

<svelte:head>
	<title>civ6.ch • Leaderboard & Statistics</title>
	<link rel="icon" type="image/png" href={favicon} />
	<link rel="apple-touch-icon" href={favicon_apple} />
</svelte:head>

<svelte:window onclick={() => dropdownOpen = false} />

<div class="flex flex-col min-h-dvh">
	<header class="w-full relative h-16 flex justify-between px-4 md:px-12 items-center bg-background/60 z-50 backdrop-blur-sm">
		<a href="/" class="text-3xl font-bold z-10 hover:text-font-clear transition-colors duration-120 ease-out font-fancy" style="text-shadow: 1px 1px 0px var(--color-primary-shadow);">civ6.ch</a>
		<nav class="absolute left-0 hidden md:flex w-full gap-6 justify-center p-1 z-9">
			{#each links as link}
				<a href={link.href} class="relative text-m font-semibold tracking-wider hover:text-font-clear transition-colors duration-150 ease-in-out group font-fancy">
					{link.label}
					<span class="absolute bottom-0 left-0 w-full h-0.5 rounded-full bg-font-clear scale-x-0 group-hover:scale-x-100 transition-transform duration-150 ease-in-out"></span>
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
	</header>

	<!-- Mobile nav strip -->
	<nav class="md:hidden flex border-b border-card-edge">
		{#each links as link}
			<a href={link.href}
			   class="flex-1 text-center font-fancy text-xs py-3 text-font-dimest hover:text-font-clear transition-colors duration-150">
				{link.label}
			</a>
		{/each}
	</nav>

	<main class="flex-1 flex flex-col mt-4 md:mt-4">
		{@render children()}
	</main>
</div>