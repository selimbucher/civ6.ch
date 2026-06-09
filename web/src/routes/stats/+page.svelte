<script lang="ts">
    import domination from '$lib/assets/icons/vcondition/domination.png';
    import religious from '$lib/assets/icons/vcondition/religious.png';
    import diplomatic from '$lib/assets/icons/vcondition/diplomatic.png';
    import science from '$lib/assets/icons/vcondition/science.png';
    import culture from '$lib/assets/icons/vcondition/culture.png';
    import score from '$lib/assets/icons/vcondition/score.png';
    import capitulation from '$lib/assets/icons/vcondition/capitulation.png';
    import {
        Sword, BookOpen, Palette, Zap, Globe, Flag,
        TrendingUp, TrendingDown, Clock, Gamepad2, Shield, Timer,
        Trophy, Users,

		Medal,

		Star


    } from '@lucide/svelte';

    const leaderAssets = import.meta.glob<{ default: string }>(
        '$lib/assets/icons/leaders/*.webp',
        { eager: true }
    );

    function leaderPortrait(leader: string | null): string | null {
        if (!leader) return null;
        const slug = leader.trim().replace(/\s+/g, '_');
        const norm = (s: string) =>
            s.normalize('NFD').replace(/[̀-ͯ]/g, '').toLowerCase();
        const keys = Object.keys(leaderAssets);
        const key =
            keys.find(k => k.includes(`/${slug}_(Civ6).`)) ??
            keys.find(k => k.toLowerCase().includes(`/${slug.toLowerCase()}_(civ6).`)) ??
            keys.find(k => norm(k).includes(`/${norm(slug)}_(civ6).`));
        return key ? leaderAssets[key].default : null;
    }

    let { data } = $props();

    const { globalStats, byType, mostLosses, mostGames, biggestGain, biggestLoss, mostTurns, shortestGame, victoryDist, topLeaders } = $derived(data);

    const victoryIcons: Record<string, string> = {
        Domination: domination,
        Religious: religious,
        Diplomatic: diplomatic,
        Science: science,
        Culture: culture,
        Score: score,
        Capitulation: capitulation,
    };

    const victoryAccent: Record<string, string> = {
        Domination:  'var(--color-army-1)',
        Religious:   'var(--color-faith-1)',
        Diplomatic:  'var(--color-diplo-1)',
        Science:     'var(--color-science-1)',
        Culture:     'var(--color-culture-1)',
        Score:       'var(--color-score-1)',
        Capitulation:'var(--color-font-dimest)',
    };

    const totalVictories = $derived(
        victoryDist.reduce((s: number, v: any) => s + Number(v.count), 0)
    );

    const topLeaderMax = $derived(
        topLeaders.length > 0 ? Math.max(...topLeaders.map((l: any) => Number(l.picks))) : 1
    );

    function shortName(name: string) {
        const parts = name.split(' ');
        return parts[0];
    }

    const awards = $derived([
        {
            title: 'Supreme Warlord',
            desc: 'A sword is better than a vote.',
            icon: Sword,
            accent: 'var(--color-army-1)',
            player: byType['Domination']?.name,
            stat: byType['Domination']?.count,
            unit: 'dom. wins',
        },
        {
            title: 'Grand Inquisitor',
            desc: 'Converting the heathens since 4000 BC.',
            icon: BookOpen,
            accent: 'var(--color-faith-1)',
            player: byType['Religious']?.name,
            stat: byType['Religious']?.count,
            unit: 'rel. wins',
        },
        {
            title: 'Culture Vulture',
            desc: 'Making everyone watch their movies.',
            icon: Palette,
            accent: 'var(--color-culture-1)',
            player: byType['Culture']?.name,
            stat: byType['Culture']?.count,
            unit: 'cult. wins',
        },
        {
            title: 'Rocket Surgeon',
            desc: 'Actually read the research tree.',
            icon: Zap,
            accent: 'var(--color-science-1)',
            player: byType['Science']?.name,
            stat: byType['Science']?.count,
            unit: 'sci. wins',
        },
        {
            title: 'Professional Hand-Shaker',
            desc: 'Friends with everyone. Somehow.',
            icon: Globe,
            accent: 'var(--color-diplo-1)',
            player: byType['Diplomatic']?.name,
            stat: byType['Diplomatic']?.count,
            unit: 'dipl. wins',
        },
        {
            title: 'Terror Incarnate',
            desc: 'Opponents quit before the battle begins.',
            icon: Flag,
            accent: 'var(--color-primary)',
            player: byType['Capitulation']?.name,
            stat: byType['Capitulation']?.count,
            unit: 'capit. wins',
        },
        {
            title: 'Dedicated Citizen',
            desc: 'Has no concept of stopping.',
            icon: Gamepad2,
            accent: 'var(--color-font-dimer)',
            player: mostGames?.name,
            stat: mostGames?.count,
            unit: 'games played',
        },
        {
            title: 'Professional Loser',
            desc: 'The floor of other people\'s Elo gains.',
            icon: Shield,
            accent: 'var(--color-font-bad)',
            player: mostLosses?.name,
            stat: mostLosses?.count,
            unit: 'losses',
        },
        {
            title: 'Elo Alchemist',
            desc: 'Turned a single match into serious gains.',
            icon: TrendingUp,
            accent: 'var(--color-font-good)',
            player: biggestGain?.name,
            stat: biggestGain?.delta ? `+${biggestGain.delta}` : undefined,
            unit: 'in one game',
        },
        {
            title: 'Falling Star',
            desc: 'What goes up must come down.',
            icon: TrendingDown,
            accent: 'var(--color-font-bad)',
            player: biggestLoss?.name,
            stat: biggestLoss?.delta ? `-${biggestLoss.delta}` : undefined,
            unit: 'in one game',
        },
        {
            title: 'Civilization Historian',
            desc: 'Has personally witnessed the most history.',
            icon: Clock,
            accent: 'var(--color-secondary)',
            player: mostTurns?.name,
            stat: mostTurns?.count?.toLocaleString(),
            unit: 'total turns',
        },
        {
            title: 'Speed Merchant',
            desc: 'Either very skilled or very frightening.',
            icon: Timer,
            accent: 'var(--color-gold-1)',
            player: shortestGame?.name,
            stat: shortestGame?.turns,
            unit: 'turn game',
        },
    ]);
