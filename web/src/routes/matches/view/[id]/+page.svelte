<script lang="ts">
    import dominationv from '$lib/assets/icons/vcondition/domination.png';
    import religiousv from '$lib/assets/icons/vcondition/religious.png';
    import diplomaticv from '$lib/assets/icons/vcondition/diplomatic.png';
    import sciencev from '$lib/assets/icons/vcondition/science.png';
    import culturev from '$lib/assets/icons/vcondition/culture.png';
    import scorev from '$lib/assets/icons/vcondition/score.png';
    import capitulationv from '$lib/assets/icons/vcondition/capitulation.png';
    import universalv from '$lib/assets/icons/vcondition/universal.png';

    import scoreIcon from '$lib/assets/icons/yields/score_upscaled.png';
    import scienceIcon from '$lib/assets/icons/yields/science.png';
    import cultureIcon from '$lib/assets/icons/yields/culture.png';
    import faithIcon from '$lib/assets/icons/yields/faith.png';
    import goldIcon from '$lib/assets/icons/yields/gold.png';
    import foodIcon from '$lib/assets/icons/yields/food.png';
    import productionIcon from '$lib/assets/icons/yields/production.png';
    import populationIcon from '$lib/assets/icons/emblems/population.png';
    import tourismIcon from '$lib/assets/icons/yields/tourism.png';
    import favorIcon from '$lib/assets/icons/yields/favor.png';

    import cityIcon from '$lib/assets/icons/emblems/city.png';

    import { CloudDownload, MapPin, Map, Gauge, FlaskRound } from '@lucide/svelte';

    import type { PageData } from './$types';

    let { data }: { data: PageData } = $props();

    const game    = $derived(data.game);
    const hasMap  = $derived(data.hasMap);
    const hasSave = $derived(data.hasSave);

    const leaderAssets = import.meta.glob<{ default: string }>(
        '$lib/assets/icons/leaders/*.webp',
        { eager: true }
    );

    const wonderAssets = import.meta.glob<{ default: string }>(
        '$lib/assets/icons/wonders/*.png',
        { eager: true }
    );
    function wonderIcon(wonder: string): string | null {
        const slug = wonder.trim().toLowerCase().replace(/[\s.\-]+/g, '_').replace(/'/g, '');
        const key = Object.keys(wonderAssets).find(k => k.toLowerCase().includes(`/${slug}.png`));
        return key ? wonderAssets[key].default : null;
    }

    const religionAssets = import.meta.glob<{ default: string }>(
        '$lib/assets/icons/religions/*.png',
        { eager: true }
    );
    function religionIcon(religion: string): string | null {
        const slug = religion.trim().toLowerCase().replace(/\s+/g, '_');
        const key = Object.keys(religionAssets).find(k =>
            k.toLowerCase().includes(`/${slug}.png`) || k.toLowerCase().includes(`_${slug}.png`)
        );
        return key ? religionAssets[key].default : null;
    }

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

    const victoryIcons: Record<string, string> = {
        'Domination':   dominationv,
        'Religious':    religiousv,
        'Diplomatic':   diplomaticv,
        'Science':      sciencev,
        'Culture':      culturev,
        'Score':        scorev,
        'Capitulation': capitulationv,
    };

    const victoryColor: Record<string, string> = {
        'Domination':   'var(--color-army-2)',
        'Religious':    'var(--color-faith-2)',
        'Diplomatic':   'var(--color-diplo-2)',
        'Science':      'var(--color-science-2)',
        'Culture':      'var(--color-culture-2)',
        'Score':        'var(--color-score-2)',
        'Capitulation': 'var(--color-font-dimer)',
    };

    function normaliseLabel(value: string | null | undefined): string {
        if (!value) return '—';
        const prefixes = ['MAPSIZE_', 'GAMESPEED_'];
        let s = value;
        for (const p of prefixes) {
            if (s.startsWith(p)) { s = s.slice(p.length); break; }
        }
        if (s !== s.toUpperCase()) return s;
        return s.split('_')
                .map(w => w.length ? w[0].toUpperCase() + w.slice(1).toLowerCase() : '')
                .join(' ');
    }

    function normaliseCityName(name: string): string {
        let s = name;
        let isLoc = false;
        for (const p of ['LOC_CITY_NAME_', 'LOC_CITY_']) {
            if (s.startsWith(p)) { s = s.slice(p.length); isLoc = true; break; }
        }
        if (s.includes('_')) isLoc = true;
        if (!isLoc) return name;
        return s.split('_')
                .map(w => w.length ? w[0].toUpperCase() + w.slice(1).toLowerCase() : '')
                .join(' ');
    }

    const categoryLabel: Record<string, string> = {
        ffa:   'FFA',
        teams: 'Teams',
        '1v1': '1v1',
    };

    function formatDate(date: string) {
        const d = new Date(date);
        const diff = Date.now() - d.getTime();
        const days = Math.floor(diff / 86400000);
        if (days === 0) return 'today';
        if (days === 1) return 'yesterday';
        if (days < 7) return `${days} days ago`;
        if (days < 14) return 'last week';
        if (days < 30) return `${Math.floor(days / 7)} weeks ago`;
        if (days < 60) return 'last month';
        if (days < 365) return `${Math.floor(days / 30)} months ago`;
        if (days < 730) return '1 year ago';
        return `${Math.floor(days / 365.25)} years ago`;
    }

    function formatDelta(pre: number, post: number) {
        const delta = Math.round(post - pre);
        if (delta > 0) return { text: `+${delta}`, cls: 'text-font-good' };
        if (delta < 0) return { text: `${delta}`, cls: 'text-font-bad' };
        return { text: '+0', cls: 'text-font-dim' };
    }

    function fmt(val: number | null): string {
        return val != null ? String(val) : '—';
    }

    function getTeams(players: any[]): any[][] {
        const map: Record<number, any[]> = {};
        for (const p of players) {
            if (!map[p.team]) map[p.team] = [];
            map[p.team].push(p);
        }
        return Object.values(map).sort((a, b) => {
            const aWins = a.some((p: any) => p.winner) ? 1 : 0;
            const bWins = b.some((p: any) => p.winner) ? 1 : 0;
            return bWins - aWins;
        });
    }

    const teams = $derived(getTeams(game.players));

    const victoryConditions = $derived([
        { icon: dominationv,  allowed: game.allow_conquest   ?? true, label: 'Domination' },
        { icon: sciencev,     allowed: game.allow_science    ?? true, label: 'Science' },
        { icon: culturev,     allowed: game.allow_culture    ?? true, label: 'Culture' },
        { icon: scorev,       allowed: game.allow_score      ?? true, label: 'Score' },
        { icon: religiousv,   allowed: game.allow_religious  ?? true, label: 'Religious' },
        { icon: diplomaticv,  allowed: game.allow_diplomatic ?? true, label: 'Diplomatic' },
    ]);
</script>

<div class="mx-12 mb-12 flex flex-col gap-4">

    <!-- ── Top section ───────────────────────────────────────────────────── -->
    <div class="flex gap-4 items-start">

        <!-- Left: game header + players -->
        <div class="flex flex-col gap-4 w-[30rem] shrink-0">

            <!-- Game header -->
            <div class="rounded-2xl border border-card-edge bg-card shadow-md shadow-darken overflow-hidden">
                <div class="h-[3px] w-full bg-gradient-primary"></div>
                <div class="p-5 flex items-start gap-4">
                    <img src={victoryIcons[game.victory_type] ?? dominationv} alt="" class="h-14 w-14 shrink-0 mt-0.5" />
                    <div class="flex-1 min-w-0">
                        <div class="font-fancy font-semibold text-lg tracking-wide leading-tight"
                             style="color: {victoryColor[game.victory_type] ?? 'var(--color-font-clear)'};">
                            {game.victory_type} Victory
                        </div>
                        <div class="flex items-center gap-2 mt-1.5 flex-wrap">
                            <span class="font-fancy text-[10px] font-semibold tracking-wider uppercase px-2 py-0.5 rounded-full bg-card-2 border border-card-edge text-font-dim">
                                {categoryLabel[game.category] ?? game.category}
                            </span>
                            {#if game.map}
                                <span class="text-xs text-font-dimer">{game.map}</span>
                            {/if}
                            {#if game.turns}
                                <span class="text-xs text-font-dimest">· {game.turns} turns</span>
                            {/if}
                        </div>
                    </div>
                    <span class="text-xs text-font-dimest shrink-0">{formatDate(game.date)}</span>
                </div>
            </div>

            <!-- Players -->
            <div class="rounded-2xl border border-card-edge bg-card shadow-md shadow-darken overflow-hidden">
                <div class="px-5 pt-4 pb-2.5 border-b border-card-edge">
                    <span class="font-fancy text-xs font-semibold tracking-widest uppercase text-font-dimer">Players</span>
                </div>
                {#each teams as team, ti}
                    {#if ti > 0}
                        <div class="h-px bg-card-edge"></div>
                    {/if}
                    {#each team as player}
                        {@const d = formatDelta(player.pre_rating_overall, player.post_rating_overall)}
                        <div class="relative flex items-center px-5 py-3 hover:bg-select transition-colors duration-100">
                            <!-- Win/loss accent -->
                            <div class="absolute left-0 top-3 bottom-3 w-[3px] rounded-r-full {player.winner ? 'bg-font-good' : 'bg-font-bad'}"></div>

                            <!-- Portrait -->
                            <div class="h-11 w-11 rounded-full bg-card-edge overflow-hidden shrink-0 mr-3">
                                {#if leaderPortrait(player.leader)}
                                    <img src={leaderPortrait(player.leader)!} alt=""
                                         class="h-full w-full object-cover"
                                         onerror={(e) => ((e.currentTarget as HTMLImageElement).style.display = 'none')} />
                                {/if}
                            </div>

                            <!-- Name + meta -->
                            <div class="flex-1 min-w-0">
                                <div class="flex items-center gap-1.5">
                                    <a href="/profile/{player.player_id}"
                                       class="text-sm {player.winner ? 'text-font-clear font-semibold' : 'text-font-dim'} hover:text-font-clear transition-colors duration-150 truncate">
                                        {player.name}
                                    </a>
                                    {#if player.winner}
                                        <img src={universalv} alt="" class="h-4 w-4 shrink-0 opacity-75" />
                                    {/if}
                                </div>
                                {#if player.pseudo_name}
                                    <div class="text-xs text-font-dimest mt-0.5 truncate">{player.pseudo_name}</div>
                                {/if}
                            </div>

                            <!-- Rating -->
                            <div class="text-right shrink-0 ml-4">
                                <div class="text-sm font-semibold tabular-nums {d.cls}">{d.text}</div>
                                <div class="text-xs text-font-dimest tabular-nums">{Math.round(player.pre_rating_overall)}</div>
                            </div>
                        </div>
                    {/each}
                {/each}
            </div>
        </div>

        <!-- Right: map + settings row -->
        <div class="flex-1 flex flex-col gap-4 min-w-0">

            <!-- Map -->
            <div class="rounded-2xl border border-card-edge bg-card shadow-md shadow-darken overflow-hidden">
                {#if hasMap}
                    <img src="/files/maps/{game.id}" alt="" class="w-full block" />
                {:else}
                    <div class="flex flex-col items-center justify-center gap-2 p-16 text-font-dimest">
                        <Map strokeWidth={1} class="h-10 w-10" />
                        <span class="text-sm">Map not available</span>
                    </div>
                {/if}
            </div>

            <!-- Settings row -->
            <div class="flex gap-4">

                <!-- Game settings -->
                <div class="flex-1 rounded-2xl border border-card-edge bg-card shadow-md shadow-darken p-4 flex flex-col gap-3">
                    <span class="font-fancy text-[10px] font-semibold tracking-widest uppercase text-font-dimest">Settings</span>
                    <div class="flex flex-col gap-2">
                        <div class="flex items-center gap-2 text-sm text-font-dim">
                            <MapPin class="h-3.5 w-3.5 text-font-dimest shrink-0" />{game.map ?? '—'}
                        </div>
                        <div class="flex items-center gap-2 text-sm text-font-dim">
                            <Map class="h-3.5 w-3.5 text-font-dimest shrink-0" />{normaliseLabel(game.map_size)}
                        </div>
                        <div class="flex items-center gap-2 text-sm text-font-dim">
                            <Gauge class="h-3.5 w-3.5 text-font-dimest shrink-0" />{normaliseLabel(game.game_speed)}
                        </div>
                        <div class="flex items-center gap-2 text-sm text-font-dim">
                            <FlaskRound class="h-3.5 w-3.5 text-font-dimest shrink-0" />{game.shuffle_techs ? 'Shuffled Techs' : 'Default Techs'}
                        </div>
                    </div>
                </div>

                <!-- Victory conditions -->
                <div class="flex-1 rounded-2xl border border-card-edge bg-card shadow-md shadow-darken p-4 flex flex-col gap-3">
                    <span class="font-fancy text-[10px] font-semibold tracking-widest uppercase text-font-dimest">Victory Conditions</span>
                    <div class="grid grid-cols-3 gap-x-2 gap-y-3">
                        {#each victoryConditions as vc}
                            <div class="flex flex-col items-center gap-1 transition-opacity duration-150 {vc.allowed ? 'opacity-100' : 'opacity-25'}">
                                <img src={vc.icon} alt={vc.label} class="h-8 w-8" title={vc.label} />
                                <span class="font-fancy text-[9px] tracking-wide uppercase text-font-dimest">{vc.label}</span>
                            </div>
                        {/each}
                    </div>
                </div>

                <!-- Save file -->
                <div class="flex-1 rounded-2xl border border-card-edge bg-card shadow-md shadow-darken p-4 flex flex-col items-center justify-center gap-3">
                    {#if hasSave}
                        <CloudDownload strokeWidth={1.2} class="h-7 w-7 text-font-dim" />
                        <div class="text-xs text-font-dimer text-center leading-snug">
                            AutoSave_{String(game.turns).padStart(4, '0')}.Civ6Save
                        </div>
                        <a
                            href="/files/saves/{game.id}"
                            download="AutoSave_{String(game.turns).padStart(4, '0')}.Civ6Save"
                            class="rounded-lg opacity-85 bg-gradient-primary text-black text-xs font-semibold px-3 py-1.5 hover:opacity-100 transition-opacity duration-250"
                        >Download</a>
                    {:else}
                        <CloudDownload strokeWidth={1} class="h-7 w-7 text-font-dimest" />
                        <span class="text-xs text-font-dimest">Not available</span>
                    {/if}
                </div>

            </div>
        </div>
    </div>

    <!-- ── Player Yields ─────────────────────────────────────────────────── -->
    <div class="flex items-center gap-4">
        <div class="h-px flex-1 bg-card-edge"></div>
        <span class="font-fancy text-xs tracking-widest uppercase text-font-dimest">Player Yields</span>
        <div class="h-px flex-1 bg-card-edge"></div>
    </div>
    <div class="rounded-2xl border border-card-edge bg-card shadow-md shadow-darken overflow-hidden">
        <table class="w-full">
            <thead>
                <tr class="border-b border-card-edge">
                    <th class="w-12"></th>
                    <th class="text-left px-3 py-2.5 font-fancy text-[10px] tracking-widest uppercase text-font-dimest">Player</th>
                    <th class="text-right pr-3 py-2.5 font-fancy text-[10px] tracking-widest uppercase text-font-dimest">Score</th>
                    <th class="text-right pr-3 py-2.5 font-fancy text-[10px] tracking-widest uppercase text-font-dimest">Pop</th>
                    <th class="text-right pr-3 py-2.5 font-fancy text-[10px] tracking-widest uppercase text-font-dimest">Science</th>
                    <th class="text-right pr-3 py-2.5 font-fancy text-[10px] tracking-widest uppercase text-font-dimest">Culture</th>
                    <th class="text-right pr-3 py-2.5 font-fancy text-[10px] tracking-widest uppercase text-font-dimest">Food</th>
                    <th class="text-right pr-3 py-2.5 font-fancy text-[10px] tracking-widest uppercase text-font-dimest">Prod</th>
                    <th class="text-right pr-3 py-2.5 font-fancy text-[10px] tracking-widest uppercase text-font-dimest">Gold</th>
                    <th class="text-right pr-3 py-2.5 font-fancy text-[10px] tracking-widest uppercase text-font-dimest">Faith</th>
                    <th class="text-right pr-3 py-2.5 font-fancy text-[10px] tracking-widest uppercase text-font-dimest">Tourism</th>
                    <th class="text-right pr-5 py-2.5 font-fancy text-[10px] tracking-widest uppercase text-font-dimest">Favor</th>
                </tr>
            </thead>
            <tbody>
                {#each game.players as player}
                    <tr class="relative not-last:border-b border-card-edge hover:bg-select transition-colors duration-100">
                        <td class="py-2 pl-2 pr-0">
                            <div class="h-9 w-9 rounded-full bg-card-edge overflow-hidden">
                                {#if leaderPortrait(player.leader)}
                                    <img src={leaderPortrait(player.leader)!} alt=""
                                         class="h-full w-full object-cover"
                                         onerror={(e) => ((e.currentTarget as HTMLImageElement).style.display = 'none')} />
                                {/if}
                            </div>
                        </td>
                        <td class="px-3 py-3">
                            <a href="/profile/{player.player_id}"
                               class="{player.winner ? 'text-font-clear font-semibold' : 'text-font-dim'} hover:text-font-clear transition-colors duration-150 text-sm">
                                {player.name}
                                {#if player.winner}
                                    <img src={universalv} alt="" class="ml-0.5 inline-block h-4 opacity-75 mb-[0.05rem]" />
                                {/if}
                            </a>
                            {#if player.pseudo_name}
                                <div class="text-xs text-font-dimest mt-0.5">{player.pseudo_name}</div>
                            {/if}
                        </td>
                        <td class="pr-3 py-3"><span class="flex items-center justify-end gap-1"><img src={scoreIcon}      alt="" class="h-5 shrink-0" /><span class="text-score       text-sm font-bold tabular-nums">{fmt(player.score)}</span></span></td>
                        <td class="pr-3 py-3"><span class="flex items-center justify-end gap-1"><img src={populationIcon} alt="" class="h-5 shrink-0" /><span class="text-font-clear  text-sm font-bold tabular-nums">{fmt(player.population)}</span></span></td>
                        <td class="pr-3 py-3"><span class="flex items-center justify-end gap-1"><img src={scienceIcon}    alt="" class="h-5 shrink-0" /><span class="text-science     text-sm font-bold tabular-nums">{fmt(player.science)}</span></span></td>
                        <td class="pr-3 py-3"><span class="flex items-center justify-end gap-1"><img src={cultureIcon}    alt="" class="h-5 shrink-0" /><span class="text-culture     text-sm font-bold tabular-nums">{fmt(player.culture)}</span></span></td>
                        <td class="pr-3 py-3"><span class="flex items-center justify-end gap-1"><img src={foodIcon}       alt="" class="h-5 shrink-0" /><span class="text-food        text-sm font-bold tabular-nums">{fmt(player.food)}</span></span></td>
                        <td class="pr-3 py-3"><span class="flex items-center justify-end gap-1"><img src={productionIcon} alt="" class="h-5 shrink-0" /><span class="text-production  text-sm font-bold tabular-nums">{fmt(player.production)}</span></span></td>
                        <td class="pr-3 py-3"><span class="flex items-center justify-end gap-1"><img src={goldIcon}       alt="" class="h-5 shrink-0" /><span class="text-gold        text-sm font-bold tabular-nums">{fmt(player.gold)}</span></span></td>
                        <td class="pr-3 py-3"><span class="flex items-center justify-end gap-1"><img src={faithIcon}      alt="" class="h-5 shrink-0" /><span class="text-faith       text-sm font-bold tabular-nums">{fmt(player.faith)}</span></span></td>
                        <td class="pr-3 py-3"><span class="flex items-center justify-end gap-1"><img src={tourismIcon}    alt="" class="h-4 shrink-0" /><span class="text-production  text-sm font-bold tabular-nums">{fmt(player.tourism)}</span></span></td>
                        <td class="pr-5 py-3"><span class="flex items-center justify-end gap-1"><img src={favorIcon}      alt="" class="h-5 shrink-0" /><span class="text-diplo       text-sm font-bold tabular-nums">{fmt(player.favor)}</span></span></td>
                    </tr>
                {/each}
            </tbody>
        </table>
    </div>

    <!-- ── Cities ───────────────────────────────────────────────────────── -->
    <div class="flex items-center gap-4">
        <div class="h-px flex-1 bg-card-edge"></div>
        <span class="font-fancy text-xs tracking-widest uppercase text-font-dimest">Cities</span>
        <div class="h-px flex-1 bg-card-edge"></div>
    </div>
    <div class="rounded-2xl border border-card-edge bg-card shadow-md shadow-darken overflow-hidden">
        <table class="w-full">
            <tbody>
                {#each game.players as player, pi}
                    <!-- Player group header -->
                    <tr class="{pi > 0 ? 'border-t' : ''} border-card-edge bg-zebra-2">
                        <td colspan="99" class="px-5 py-3">
                            <div class="flex items-center gap-3">
                                <div class="h-10 w-10 rounded-full bg-card-edge overflow-hidden shrink-0">
                                    {#if leaderPortrait(player.leader)}
                                        <img src={leaderPortrait(player.leader)!} alt=""
                                             class="h-full w-full object-cover"
                                             onerror={(e) => ((e.currentTarget as HTMLImageElement).style.display = 'none')} />
                                    {/if}
                                </div>
                                <div>
                                    <a href="/profile/{player.player_id}"
                                       class="text-sm font-semibold {player.winner ? 'text-font-clear' : 'text-font-dim'} hover:text-primary transition-colors duration-150">
                                        {player.name}
                                        {#if player.winner}
                                            <img src={universalv} alt="" class="inline h-4 opacity-70 ml-1 mb-0.5" />
                                        {/if}
                                    </a>
                                    {#if player.pseudo_name}
                                        <div class="text-xs text-font-dimest mt-0.5">{player.pseudo_name}</div>
                                    {/if}
                                </div>
                            </div>
                        </td>
                    </tr>
                    {#each player.cities as city}
                        <tr class="border-t border-card-edge hover:bg-select transition-colors duration-100">
                            <td class="w-4"></td>
                            <td class="py-2 w-52 max-w-52 whitespace-nowrap overflow-hidden text-ellipsis">
                                <span class="flex items-center">
                                    <div class="h-5 w-5 p-0.5 rounded-full bg-select overflow-hidden mr-2 shrink-0">
                                        <img src={cityIcon} alt="" />
                                    </div>
                                    <span class="text-font-dim text-sm truncate" title={normaliseCityName(city.name)}>
                                        {normaliseCityName(city.name)}
                                    </span>
                                    {#if city.religion && religionIcon(city.religion)}
                                        <img src={religionIcon(city.religion)!} alt={city.religion}
                                             class="shrink-0 h-7 ml-2"
                                             style="filter: sepia(1) brightness(0.8) saturate(20) hue-rotate(348deg);"
                                             title={city.religion} />
                                    {/if}
                                </span>
                            </td>
                            <td>
                                <div class="flex items-center gap-1 overflow-hidden py-1">
                                    {#if city.wonders}
                                        {#each city.wonders as wonder}
                                            {#if wonderIcon(wonder)}
                                                <img src={wonderIcon(wonder)!} alt={wonder} title={wonder} class="h-7 object-contain" />
                                            {:else}
                                                <span class="text-xs border border-card-edge rounded px-1 text-font-dimest" title={wonder}>W</span>
                                            {/if}
                                        {/each}
                                    {/if}
                                </div>
                            </td>
                            <td class="w-6"></td>
                            <td class="pr-2 py-2"><span class="flex items-center justify-end gap-0.5"><img src={populationIcon} alt="" class="h-5 shrink-0" /><span class="text-font-clear  text-sm font-bold tabular-nums">{fmt(city.population)}</span></span></td>
                            <td class="pr-2 py-2"><span class="flex items-center justify-end gap-0.5"><img src={scienceIcon}    alt="" class="h-5 shrink-0" /><span class="text-science    text-sm font-bold tabular-nums">{fmt(Math.round(city.science))}</span></span></td>
                            <td class="pr-2 py-2"><span class="flex items-center justify-end gap-0.5"><img src={cultureIcon}    alt="" class="h-5 shrink-0" /><span class="text-culture    text-sm font-bold tabular-nums">{fmt(Math.round(city.culture))}</span></span></td>
                            <td class="pr-2 py-2"><span class="flex items-center justify-end gap-0.5"><img src={foodIcon}       alt="" class="h-5 shrink-0" /><span class="text-food       text-sm font-bold tabular-nums">{fmt(Math.round(city.food))}</span></span></td>
                            <td class="pr-2 py-2"><span class="flex items-center justify-end gap-0.5"><img src={productionIcon} alt="" class="h-5 shrink-0" /><span class="text-production text-sm font-bold tabular-nums">{fmt(Math.round(city.production))}</span></span></td>
                            <td class="pr-2 py-2"><span class="flex items-center justify-end gap-0.5"><img src={goldIcon}       alt="" class="h-5 shrink-0" /><span class="text-gold       text-sm font-bold tabular-nums">{fmt(Math.round(city.gold))}</span></span></td>
                            <td class="pr-3 py-2"><span class="flex items-center justify-end gap-0.5"><img src={faithIcon}      alt="" class="h-5 shrink-0" /><span class="text-faith      text-sm font-bold tabular-nums">{fmt(Math.round(city.faith))}</span></span></td>
                            <td class="w-4"></td>
                        </tr>
                    {/each}
                {/each}
            </tbody>
        </table>
    </div>
</div>
