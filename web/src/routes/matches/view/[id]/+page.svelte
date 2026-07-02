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
    import religionGeneric from '$lib/assets/icons/religions/Religion_Generic.png';

    import { CloudDownload, MapPin, Map, Gauge, Clock, Layers, Swords } from '@lucide/svelte';
    import Tooltip from '$lib/Tooltip.svelte';

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
    // The save records each religion's real symbol (its RELIGION_* type), stored
    // per city as religion_icon (e.g. "Custom10"). Resolve that to its distinct
    // glyph so every religion looks different instead of one generic icon.
    function religionIconByKey(key: string | null | undefined): string | null {
        if (!key) return null;
        // Prefer the solid "Pressure" silhouette — it tints cleanly into a bold
        // coloured symbol, unlike the thin outline of the base icon.
        const want = `/religion_pressure_${key}.png`.toLowerCase();
        const k = Object.keys(religionAssets).find((a) => a.toLowerCase().endsWith(want));
        return k ? religionAssets[k].default : null;
    }
    // Prefer the parsed symbol; fall back to name-matching, then the generic glyph
    // (covers rows parsed before religion_icon existed).
    function cityReligionIcon(city: any): string | null {
        return (
            religionIconByKey(city.religion_icon) ??
            (city.religion ? religionIcon(city.religion) : null) ??
            (city.religion ? religionGeneric : null)
        );
    }
    // The custom name, or — for the pre-baked religions (Islam, Catholicism, …)
    // which can't be renamed and so carry no custom name — the religion's own
    // name, which equals its icon key.
    function religionName(city: any): string {
        if (city.religion) return city.religion;
        const k = city.religion_icon;
        if (k && !String(k).startsWith('Custom') && k !== 'Generic') return k;
        return '';
    }
    // The religion a player founded (if any), resolved to symbol + name + colour.
    function foundedReligion(
        player: any
    ): { icon: string | null; name: string; color: string } | null {
        const key = player.founded_religion_icon;
        if (!key && !player.founded_religion) return null;
        let name = player.founded_religion || '';
        if (!name && key && !String(key).startsWith('Custom') && key !== 'Generic') name = key;
        return {
            icon: religionIconByKey(key),
            name,
            color: player.founded_religion_color ?? 'var(--color-font-dimer)'
        };
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

    function rulesetLabel(r: string | null | undefined): string {
        if (!r) return '—';
        const map: Record<string, string> = {
            STANDARD:    'Standard',
            EXPANSION_1: 'Rise & Fall',
            EXPANSION_2: 'Gathering Storm',
        };
        return map[r] ?? normaliseLabel(r);
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
    const hasCities = $derived(game.players.some((p: any) => p.cities?.length > 0));
    const hasYields = $derived(game.players.some((p: any) => p.score != null));

    const gameModes = $derived([
        { enabled: game.shuffle_techs,      label: 'Tech Shuffle',     color: 'text-science border-science/30 bg-science/10' },
        { enabled: game.secret_societies,   label: 'Secret Societies', color: 'text-secondary border-secondary/30 bg-secondary/10' },
        { enabled: game.heroes_and_legends, label: 'Heroes & Legends', color: 'text-font-good border-font-good/30 bg-font-good/10' },
        { enabled: game.apocalypse_mode,    label: 'Apocalypse',       color: 'text-font-bad border-font-bad/30 bg-font-bad/10' },
        { enabled: game.monopolies,         label: 'Monopolies',       color: 'text-gold-1 border-gold-1/30 bg-gold-1/10' },
        { enabled: game.barbarian_clans,    label: 'Barbarian Clans',  color: 'text-army-1 border-army-1/30 bg-army-1/10' },
        { enabled: game.zombie_defense,     label: 'Zombie Defence',   color: 'text-font-bad border-font-bad/30 bg-font-bad/10' },
    ].filter(m => m.enabled));

    const victoryConditions = $derived([
        { icon: dominationv,  allowed: game.allow_conquest   ?? true, label: 'Domination' },
        { icon: sciencev,     allowed: game.allow_science    ?? true, label: 'Science' },
        { icon: culturev,     allowed: game.allow_culture    ?? true, label: 'Culture' },
        { icon: scorev,       allowed: game.allow_score      ?? true, label: 'Score' },
        { icon: religiousv,   allowed: game.allow_religious  ?? true, label: 'Religious' },
        { icon: diplomaticv,  allowed: game.allow_diplomatic ?? true, label: 'Diplomatic' },
    ]);

    // One source of truth for every yield column, shared by the Player-Yields
    // table and the per-city tooltips so the two stay visually consistent.
    type YieldCol = { key: string; icon: string; label: string; cls: string; h: string };
    const yieldCols: YieldCol[] = [
        { key: 'score',      icon: scoreIcon,      label: 'Score',      cls: 'text-score',      h: 'h-4' },
        { key: 'population', icon: populationIcon, label: 'Population', cls: 'text-font-clear',  h: 'h-4' },
        { key: 'science',    icon: scienceIcon,    label: 'Science',    cls: 'text-science',     h: 'h-4' },
        { key: 'culture',    icon: cultureIcon,    label: 'Culture',    cls: 'text-culture',     h: 'h-4' },
        { key: 'food',       icon: foodIcon,       label: 'Food',       cls: 'text-food',        h: 'h-4' },
        { key: 'production', icon: productionIcon, label: 'Production', cls: 'text-production',   h: 'h-4' },
        { key: 'gold',       icon: goldIcon,       label: 'Gold',       cls: 'text-gold',        h: 'h-4' },
        { key: 'faith',      icon: faithIcon,      label: 'Faith',      cls: 'text-faith',       h: 'h-4' },
        { key: 'tourism',    icon: tourismIcon,    label: 'Tourism',    cls: 'text-production',   h: 'h-3.5' },
        { key: 'favor',      icon: favorIcon,      label: 'Favor',      cls: 'text-diplo',        h: 'h-4' },
    ];
    // Per-column leader, so the eye can jump to who topped each yield.
    const yieldLeaders = $derived.by(() => {
        const best: Record<string, number> = {};
        for (const col of yieldCols) {
            let max = -Infinity;
            for (const p of game.players) {
                const v = (p as any)[col.key];
                if (v != null && v > max) max = v;
            }
            best[col.key] = max;
        }
        return best;
    });
    // Yields shown inside a city's hover card (no Score/Tourism/Favor at city level).
    const cityYieldCols = yieldCols.filter((c) =>
        ['population', 'science', 'culture', 'food', 'production', 'gold', 'faith'].includes(c.key)
    );
</script>

<div class="mx-3 md:mx-12 mb-12 flex flex-col gap-4">

    <!-- ── Top section ───────────────────────────────────────────────────── -->
    <div class="flex flex-col md:flex-row gap-4 items-start">

        <!-- Left: game header + players -->
        <div class="flex flex-col gap-4 w-full md:w-[30rem] md:shrink-0">

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
                            <div class="h-11 w-11 rounded-full bg-card-edge overflow-hidden shrink-0 mr-3 {player.eliminated || player.left_game ? 'grayscale opacity-60' : ''}">
                                {#if leaderPortrait(player.leader)}
                                    <img src={leaderPortrait(player.leader)!} alt=""
                                         class="h-full w-full object-cover"
                                         onerror={(e) => ((e.currentTarget as HTMLImageElement).style.display = 'none')} />
                                {:else if player.leader}
                                    <div class="h-full w-full flex items-center justify-center text-font-dimest text-xs font-bold select-none">?</div>
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
                                    {#if player.eliminated}
                                        <span class="text-[9px] uppercase tracking-wider px-1.5 py-px rounded-full bg-font-bad/10 text-font-bad border border-font-bad/20 shrink-0">Eliminated</span>
                                    {:else if player.left_game}
                                        <span class="text-[9px] uppercase tracking-wider px-1.5 py-px rounded-full bg-font-dimer/10 text-font-dimer border border-font-dimer/20 shrink-0">Left</span>
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
            <div class="rounded-2xl {hasMap ? '' : 'border border-card-edge bg-card shadow-md shadow-darken'}  overflow-hidden">
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
                    <div class="grid grid-cols-2 gap-x-3 gap-y-2">
                        {#if game.map}
                        <div class="flex items-center gap-2 text-sm text-font-dim">
                            <MapPin class="h-3.5 w-3.5 text-font-dimest shrink-0" /><span class="truncate">{game.map}</span>
                        </div>
                        {/if}
                        {#if game.map_size}
                        <div class="flex items-center gap-2 text-sm text-font-dim">
                            <Map class="h-3.5 w-3.5 text-font-dimest shrink-0" /><span class="truncate">{normaliseLabel(game.map_size)}</span>
                        </div>
                        {/if}
                        {#if game.game_speed}
                        <div class="flex items-center gap-2 text-sm text-font-dim">
                            <Gauge class="h-3.5 w-3.5 text-font-dimest shrink-0" /><span class="truncate">{normaliseLabel(game.game_speed)}</span>
                        </div>
                        {/if}
                        {#if game.difficulty}
                        <div class="flex items-center gap-2 text-sm text-font-dim">
                            <Swords class="h-3.5 w-3.5 text-font-dimest shrink-0" /><span class="truncate">{normaliseLabel(game.difficulty)}</span>
                        </div>
                        {/if}
                        {#if game.era}
                        <div class="flex items-center gap-2 text-sm text-font-dim">
                            <Clock class="h-3.5 w-3.5 text-font-dimest shrink-0" /><span class="truncate">{normaliseLabel(game.era)} Era</span>
                        </div>
                        {/if}
                        {#if game.ruleset}
                        <div class="flex items-center gap-2 text-sm text-font-dim">
                            <Layers class="h-3.5 w-3.5 text-font-dimest shrink-0" /><span class="truncate">{rulesetLabel(game.ruleset)}</span>
                        </div>
                        {/if}
                    </div>
                    {#if gameModes.length > 0}
                        <div class="flex flex-wrap gap-1.5 pt-2 mt-1 border-t border-card-edge">
                            {#each gameModes as mode}
                                <span class="text-[10px] font-semibold px-2 py-0.5 rounded-full border {mode.color}">
                                    {mode.label}
                                </span>
                            {/each}
                        </div>
                    {/if}
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
                <div class="flex-1 rounded-2xl border border-card-edge bg-card shadow-md shadow-darken p-3 flex flex-col items-center justify-center gap-2.5">
                    {#if hasSave}
                        <CloudDownload strokeWidth={1.2} class="h-9 w-9 text-font-dim" />
                        <div class="text-sm text-font-dimer text-center leading-snug mb-2">
                            {game.save_filename ?? `AutoSave_${String(game.turns).padStart(4, '0')}.Civ6Save`}
                        </div>
                        <a
                            href="/files/saves/{game.id}"
                            download={game.save_filename ?? `AutoSave_${String(game.turns).padStart(4, '0')}.Civ6Save`}
                            class="rounded-lg border border-primary text-primary text-xs font-semibold px-3 py-1.5 hover:bg-primary hover:text-black transition-colors duration-200"
                        >Download</a>
                    {:else}
                        <CloudDownload strokeWidth={1} class="h-9 w-9 text-font-dimest" />
                        <span class="text-sm text-font-dimest">Not available</span>
                    {/if}
                </div>

            </div>
        </div>
    </div>

    <!-- ── Empires: each player's totals, founded religion and cities ──────── -->
    {#if hasYields || hasCities}
    <div class="flex items-center gap-4">
        <div class="h-px flex-1 bg-card-edge"></div>
        <span class="font-fancy text-xs tracking-widest uppercase text-font-dimest">Empires</span>
        <div class="h-px flex-1 bg-card-edge"></div>
    </div>
    <div class="flex flex-col gap-4">
        {#each game.players as player}
            {@const founded = foundedReligion(player)}
            {@const hasStats = yieldCols.some((c) => (player as any)[c.key] != null)}
            <div class="rounded-2xl border border-card-edge bg-card shadow-md shadow-darken">
                <!-- Player group header -->
                <div class="flex items-center gap-3 px-5 py-3 border-b border-card-edge bg-zebra-2 rounded-t-2xl">
                    <div class="h-10 w-10 rounded-full bg-card-edge overflow-hidden shrink-0">
                        {#if leaderPortrait(player.leader)}
                            <img src={leaderPortrait(player.leader)!} alt=""
                                 class="h-full w-full object-cover"
                                 onerror={(e) => ((e.currentTarget as HTMLImageElement).style.display = 'none')} />
                        {:else if player.leader}
                            <div class="h-full w-full flex items-center justify-center text-font-dimest text-[10px] font-bold select-none">?</div>
                        {/if}
                    </div>
                    <div class="min-w-0 flex-1">
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
                    {#if founded}
                        <span class="hidden items-center gap-1 max-w-[16rem] shrink-0 sm:flex" title={founded.name ? `Founded ${founded.name}` : 'Founded a religion'}>
                            {#if founded.icon}
                                <span class="inline-block h-7 w-7 shrink-0"
                                      style="background-color:{founded.color};-webkit-mask:url({founded.icon}) center/contain no-repeat;mask:url({founded.icon}) center/contain no-repeat;"></span>
                            {/if}
                            <span class="truncate text-sm italic" style="color:{founded.color};">{founded.name || '—'}</span>
                        </span>
                    {/if}
                    <span class="font-fancy text-[10px] font-semibold tracking-widest uppercase text-font-dimest shrink-0 {founded ? 'ml-5' : ''}">
                        {player.cities.length} {player.cities.length === 1 ? 'city' : 'cities'}
                    </span>
                </div>

                {#if hasStats}
                    <div class="flex flex-wrap items-center gap-x-5 gap-y-2 px-5 py-3 {player.cities.length > 0 ? 'border-b border-card-edge' : ''}">
                        {#each yieldCols as col}
                            {@const v = (player as any)[col.key]}
                            {@const lead = v != null && v === yieldLeaders[col.key]}
                            <span class="flex items-center gap-1.5" title={col.label}>
                                <img src={col.icon} alt={col.label} class="{col.h}" />
                                <span class="{col.cls} text-sm font-bold tabular-nums transition-opacity duration-100 {v == null ? 'opacity-30' : lead ? 'opacity-100' : 'opacity-55'}">{fmt(v)}</span>
                            </span>
                        {/each}
                    </div>
                {/if}

                {#if player.cities.length > 0}
                    <!-- City cards. Glanceable info up front; full yields on hover. -->
                    <div class="grid gap-2.5 p-3" style="grid-template-columns: repeat(auto-fit, minmax(15rem, 1fr));">
                        {#each player.cities as city}
                            {@const cname = normaliseCityName(city.name)}
                            {@const relIcon = cityReligionIcon(city)}
                            {@const relColor = city.religion_color ?? 'var(--color-font-dimer)'}
                            {@const relName = religionName(city)}
                            <Tooltip class="w-full" placement="top">
                                {#snippet children()}
                                    <div class="flex h-full w-full items-center gap-3 rounded-xl border border-card-edge bg-card-2 px-3 py-2 transition-colors duration-100 hover:border-card-edge-2 hover:bg-select">
                                        <div class="h-9 w-9 shrink-0 rounded-full bg-select p-1.5">
                                            <img src={cityIcon} alt="" class="h-full w-full" />
                                        </div>
                                        <div class="min-w-0 flex-1">
                                            <div class="flex items-center gap-1.5">
                                                <span class="truncate text-sm font-medium text-font-dim">{cname}</span>
                                                {#if relIcon}
                                                    <span class="inline-block h-[23px] w-[23px] shrink-0" title={relName}
                                                          style="background-color:{relColor};-webkit-mask:url({relIcon}) center/contain no-repeat;mask:url({relIcon}) center/contain no-repeat;"></span>
                                                {/if}
                                            </div>
                                            <div class="mt-0.5 flex min-h-[1.125rem] items-center gap-2">
                                                <span class="flex shrink-0 items-center gap-1 text-xs text-font-dimer">
                                                    <img src={populationIcon} alt="pop" class="h-3.5" />
                                                    <span class="tabular-nums font-semibold">{fmt(city.population)}</span>
                                                </span>
                                                {#if city.wonders?.length}
                                                    <span class="ml-auto flex shrink-0 items-center gap-0.5">
                                                        {#each city.wonders as wonder}
                                                            {#if wonderIcon(wonder)}
                                                                <img src={wonderIcon(wonder)!} alt={wonder} class="h-[18px] shrink-0 object-contain" />
                                                            {:else}
                                                                <span class="rounded border border-card-edge px-1 text-[10px] text-font-dimest">W</span>
                                                            {/if}
                                                        {/each}
                                                    </span>
                                                {/if}
                                            </div>
                                        </div>
                                    </div>
                                {/snippet}
                                {#snippet content()}
                                    <div class="flex flex-col gap-2 px-0.5 py-0.5">
                                        <div class="grid grid-cols-2 gap-x-4 gap-y-1">
                                            {#each cityYieldCols as col}
                                                {@const raw = (city as any)[col.key]}
                                                <span class="flex items-center justify-between gap-3">
                                                    <span class="flex items-center gap-1.5 text-font-dimer">
                                                        <img src={col.icon} alt="" class="h-3.5" />{col.label}
                                                    </span>
                                                    <span class="{col.cls} font-bold tabular-nums">{fmt(raw == null ? null : Math.round(raw))}</span>
                                                </span>
                                            {/each}
                                        </div>
                                        {#if relName || relIcon}
                                            <div class="flex items-center gap-1.5 border-t border-card-edge pt-1.5 text-[11px]">
                                                {#if relIcon}
                                                    <span class="inline-block h-[22px] w-[22px] shrink-0"
                                                          style="background-color:{relColor};-webkit-mask:url({relIcon}) center/contain no-repeat;mask:url({relIcon}) center/contain no-repeat;"></span>
                                                {/if}
                                                <span class="text-font-dimest">Religion</span>
                                                <span class="ml-auto italic" style="color:{relColor};">{relName || '—'}</span>
                                            </div>
                                        {/if}
                                        {#if city.wonders?.length}
                                            <div class="flex flex-wrap items-center gap-1 border-t border-card-edge pt-1.5">
                                                {#each city.wonders as wonder}
                                                    {#if wonderIcon(wonder)}
                                                        <span class="flex items-center gap-1 text-[10px] text-font-dimer">
                                                            <img src={wonderIcon(wonder)!} alt="" class="h-4 object-contain" />{wonder}
                                                        </span>
                                                    {/if}
                                                {/each}
                                            </div>
                                        {/if}
                                    </div>
                                {/snippet}
                            </Tooltip>
                        {/each}
                    </div>
                {/if}
            </div>
        {/each}
    </div>
    {/if}
</div>
