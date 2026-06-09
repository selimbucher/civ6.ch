<script lang="ts">
    import universalv from '$lib/assets/icons/vcondition/universal.png';
    import dominationv from '$lib/assets/icons/vcondition/domination.png';
    import religiousv from '$lib/assets/icons/vcondition/religious.png';
    import diplomaticv from '$lib/assets/icons/vcondition/diplomatic.png';
    import sciencev from '$lib/assets/icons/vcondition/science.png';
    import culturev from '$lib/assets/icons/vcondition/culture.png';
    import scorev from '$lib/assets/icons/vcondition/score.png';
    import capitulationv from '$lib/assets/icons/vcondition/capitulation.png';
    import { Trophy, Flame, Star, Clock, ChevronRight, Medal } from '@lucide/svelte';
    import RatingChart from '$lib/RatingChart.svelte';

    const leaderAssets = import.meta.glob<{ default: string }>(
        '$lib/assets/icons/leaders/*.webp',
        { eager: true }
    );

    function leaderPortrait(leader: string | null): string | null {
        if (!leader) return null;
        const slug = leader.trim().replace(/\s+/g, '_');
        const norm = (s: string) =>
            s.normalize('NFD').replace(/[\u0300-\u036f]/g, '').toLowerCase();
        const keys = Object.keys(leaderAssets);
        const key =
            keys.find(k => k.includes(`/${slug}_(Civ6).`)) ??
            keys.find(k => k.toLowerCase().includes(`/${slug.toLowerCase()}_(civ6).`)) ??
            keys.find(k => norm(k).includes(`/${norm(slug)}_(civ6).`));
        return key ? leaderAssets[key].default : null;
    }

    const victoryIcons: Record<string, string> = {
        'Domination':  dominationv,
        'Religious':   religiousv,
        'Diplomatic':  diplomaticv,
        'Science':     sciencev,
        'Culture':     culturev,
        'Score':       scorev,
        'Capitulation': capitulationv,
    };

    const victoryColors: Record<string, string> = {
        'Domination':  'var(--color-army-1)',
        'Religious':   'var(--color-faith-1)',
        'Diplomatic':  'var(--color-diplo-1)',
        'Science':     'var(--color-science-1)',
        'Culture':     'var(--color-culture-1)',
        'Score':       'var(--color-score-1)',
        'Capitulation':'var(--color-font-dimest)',
    };

    const categoryLabel: Record<string, string> = {
        overall: 'Overall', ffa: 'FFA', '1v1': '1v1', teams: 'Teams',
    };

    const difficultyLabel: Record<number, string> = {
        [-1]: 'Secret', 0: 'Common', 1: 'Uncommon', 2: 'Rare', 3: 'Epic',
    };
    const difficultyColor: Record<number, string> = {
        [-1]: 'text-secondary',
        0: 'text-font-dimer',
        1: 'text-font-good',
        2: 'text-primary',
        3: 'text-font-bad',
    };

    let { data } = $props();
    const pid = $derived(data.player.id);

    function getRating(category: string) {
        return data.ratings.find((r: any) => r.category === category);
    }
    function getStats(category: string) {
        return data.stats.find((s: any) => s.category === category);
    }
    function getRank(category: string): number | null {
        return data.ranks.find((r: any) => r.category === category)?.rank ?? null;
    }

    const overallRating = $derived(getRating('overall'));
    const overallStats  = $derived(getStats('overall'));
    const overallRank   = $derived(getRank('overall'));

    function winRate(wins: number, games: number) {
        if (!games) return 0;
        return Math.round((wins / games) * 100);
    }

    function formatDate(date: string) {
        const d = new Date(date);
        const diff = Date.now() - d.getTime();
        const days = Math.floor(diff / 86400000);
        if (days === 0) return 'today';
        if (days === 1) return 'yesterday';
        if (days < 7) return `${days}d ago`;
        if (days < 30) return `${Math.floor(days / 7)}w ago`;
        if (days < 365) return `${Math.floor(days / 30)}mo ago`;
        return `${Math.floor(days / 365)}y ago`;
    }

    function formatDelta(pre: number, post: number) {
        const delta = Math.round(post - pre);
        if (delta > 0) return { text: `+${delta}`, cls: 'text-font-good' };
        if (delta < 0) return { text: `${delta}`, cls: 'text-font-bad' };
        return { text: '+0', cls: 'text-font-dimer' };
    }

    // ── Achievements ─────────────────────────────────────────────────────────
    const earnedIds = $derived.by(() => {
        const bits = BigInt(data.player.achievement_bitstring ?? 0);
        return new Set(
            data.achievements
                .filter((_: any, i: number) => {
                    const a = data.achievements[i];
                    return (bits >> BigInt(a.id - 1)) & 1n;
                })
                .map((a: any) => a.id)
        );
    });

    const earnedCount = $derived(earnedIds.size);
    const totalCount  = $derived(data.achievements.length);

    // ── Categories to display ─────────────────────────────────────────────────
    const cats = ['overall', '1v1', 'ffa', 'teams'] as const;

    const victoryMaxTotal = $derived(
        data.victoryStats.length > 0
            ? Math.max(...data.victoryStats.map((v: any) => Number(v.total)))
            : 1
    );
