<script lang="ts">
	import bg from '$lib/assets/hero.webp';
	import { Swords } from '@lucide/svelte';
</script>

<div class="flex absolute items-center justify-center overflow-hidden h-screen w-full top-0">
	<div
		class="absolute inset-0 hero-zoom"
		style="background-image: url('{bg}'); background-size: cover; background-position: 60% 20%;"
	></div>

	<!-- overlays -->
	<div class="absolute inset-0 bg-gradient-to-b from-black/10 via-black/40 to-background"></div>
	<div class="absolute inset-0" style="background: radial-gradient(ellipse at center, transparent 30%, rgba(0,0,0,0.6) 100%)"></div>

	<!-- content -->
	<div class="relative z-10 text-center flex flex-col items-center gap-5 px-8 -mt-16">

		<h1
			class="select-none font-fancy font-bold text-primary"
			style="font-size: clamp(4rem, 11vw, 8.5rem); line-height: 1; text-shadow: 3px 3px 0px var(--color-primary-shadow), 0 0 100px hsl(49.88deg 100% 40% / 30%);     filter: drop-shadow(0.1rem 0.25rem 0.5rem #00000060);"
		>
			civ6.ch
		</h1>

		<p class="font-fancy tracking-[0.45em] uppercase text-black text-sm" style="text-shadow: 0.02rem 0.05rem 0.2rem #000000;">
			Swiss Civilization VI League
		</p>

		<div class="flex items-center gap-3 mt-1 opacity-40">
			<div class="h-px w-28 bg-gradient-to-r from-transparent to-primary"></div>
			<div class="h-1.5 w-1.5 rotate-45 bg-primary"></div>
			<div class="h-px w-28 bg-gradient-to-l from-transparent to-primary"></div>
		</div>

		<div class="flex gap-3 mt-6">
			<a
				href="/leaderboard"
				class="font-fancy border-none tracking-widest text-sm px-8 py-3 rounded-lg bg-gradient-primary text-black font-bold hover:brightness-110 transition-all duration-250"
				style="box-shadow: 0 4px 28px hsl(49.88deg 73% 50% / 30%) filter: drop-shadow(0.1rem 0.25rem 0.2rem #000000fa);"
			>
				Leaderboard
			</a>
			<a
				href="/matches"
				class="font-fancy tracking-widest text-sm px-8 py-3 rounded-lg bg-background/30 text-font-dim hover:text-font-clear hover:bg-background/70 backdrop-blur-sm transition-colors duration-250"
			>
				Matches
			</a>
		</div>
	</div>
</div>

<style>

	.hero-zoom {
		animation: heroZoom 1000ms ease forwards;
        transform-origin: 45% 50%;
	}

	@keyframes heroZoom {
		from { transform: scale(1); }
		to   { transform: scale(1.07); }
	}
</style>