</script>

<div class="mx-4 md:mx-12 mb-12 flex flex-col gap-6">

    <!-- ── Page Header ──────────────────────────────────────────────────── -->
    <div class="relative rounded-2xl border border-card-edge bg-card shadow-md shadow-darken overflow-hidden">
        <div class="absolute top-0 left-0 w-full h-0.75 bg-gradient-primary"></div>
        <div class="px-10 py-8 flex items-center justify-between">
            <div>
                <h1 class="font-fancy text-4xl font-bold text-font-clear tracking-wide"
                    style="text-shadow: 1px 1px 0px var(--color-primary-shadow);">
                    Hall of Records
                </h1>
                <p class="text-font-dimer mt-1 text-sm">
                    A rigorous scientific analysis of competitive Civilization VI performance.
                </p>
            </div>
            <Star strokeWidth={1.5} class="h-10 w-10 text-primary opacity-100 shrink-0" />
        </div>
    </div>

    <!-- ── Global Stats ─────────────────────────────────────────────────── -->
    <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
        <div class="rounded-2xl border border-card-edge bg-card shadow-md shadow-darken p-5 flex flex-col gap-1">
            <span class="font-fancy text-[10px] font-semibold tracking-widest uppercase text-font-dimest">Total Games</span>
            <span class="font-fancy text-4xl font-bold text-gradient-primary">{globalStats.total_games}</span>
            <span class="text-xs text-font-dimer">civilizations risen & fallen</span>
        </div>
        <div class="rounded-2xl border border-card-edge bg-card shadow-md shadow-darken p-5 flex flex-col gap-1">
            <span class="font-fancy text-[10px] font-semibold tracking-widest uppercase text-font-dimest">Turns Played</span>
            <span class="font-fancy text-4xl font-bold text-gradient-primary">{globalStats.total_turns.toLocaleString()}</span>
            <span class="text-xs text-font-dimer">decisions of questionable wisdom</span>
        </div>
        <div class="rounded-2xl border border-card-edge bg-card shadow-md shadow-darken p-5 flex flex-col gap-1">
            <span class="font-fancy text-[10px] font-semibold tracking-widest uppercase text-font-dimest">Players</span>
            <span class="font-fancy text-4xl font-bold text-gradient-primary">{globalStats.total_players}</span>
            <span class="text-xs text-font-dimer">civilizations with wifi access</span>
        </div>
        <div class="rounded-2xl border border-card-edge bg-card shadow-md shadow-darken p-5 flex flex-col gap-1">
            <span class="font-fancy text-[10px] font-semibold tracking-widest uppercase text-font-dimest">Avg. Length</span>
            <span class="font-fancy text-4xl font-bold text-gradient-primary">{globalStats.avg_turns}</span>
            <span class="text-xs text-font-dimer">turns before someone gives up</span>
        </div>
    </div>

    <!-- ── Order of Merit ───────────────────────────────────────────────── -->
    <div>
        <div class="flex items-center gap-4 mb-4">
            <span class="font-fancy text-xs font-semibold tracking-widest uppercase text-font-dimest">Order of Merit</span>
            <div class="h-px flex-1 bg-card-edge"></div>
        </div>
        <div class="grid grid-cols-2 md:grid-cols-4 gap-3">
            {#each awards as award}
                {#if award.player}
                    <div class="rounded-2xl border border-card-edge bg-card shadow-md shadow-darken overflow-hidden flex flex-col">
                        <!-- accent bar -->
                        <div class="h-[2px] w-full shrink-0" style="background: {award.accent};"></div>
                        <div class="p-4 flex flex-col gap-3 flex-1">
                            <div class="flex items-start gap-2.5">
                                <div class="shrink-0 mt-0.5 p-1.5 rounded-lg bg-card-2 border border-card-edge">
                                    <award.icon class="h-4 w-4" style="color: {award.accent};" strokeWidth={1.8} />
                                </div>
                                <div class="min-w-0">
                                    <div class="font-fancy text-xs font-semibold text-font-clear leading-tight">{award.title}</div>
                                    <div class="text-[10px] text-font-dimest leading-tight mt-0.5 italic">{award.desc}</div>
                                </div>
                            </div>
                            <div class="mt-auto pt-2 border-t border-card-edge flex items-baseline justify-between gap-1">
                                <span class="text-sm text-font-dim truncate">{shortName(award.player)}</span>
                                <div class="flex items-baseline gap-1 shrink-0">
                                    <span class="font-fancy font-bold text-font-clear tabular-nums">{award.stat}</span>
                                    <span class="text-[10px] text-font-dimest">{award.unit}</span>
                                </div>
                            </div>
                        </div>
                    </div>
                {/if}
            {/each}
        </div>
    </div>

    <!-- ── Bottom Section ───────────────────────────────────────────────── -->
    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">

        <!-- Victory Chronicle -->
        <div class="rounded-2xl border border-card-edge bg-card shadow-md shadow-darken p-5 flex flex-col gap-4">
            <div class="flex items-center gap-3">
                <span class="font-fancy text-xs font-semibold tracking-widest uppercase text-font-dimer">Victory Chronicle</span>
                <div class="h-px flex-1 bg-card-edge"></div>
                <span class="text-xs text-font-dimest tabular-nums">{totalVictories} games</span>
            </div>
            <div class="flex flex-col gap-3">
                {#each victoryDist as v}
                    {@const pct = Math.round(Number(v.count) / totalVictories * 100)}
                    <div>
                        <div class="flex items-center justify-between mb-1.5">
                            <div class="flex items-center gap-2">
                                <img src={victoryIcons[v.victory_type] ?? domination} alt="" class="h-5 w-5" />
                                <span class="text-sm text-font-dim">{v.victory_type}</span>
                            </div>
                            <div class="flex items-center gap-3 text-xs">
                                <span class="font-semibold text-font-clear tabular-nums">{v.count}</span>
                                <span class="text-font-dimest w-8 text-right tabular-nums">{pct}%</span>
                            </div>
                        </div>
                        <div class="h-2 rounded-full bg-card-edge overflow-hidden">
                            <div
                                class="h-full rounded-full transition-all duration-500"
                                style="width: {pct}%; background: {victoryAccent[v.victory_type] ?? 'var(--color-font-dimest)'};"
                            ></div>
                        </div>
                    </div>
                {/each}
            </div>
        </div>

        <!-- Leader Tier List -->
        <div class="rounded-2xl border border-card-edge bg-card shadow-md shadow-darken p-5 flex flex-col gap-4">
            <div class="flex items-center gap-3">
                <span class="font-fancy text-xs font-semibold tracking-widest uppercase text-font-dimer">Leader Tier List</span>
                <div class="h-px flex-1 bg-card-edge"></div>
                <span class="text-xs text-font-dimest">by popularity</span>
            </div>
            <div class="flex flex-col gap-2.5">
                {#each topLeaders as l}
                    {@const wr = l.picks > 0 ? Math.round(Number(l.wins) / Number(l.picks) * 100) : 0}
                    {@const barWidth = Math.round(Number(l.picks) / topLeaderMax * 100)}
                    <div class="flex items-center gap-3">
                        <div class="h-8 w-8 rounded-full bg-card-edge overflow-hidden shrink-0">
                            {#if leaderPortrait(l.leader)}
                                <img
                                    src={leaderPortrait(l.leader)!}
                                    alt={l.leader}
                                    class="h-full w-full object-cover"
                                    onerror={(e) => ((e.currentTarget as HTMLImageElement).style.display = 'none')}
                                />
                            {/if}
                        </div>
                        <div class="flex-1 min-w-0">
                            <div class="flex items-baseline justify-between mb-0.5">
                                <span class="text-sm text-font-dim truncate">{l.leader}</span>
                                <span class="text-xs text-font-dimer shrink-0 ml-2 tabular-nums">
                                    {l.picks}g · <span class="{wr >= 50 ? 'text-font-good' : wr >= 35 ? 'text-font-dim' : 'text-font-bad'}">{wr}%</span>
                                </span>
                            </div>
                            <div class="h-1.5 rounded-full bg-card-edge overflow-hidden">
                                <div
                                    class="h-full rounded-full transition-all duration-500"
                                    style="width: {barWidth}%; background: linear-gradient(90deg, var(--color-primary), var(--color-primary-2));"
                                ></div>
                            </div>
                        </div>
                    </div>
                {/each}
            </div>
        </div>

    </div>
</div>