</script>

<div class="mx-3 md:mx-12 mb-12 flex flex-col gap-4">

    <!-- ── Hero ────────────────────────────────────────────────────────────── -->
    <div class="relative rounded-2xl border border-card-edge bg-card shadow-md shadow-darken overflow-hidden">
        <!-- top accent bar -->
        <div class="absolute top-0 left-0 w-full h-[3px] bg-gradient-primary"></div>

        <div class="flex flex-col md:flex-row items-center md:items-start gap-5 px-5 md:px-10 py-6 md:py-8">
            <!-- Avatar placeholder -->
            <div class="shrink-0 relative">
                <div class="h-20 w-20 rounded-full border-2 border-card-edge bg-card-2 flex items-center justify-center text-font-dimest">
                    <span class="font-fancy text-3xl font-bold text-primary select-none">
                        {data.player.name.charAt(0).toUpperCase()}
                    </span>
                </div>
                {#if overallRank === 1}
                    <div class="absolute bottom-0 right-0 text-primary">
                        <Trophy class="h-6 w-6" strokeWidth={1.8} />
                    </div>
                {/if}
            </div>

            <!-- Name + meta -->
            <div class="flex-1 min-w-0">
                <h1 class="font-fancy text-4xl font-bold text-font-clear tracking-wide truncate"
                    style="text-shadow: 1px 1px 0px var(--color-primary-shadow);">
                    {data.player.name}
                </h1>
                <div class="mt-2 flex flex-wrap gap-3 items-center">
                    {#if overallRank}
                        <span class="flex items-center gap-1 text-sm font-semibold px-2.5 py-0.5 rounded-full bg-primary/15 text-primary border border-primary/25">
                            #{overallRank} Overall
                        </span>
                    {/if}
                    {#if overallStats && overallStats.streak > 1}
                        <span class="flex items-center gap-1 text-sm px-2.5 py-0.5 rounded-full bg-font-good/10 text-font-good border border-font-good/20">
                            <Flame class="h-3.5 w-3.5" strokeWidth={2} />{overallStats.streak} win streak
                        </span>
                    {/if}
                    {#if data.player.achievement_points > 0}
                        <span class="flex items-center gap-1 text-sm px-2.5 py-0.5 rounded-full bg-card-2 text-font-dim border border-card-edge">
                            <Star class="h-3.5 w-3.5" strokeWidth={2} />{data.player.achievement_points} pts
                        </span>
                    {/if}
                    {#if overallRating?.last_played}
                        <span class="flex items-center gap-1 text-sm text-font-dimer">
                            <Clock class="h-3.5 w-3.5" strokeWidth={1.5} />Last active {formatDate(overallRating.last_played)}
                        </span>
                    {/if}
                </div>
            </div>

            <!-- Big rating stat -->
            {#if overallRating}
                <div class="shrink-0 text-center md:text-right">
                    <div class="font-fancy text-5xl font-bold text-font-clear">
                        {Math.round(Number(overallRating.rating))}
                    </div>
                    <div class="text-font-dimer text-sm mt-0.5">
                        {#if Number(overallRating.rd) >= 110}
                            ±{Math.round(Number(overallRating.rd))} RD
                        {:else}
                            Overall Rating
                        {/if}
                    </div>
                </div>
            {/if}
        </div>
    </div>

    <!-- ── Body ────────────────────────────────────────────────────────────── -->
    <div class="flex flex-col md:flex-row gap-4 items-start">

        <!-- ── Sidebar ───────────────────────────────────────────────────── -->
        <div class="w-full md:w-64 md:shrink-0 flex flex-col gap-4">

            <!-- Category ratings -->
            <div class="rounded-2xl border border-card-edge bg-card shadow-md shadow-darken overflow-hidden">
                <div class="px-4 pt-4 pb-2">
                    <span class="font-fancy text-xs font-semibold tracking-widest uppercase text-font-dimer">Performance</span>
                </div>
                {#each cats as cat}
                    {@const r = getRating(cat)}
                    {@const s = getStats(cat)}
                    {@const rank = getRank(cat)}
                    {#if r || s}
                        <div class="flex items-center px-4 py-2.5 not-last:border-b border-card-edge hover:bg-select transition-colors duration-100">
                            <div class="flex-1 min-w-0">
                                <div class="font-fancy text-sm font-semibold text-font-dim">{categoryLabel[cat]}</div>
                                {#if s}
                                    <div class="text-xs text-font-dimest mt-0.5">
                                        {s.games} games · {winRate(s.wins, s.games)}% WR
                                    </div>
                                {/if}
                            </div>
                            <div class="text-right shrink-0 ml-3">
                                {#if r}
                                    <div class="text-lg font-bold text-font-clear">{Math.round(Number(r.rating))}</div>
                                    {#if rank}
                                        <div class="text-xs text-font-dimest">#{rank}</div>
                                    {/if}
                                {:else}
                                    <div class="text-font-dimest text-sm">—</div>
                                {/if}
                            </div>
                        </div>
                    {/if}
                {/each}
            </div>

            <!-- Streaks & records -->
            {#if overallStats}
                <div class="rounded-2xl border border-card-edge bg-card shadow-md shadow-darken p-4 flex flex-col gap-3">
                    <span class="font-fancy text-xs font-semibold tracking-widest uppercase text-font-dimer">Records</span>
                    <div class="flex justify-between items-center">
                        <span class="text-sm text-font-dim">Current streak</span>
                        <span class="font-semibold {overallStats.streak > 0 ? 'text-font-good' : overallStats.streak < 0 ? 'text-font-bad' : 'text-font-dimer'}">
                            {overallStats.streak > 0 ? `+${overallStats.streak}` : overallStats.streak}
                        </span>
                    </div>
                    <div class="flex justify-between items-center">
                        <span class="text-sm text-font-dim">Best win streak</span>
                        <span class="font-semibold text-font-clear">{overallStats.highest_winstreak}</span>
                    </div>
                    <div class="flex justify-between items-center">
                        <span class="text-sm text-font-dim">Total games</span>
                        <span class="font-semibold text-font-clear">{overallStats.games}</span>
                    </div>
                    <div class="flex justify-between items-center">
                        <span class="text-sm text-font-dim">Win / Loss</span>
                        <span class="font-semibold">
                            <span class="text-font-good">{overallStats.wins}</span>
                            <span class="text-font-dimest mx-1">/</span>
                            <span class="text-font-bad">{overallStats.games - overallStats.wins}</span>
                        </span>
                    </div>
                    {#if data.highestRating?.max_rating}
                        <div class="flex justify-between items-center">
                            <span class="text-sm text-font-dim">Highest rating</span>
                            <span class="font-semibold text-font-clear">{data.highestRating.max_rating}</span>
                        </div>
                    {/if}
                </div>
            {/if}

            <!-- Personal stat -->
            {#if data.personalStats}
                {@const now = new Date()}
                {@const lastGame = data.personalStats.last_game_date ? new Date(data.personalStats.last_game_date) : null}
                {@const firstGame = data.personalStats.first_game_date ? new Date(data.personalStats.first_game_date) : null}
                {@const daysSinceLast = lastGame ? Math.floor((now.getTime() - lastGame.getTime()) / (1000 * 60 * 60 * 24)) : 0}
                {@const daysSinceFirst = firstGame ? Math.floor((now.getTime() - firstGame.getTime()) / (1000 * 60 * 60 * 24)) : 0}
                <div class="rounded-2xl border border-primary/15 bg-card shadow-md shadow-darken p-4 flex flex-col gap-2">
                    {#if pid === 2}
                        <span class="text-xs text-center text-font-dim">Was robbed of the Colosseum Wonder:</span>
                        <span class="text-lg text-center font-fancy text-gradient-primary font-black">{data.personalStats.colosseum_robbed} times</span>
                    {:else if pid === 4}
                        <span class="text-xs text-center text-font-dim">Took the diplomatic exit:</span>
                        <span class="text-lg text-center font-fancy text-gradient-primary font-black">{data.personalStats.diplomatic_wins} {data.personalStats.diplomatic_wins === 1 ? 'time' : 'times'}</span>
                    {:else if pid === 5}
                        <span class="text-xs text-center text-font-dim">Hasn't touched this game since:</span>
                        <span class="text-lg text-center font-fancy text-gradient-primary font-black">{daysSinceLast} days</span>
                    {:else if pid === 3}
                        <span class="text-xs text-center text-font-dim">Plays Civ VI on a toaster since:</span>
                        <span class="text-lg text-center font-fancy text-gradient-primary font-black">{daysSinceFirst} days</span>
                    {:else if pid === 7}
                        <span class="text-xs text-center text-font-dim">Has founded:</span>
                        <span class="text-lg text-center font-fancy text-gradient-primary font-black">{data.personalStats.cities_founded} cities</span>
                    {:else if pid === 6}
                        <span class="text-xs text-center text-font-dim">Has won a game:</span>
                        <span class="text-lg text-center font-fancy text-gradient-primary font-black">{data.personalStats.wins} {data.personalStats.wins === 1 ? 'time' : 'times'}</span>
                    {:else if pid === 8}
                        <span class="text-xs text-center text-font-dim">Asked to start with 1200 Elo:</span>
                        <span class="text-lg text-center font-fancy text-gradient-primary font-black">1 times</span>
                    {:else if pid === 1}
                        <span class="text-xs text-center text-font-dim">Spent cash on API keys for this website:</span>
                        <span class="text-lg text-center font-fancy text-gradient-primary font-black">85 CHF</span>
                    {/if}
                </div>
            {/if}
            
            <!-- Achievements -->
            {#if data.achievements.length > 0}
                <div class="rounded-2xl border border-card-edge bg-card shadow-md shadow-darken p-4 flex flex-col gap-3">
                    <div class="flex items-center justify-between">
                        <span class="font-fancy text-xs font-semibold tracking-widest uppercase text-font-dimer">Achievements</span>
                        <span class="text-xs text-font-dimest">{earnedCount}/{totalCount}</span>
                    </div>
                    <!-- Progress bar -->
                    <div class="h-1.5 rounded-full bg-card-edge overflow-hidden">
                        <div
                            class="h-full rounded-full bg-gradient-primary transition-all duration-500"
                            style="width: {Math.round(earnedCount / totalCount * 100)}%"
                        ></div>
                    </div>
                    <!-- Achievement badges -->
                    <div class="flex flex-col gap-1.5 mt-1">
                        {#each data.achievements as ach}
                            {@const earned = earnedIds.has(ach.id)}
                            <div class="flex items-start gap-2 {earned ? '' : 'opacity-35'}">
                                <span class="mt-[3px] shrink-0 text-base leading-none {difficultyColor[ach.difficulty]}">●</span>
                                <div class="min-w-0">
                                    <div class="text-xs font-semibold text-font-dim leading-tight truncate" title={ach.name}>{ach.name}</div>
                                    <div class="text-[10px] text-font-dimest leading-tight mt-0.5">{ach.description}</div>
                                </div>
                            </div>
                        {/each}
                    </div>
                </div>
            {/if}
        </div>

        <!-- ── Main content ───────────────────────────────────────────────── -->
        <div class="flex-1 min-w-0 flex flex-col gap-4">

            <!-- Rating history chart -->
            {#if data.ratingHistory.length > 1}
                <div class="rounded-2xl border border-card-edge bg-card shadow-md shadow-darken px-5 pt-5 pb-3 overflow-hidden">
                    <div class="mb-3">
                        <span class="font-fancy text-xs font-semibold tracking-widest uppercase text-font-dimer">
                            Rating History
                        </span>
                    </div>
                    <RatingChart history={data.ratingHistory as any[]} />
                </div>
            {/if}

            <!-- Statistics row -->
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">

                <!-- Victory types -->
                {#if data.victoryStats.length > 0}
                    <div class="rounded-2xl border border-card-edge bg-card shadow-md shadow-darken p-5">
                        <span class="font-fancy text-xs font-semibold tracking-widest uppercase text-font-dimer block mb-4">
                            Victory Breakdown
                        </span>
                        <div class="flex flex-col gap-3">
                            {#each data.victoryStats as v}
                                {@const wr = winRate(v.wins, v.total)}
                                <div>
                                    <div class="flex items-center justify-between mb-1">
                                        <div class="flex items-center gap-1.5">
                                            <img src={victoryIcons[v.victory_type] ?? dominationv} alt="" class="h-4.5 w-4.5" />
                                            <span class="text-sm text-font-dim">{v.victory_type}</span>
                                        </div>
                                        <div class="flex items-center gap-3 text-xs text-font-dimer">
                                            <span class="text-font-good font-semibold">{v.wins}W</span>
                                            <span>{v.total} games</span>
                                            <span class="w-8 text-right font-semibold text-font-dim">{wr}%</span>
                                        </div>
                                    </div>
                                    <!-- Background bar (total) -->
                                    <div class="h-1.5 rounded-full bg-card-edge overflow-hidden">
                                        <div class="h-full rounded-full relative" style="width: {Math.round(Number(v.total) / victoryMaxTotal * 100)}%; background: {victoryColors[v.victory_type] ?? 'var(--color-font-dimest)'}; opacity: 0.35;"></div>
                                    </div>
                                    <!-- Win bar -->
                                    <div class="h-1.5 rounded-full bg-transparent overflow-hidden -mt-1.5">
                                        <div class="h-full rounded-full" style="width: {Math.round(Number(v.wins) / victoryMaxTotal * 100)}%; background: {victoryColors[v.victory_type] ?? 'var(--color-font-dimest)'}; opacity: 0.9;"></div>
                                    </div>
                                </div>
                            {/each}
                        </div>
                    </div>
                {/if}

                <!-- Favourite leaders -->
                {#if data.leaderStats.length > 0}
                    <div class="rounded-2xl border border-card-edge bg-card shadow-md shadow-darken p-5">
                        <span class="font-fancy text-xs font-semibold tracking-widest uppercase text-font-dimer block mb-4">
                            Favourite Leaders
                        </span>
                        <div class="flex flex-col gap-2.5">
                            {#each data.leaderStats as ls}
                                {@const wr = winRate(ls.wins, ls.games)}
                                <div class="flex items-center gap-3">
                                    <div class="h-8 w-8 rounded-full bg-card-edge overflow-hidden shrink-0">
                                        {#if leaderPortrait(ls.leader)}
                                            <img
                                                src={leaderPortrait(ls.leader)}
                                                alt={ls.leader}
                                                class="h-full w-full object-cover"
                                                onerror={(e) => ((e.currentTarget as HTMLImageElement).style.display = 'none')}
                                            />
                                        {/if}
                                    </div>
                                    <div class="flex-1 min-w-0">
                                        <div class="flex items-baseline justify-between mb-0.5">
                                            <span class="text-sm text-font-dim truncate">{ls.leader}</span>
                                            <span class="text-xs text-font-dimer shrink-0 ml-2">{ls.games}g · {wr}%</span>
                                        </div>
                                        <div class="h-1 rounded-full bg-card-edge overflow-hidden">
                                            <div
                                                class="h-full rounded-full bg-gradient-primary"
                                                style="width: {wr}%"
                                            ></div>
                                        </div>
                                    </div>
                                </div>
                            {/each}
                        </div>
                    </div>
                {/if}
            </div>

            <!-- Recent games -->
            {#if data.recentGames.length > 0}
                <div class="rounded-2xl border border-card-edge bg-card shadow-md shadow-darken overflow-hidden">
                    <div class="px-5 py-4 border-b border-card-edge flex items-center justify-between">
                        <span class="font-fancy text-xs font-semibold tracking-widest uppercase text-font-dimer">Recent Games</span>
                        <a href="/matches" class="flex items-center gap-1 text-xs text-font-dimest hover:text-font-dim transition-colors duration-150">
                            All matches <ChevronRight class="h-3 w-3" />
                        </a>
                    </div>

                    {#each data.recentGames as game}
                        {@const d = formatDelta(game.pre_rating_overall, game.post_rating_overall)}
                        {@const opponents = game.players?.filter((p: any) => p.player_id !== data.player.id) ?? []}
                        <a
                            href="/matches/view/{game.id}"
                            class="flex items-center px-5 py-3 not-last:border-b border-card-edge hover:bg-select transition-colors duration-100 ease-out relative"
                        >
                            <!-- Win/loss accent -->
                            <div class="absolute left-0 top-0 bottom-0 w-0.5 {game.winner ? 'bg-font-good' : 'bg-font-bad'}"></div>

                            <!-- Leader portrait -->
                            <div class="h-9 w-9 rounded-full bg-card-edge overflow-hidden shrink-0 mr-3">
                                {#if leaderPortrait(game.leader)}
                                    <img
                                        src={leaderPortrait(game.leader)}
                                        alt={game.leader}
                                        class="h-full w-full object-cover"
                                        onerror={(e) => ((e.currentTarget as HTMLImageElement).style.display = 'none')}
                                    />
                                {/if}
                            </div>

                            <!-- Result + info -->
                            <div class="flex-1 min-w-0">
                                <div class="flex items-center gap-2">
                                    <span class="text-sm font-semibold {game.winner ? 'text-font-good' : 'text-font-bad'}">
                                        {game.winner ? 'Victory' : 'Defeat'}
                                    </span>
                                    <img src={victoryIcons[game.victory_type] ?? dominationv} alt="" class="h-4 w-4 opacity-70" />
                                    <span class="text-xs text-font-dimer">{game.victory_type}</span>
                                    <span class="text-xs text-font-dimest">&middot; {categoryLabel[game.category] ?? game.category}</span>
                                    {#if game.map}
                                        <span class="text-xs text-font-dimest hidden sm:inline">&middot; {game.map}</span>
                                    {/if}
                                </div>
                                <div class="flex items-center gap-1.5 mt-0.5">
                                    <span class="text-xs text-font-dimest">vs</span>
                                    <span class="text-xs text-font-dimer truncate">
                                        {opponents.map((p: any) => p.name).join(', ')}
                                    </span>
                                </div>
                            </div>

                            <!-- Rating + date -->
                            <div class="text-right shrink-0 ml-4">
                                <div class="text-sm font-semibold {d.cls}">{d.text}</div>
                                <div class="text-xs text-font-dimest">{formatDate(game.date)}</div>
                            </div>
                        </a>
                    {/each}
                </div>
            {:else}
                <div class="rounded-2xl border border-card-edge bg-card shadow-md shadow-darken p-10 text-center text-font-dimest">
                    No games played yet.
                </div>
            {/if}

        </div>
    </div>
</div>
