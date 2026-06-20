<script lang="ts">
    import type { PageData } from './$types';
    import { Map, Trash2, Trophy, RefreshCw, ChevronLeft } from '@lucide/svelte';
    import Dropdown from '$lib/Dropdown.svelte';
    import Avatar from '$lib/Avatar.svelte';

    import dominationv   from '$lib/assets/icons/vcondition/domination.png';
    import religiousv    from '$lib/assets/icons/vcondition/religious.png';
    import diplomaticv   from '$lib/assets/icons/vcondition/diplomatic.png';
    import sciencev      from '$lib/assets/icons/vcondition/science.png';
    import culturev      from '$lib/assets/icons/vcondition/culture.png';
    import scorev        from '$lib/assets/icons/vcondition/score.png';
    import capitulationv from '$lib/assets/icons/vcondition/capitulation.png';

    import scoreIcon      from '$lib/assets/icons/yields/score_upscaled.png';
    import scienceIcon    from '$lib/assets/icons/yields/science.png';
    import cultureIcon    from '$lib/assets/icons/yields/culture.png';
    import faithIcon      from '$lib/assets/icons/yields/faith.png';
    import goldIcon       from '$lib/assets/icons/yields/gold.png';

    let { data, form }: { data: PageData; form: any } = $props();
    const { game, rows, players } = $derived(data);

    const victoryTypes = [
        { value: 'Domination',   icon: dominationv,   color: 'var(--color-army-2)'    },
        { value: 'Science',      icon: sciencev,       color: 'var(--color-science-2)' },
        { value: 'Culture',      icon: culturev,       color: 'var(--color-culture-2)' },
        { value: 'Score',        icon: scorev,         color: 'var(--color-score-2)'   },
        { value: 'Religious',    icon: religiousv,     color: 'var(--color-faith-2)'   },
        { value: 'Diplomatic',   icon: diplomaticv,    color: 'var(--color-diplo-2)'   },
        { value: 'Capitulation', icon: capitulationv,  color: 'var(--color-font-dimer)'},
    ];

    const categoryLabel: Record<string, string> = {
        overall: 'Overall', ffa: 'FFA', '1v1': '1v1', teams: 'Teams',
    };

    const leaderAssets = import.meta.glob<{ default: string }>(
        '$lib/assets/icons/leaders/*.webp', { eager: true }
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

    function fmt(val: number | null) { return val != null ? String(val) : '—'; }

    let selectedVictory = $state('');
    let winnerTeam      = $state<number | null>(null);
    // Pre-fill assignments from Steam-linked players (one player per slot).
    let assignments     = $state<Record<number, number>>(initialAssignments());

    function initialAssignments(): Record<number, number> {
        const out: Record<number, number> = {};
        const used = new Set<number>();
        for (const r of data.rows as any[]) {
            if (r.matched_player_id && !used.has(r.matched_player_id)) {
                out[r.id] = r.matched_player_id;
                used.add(r.matched_player_id);
            }
        }
        return out;
    }
    let updateFile      = $state<File | null>(null);
    let updateInputEl   = $state<HTMLInputElement>();

    // Group rows by their parsed shared-victory team id. Solo players (FFA)
    // each form a team of one; real teams hold several members.
    // (Map is shadowed by the lucide icon import, so group with a plain object.)
    const teams = $derived.by(() => {
        const order: number[] = [];
        const byTeam: Record<number, any[]> = {};
        for (const r of rows) {
            if (!byTeam[r.team]) { byTeam[r.team] = []; order.push(r.team); }
            byTeam[r.team].push(r);
        }
        return order.map((t) => byTeam[t]);
    });

    // True if some other row already claims this player (one player per slot).
    function takenByOther(pid: number, rowId: number): boolean {
        return Object.entries(assignments).some(
            ([rid, p]) => p === pid && parseInt(rid) !== rowId
        );
    }

    const allAssigned  = $derived(rows.every((r: any) => (assignments[r.id] ?? 0) !== 0));
    const canFinalize  = $derived(allAssigned && !!selectedVictory && winnerTeam !== null);

    const selectedVictoryColor = $derived(
        victoryTypes.find(v => v.value === selectedVictory)?.color ?? 'var(--color-primary)'
    );

    function normaliseLabel(v: string | null | undefined): string {
        if (!v) return '—';
        let s = v;
        for (const p of ['MAPSIZE_', 'GAMESPEED_'])
            if (s.startsWith(p)) { s = s.slice(p.length); break; }
        if (s !== s.toUpperCase()) return s;
        return s.split('_').map(w => w ? w[0].toUpperCase() + w.slice(1).toLowerCase() : '').join(' ');
    }
</script>

<div class="mx-3 md:mx-12 mb-12 flex flex-col gap-4">

    {#if form?.error}
        <div class="rounded-xl border border-font-bad/30 bg-font-bad/10 px-4 py-2 text-sm text-font-bad">
            {form.error}
        </div>
    {/if}
    {#if form?.updated}
        <div class="rounded-xl border border-font-good/30 bg-font-good/10 px-4 py-2 text-sm text-font-good">
            ✓ {form.updateMessage}
        </div>
    {/if}
    {#if form?.updateError}
        <div class="rounded-xl border border-font-bad/30 bg-font-bad/10 px-4 py-2 text-sm text-font-bad">
            {form.updateError}
        </div>
    {/if}

    <!-- ── Header row ──────────────────────────────────────────────────────── -->
    <div class="flex flex-col md:flex-row gap-4">

        <!-- Map -->
        <div class="shrink-0 rounded-2xl border border-card-edge bg-card shadow-md shadow-darken overflow-hidden">
            {#if game.has_map}
                <img src="/files/maps/{game.id}" alt="" class="h-52 w-auto block" />
            {:else}
                <div class="h-52 w-48 flex flex-col items-center justify-center gap-2 text-font-dimest">
                    <Map strokeWidth={1} class="h-8 w-8" />
                    <span class="text-sm">No map preview</span>
                </div>
            {/if}
        </div>

        <!-- Info + victory picker -->
        <div class="flex-1 rounded-2xl border border-card-edge bg-card shadow-md shadow-darken overflow-hidden flex flex-col">
            <div class="h-[3px] bg-gradient-primary shrink-0"></div>

            <div class="p-6 flex flex-col justify-between gap-5 flex-1">
                <!-- Title + meta -->
                <div class="flex flex-col gap-2">
                    <div class="flex items-center gap-2.5">
                        <span class="font-fancy text-xl font-semibold text-font-clear">Confirm Game</span>
                        {#if game.category}
                            <span class="font-fancy text-[10px] tracking-wider uppercase px-2 py-0.5 rounded-full bg-card-2 border border-card-edge text-font-dimer">
                                {categoryLabel[game.category] ?? game.category}
                            </span>
                        {/if}
                    </div>
                    <div class="flex flex-wrap items-center gap-1.5 mt-0.5">
                        {#if game.map}
                            <span class="text-xs text-font-dim bg-card-2 border border-card-edge rounded-full px-2.5 py-0.5">{game.map}</span>
                        {/if}
                        {#if game.map_size}
                            <span class="text-xs text-font-dimer bg-card-2 border border-card-edge rounded-full px-2.5 py-0.5">{normaliseLabel(game.map_size)}</span>
                        {/if}
                        {#if game.game_speed}
                            <span class="text-xs text-font-dimer bg-card-2 border border-card-edge rounded-full px-2.5 py-0.5">{normaliseLabel(game.game_speed)}</span>
                        {/if}
                        {#if game.turns}
                            <span class="text-xs text-font-dimer bg-card-2 border border-card-edge rounded-full px-2.5 py-0.5">Turn {game.turns}</span>
                        {/if}
                    </div>
                </div>

                <!-- Victory picker -->
                <div class="flex flex-col gap-2">
                    <div class="flex items-center justify-between">
                        <p class="text-xs text-font-dimest font-fancy tracking-wide uppercase">Victory Type</p>
                        {#if selectedVictory}
                            <span class="text-xs font-semibold font-fancy" style="color: {selectedVictoryColor}">{selectedVictory}</span>
                        {:else}
                            <span class="text-xs text-font-dimest/60 italic">required to finalize</span>
                        {/if}
                    </div>
                    <div class="flex gap-2">
                        {#each victoryTypes as vt}
                            <label class="cursor-pointer group">
                                <input type="radio" name="victory_type_display" value={vt.value}
                                    class="sr-only" bind:group={selectedVictory} />
                                <img src={vt.icon} alt={vt.value} title={vt.value}
                                    class="h-10 w-10 rounded-full transition-all duration-150
                                           {selectedVictory === vt.value
                                             ? 'opacity-100 scale-110'
                                             : 'opacity-30 hover:opacity-65'}" />
                            </label>
                        {/each}
                    </div>
                </div>
            </div>
        </div>
    </div>

    <!-- ── Player table ──────────────────────────────────────────────────────── -->
    <form method="POST" action="?/confirm" id="main-form"
          class="rounded-2xl border border-card-edge bg-card shadow-md shadow-darken overflow-x-auto">

        <input type="hidden" name="victory_type" value={selectedVictory} />

        <table class="w-full">
            <thead>
                <tr class="border-b border-card-edge">
                    <th class="text-left pl-5 py-2.5 font-fancy text-[10px] tracking-widest uppercase text-font-dimest">Leader</th>
                    <th class="text-left px-4 py-2.5 font-fancy text-[10px] tracking-widest uppercase text-font-dimest">Assign to</th>
                    <th class="text-right pr-3 py-2.5 font-fancy text-[10px] tracking-widest uppercase text-font-dimest">Score</th>
                    <th class="text-right pr-3 py-2.5 font-fancy text-[10px] tracking-widest uppercase text-font-dimest">Science</th>
                    <th class="text-right pr-3 py-2.5 font-fancy text-[10px] tracking-widest uppercase text-font-dimest">Culture</th>
                    <th class="text-right pr-3 py-2.5 font-fancy text-[10px] tracking-widest uppercase text-font-dimest">Gold</th>
                    <th class="text-right pr-3 py-2.5 font-fancy text-[10px] tracking-widest uppercase text-font-dimest">Faith</th>
                    <th class="text-center pr-5 py-2.5 font-fancy text-[10px] tracking-widest uppercase text-font-dimest">Winner</th>
                </tr>
            </thead>
            {#each teams as team}
                {@const teamId      = team[0].team}
                {@const isWinner    = winnerTeam === teamId}
                {@const isTeamGame  = team.length > 1}
                <tbody class="border-b border-card-edge last:border-b-0">
                    {#each team as row, ri}
                        <tr class="transition-colors duration-100
                                   {isWinner ? 'bg-font-good/5' : 'hover:bg-select'}"
                            style={isWinner
                                     ? 'box-shadow: inset 3px 0 0 #6ab355'
                                     : (isTeamGame ? 'box-shadow: inset 3px 0 0 var(--color-card-edge-2)' : '')}>

                            <!-- Leader portrait + name + pseudo -->
                            <td class="pl-4 py-3">
                                <div class="flex items-center gap-2.5">
                                    <div class="h-9 w-9 rounded-full bg-card-edge overflow-hidden shrink-0 transition-all duration-150
                                                {isWinner ? 'ring-2 ring-font-good ring-offset-1 ring-offset-card' : ''}
                                                {row.eliminated ? 'grayscale opacity-60' : ''}">
                                        {#if leaderPortrait(row.leader)}
                                            <img src={leaderPortrait(row.leader)} alt=""
                                                 class="h-full w-full object-cover"
                                                 onerror={(e) => ((e.currentTarget as HTMLImageElement).style.display='none')} />
                                        {:else if row.leader}
                                            <div class="h-full w-full flex items-center justify-center text-font-dimest text-[9px] font-bold select-none">?</div>
                                        {/if}
                                    </div>
                                    <div class="flex flex-col leading-tight">
                                        <span class="text-sm transition-colors duration-150
                                                     {isWinner ? 'text-font-clear font-medium' : 'text-font-dim'}">
                                            {row.leader ?? '—'}
                                        </span>
                                        <div class="flex items-center gap-1.5 mt-0.5">
                                            {#if row.pseudo_name}
                                                <span class="text-xs text-font-dimest">{row.pseudo_name}</span>
                                            {/if}
                                            {#if row.eliminated}
                                                <span class="text-[9px] uppercase tracking-wider px-1.5 py-px rounded-full bg-font-bad/10 text-font-bad border border-font-bad/20">Eliminated</span>
                                            {/if}
                                        </div>
                                    </div>
                                </div>
                            </td>

                            <!-- Assign to -->
                            <td class="px-4 py-3">
                                <input type="hidden" name="row_{row.id}" value={assignments[row.id] ?? 0} />
                                <input type="hidden" name="winner_{row.id}" value={isWinner ? 'on' : ''} />
                                {#if row.matched_player_id}
                                    <!-- Recognised by Steam ID — locked to that player. -->
                                    <span class="inline-flex items-center gap-1.5 rounded-full border border-primary/25 bg-primary/10 py-1 pl-1 pr-2.5 text-xs font-medium text-primary">
                                        <Avatar id={row.matched_player_id} name={row.matched_player_name ?? '?'} avatar={row.matched_player_avatar}
                                            wrapClass="h-5 w-5 rounded-full bg-card-2 border border-card-edge shrink-0"
                                            letterClass="text-[9px] font-bold font-fancy text-primary select-none" />
                                        {row.matched_player_name ?? '—'}
                                    </span>
                                {:else}
                                    <div class="w-48">
                                        <Dropdown
                                            value={String(assignments[row.id] ?? '')}
                                            onChange={(v) => (assignments[row.id] = v ? parseInt(v) : 0)}
                                            items={players
                                                .filter((p: any) => !takenByOther(p.id, row.id))
                                                .map((p: any) => ({ value: String(p.id), label: p.name, avatarId: p.id, avatarValue: p.avatar }))}
                                            placeholder="Assign player…" />
                                    </div>
                                {/if}
                            </td>

                            <!-- Yields -->
                            <td class="pr-3 py-3"><span class="flex items-center justify-end gap-1"><img src={scoreIcon}   alt="" class="h-4.5 shrink-0" /><span class="text-score   text-sm font-bold tabular-nums">{fmt(row.score)}</span></span></td>
                            <td class="pr-3 py-3"><span class="flex items-center justify-end gap-1"><img src={scienceIcon} alt="" class="h-5   shrink-0" /><span class="text-science text-sm font-bold tabular-nums">{fmt(row.science)}</span></span></td>
                            <td class="pr-3 py-3"><span class="flex items-center justify-end gap-1"><img src={cultureIcon} alt="" class="h-5   shrink-0" /><span class="text-culture text-sm font-bold tabular-nums">{fmt(row.culture)}</span></span></td>
                            <td class="pr-3 py-3"><span class="flex items-center justify-end gap-1"><img src={goldIcon}    alt="" class="h-5   shrink-0" /><span class="text-gold    text-sm font-bold tabular-nums">{fmt(row.gold)}</span></span></td>
                            <td class="pr-3 py-3"><span class="flex items-center justify-end gap-1"><img src={faithIcon}   alt="" class="h-5   shrink-0" /><span class="text-faith   text-sm font-bold tabular-nums">{fmt(row.faith)}</span></span></td>

                            <!-- Winner toggle — one per team -->
                            {#if ri === 0}
                                <td rowspan={team.length} class="pr-5 py-3 text-center align-middle border-l border-card-edge/40">
                                    {#if isTeamGame}
                                        <div class="text-[9px] uppercase tracking-wider text-font-dimest mb-1">Team</div>
                                    {/if}
                                    <input type="checkbox"
                                        class="confirm-check"
                                        checked={isWinner}
                                        onchange={(e) => {
                                            winnerTeam = (e.currentTarget as HTMLInputElement).checked
                                                ? teamId
                                                : (winnerTeam === teamId ? null : winnerTeam);
                                        }} />
                                </td>
                            {/if}
                        </tr>
                    {/each}
                </tbody>
            {/each}
        </table>
    </form>

    <!-- ── Update save ──────────────────────────────────────────────────────── -->
    <form method="POST" action="?/update" enctype="multipart/form-data"
          class="rounded-2xl border border-card-edge bg-card p-5 shadow-md shadow-darken">
        <p class="font-fancy text-[10px] tracking-widest uppercase text-font-dimest mb-3">Update Save File</p>
        <div class="flex items-center gap-3">
            <label class="flex-1 flex items-center gap-3 rounded-xl border border-card-edge bg-card-2 px-4 py-2.5 cursor-pointer hover:bg-select transition-colors duration-150">
                <RefreshCw class="h-4 w-4 text-font-dimer shrink-0" strokeWidth={1.5} />
                <span class="text-sm text-font-dimer truncate">
                    {updateFile ? updateFile.name : 'Choose .Civ6Save file…'}
                </span>
                <input bind:this={updateInputEl} type="file" name="save" accept=".Civ6Save"
                    class="hidden"
                    onchange={(e) => { updateFile = (e.currentTarget as HTMLInputElement).files?.[0] ?? null; }} />
            </label>
            <button type="submit" disabled={!updateFile}
                class="shrink-0 rounded-lg px-4 py-2 text-sm font-semibold transition-all duration-150
                       {updateFile
                         ? 'bg-card border border-card-edge hover:border-primary/40 hover:text-primary cursor-pointer'
                         : 'text-font-dimest cursor-not-allowed border border-card-edge opacity-40'}">
                Upload
            </button>
        </div>
    </form>

    <!-- ── Actions ──────────────────────────────────────────────────────────── -->
    <div class="flex justify-between items-center">

        <button type="submit" form="main-form" formaction="?/cancel"
            class="flex items-center gap-1.5 rounded-lg border border-card-edge px-3 py-2 text-sm text-font-dimer
                   hover:border-font-bad/40 hover:text-font-bad transition-colors duration-150 cursor-pointer">
            <Trash2 class="h-3.5 w-3.5" strokeWidth={1.5} /> Delete
        </button>

        <div class="flex items-center gap-2">
            <a href="/matches"
                class="flex items-center gap-1.5 rounded-lg border border-card-edge px-4 py-2 text-sm text-font-dimer
                       hover:bg-select transition-colors duration-150">
                <ChevronLeft class="h-3.5 w-3.5" strokeWidth={1.5} /> Back
            </a>
            <button type="submit" form="main-form" formaction="?/confirm"
                disabled={!canFinalize}
                class="flex items-center gap-1.5 rounded-lg px-4 py-2 text-sm font-semibold transition-all duration-150
                       {canFinalize
                         ? 'bg-gradient-primary text-black cursor-pointer hover:brightness-125'
                         : 'bg-card border border-card-edge text-font-dimest cursor-not-allowed'}">
                <Trophy class="h-3.5 w-3.5" strokeWidth={1.5} /> Finalize
            </button>
        </div>
    </div>

</div>

<style>
    :global(.confirm-check) {
        appearance: none;
        -webkit-appearance: none;
        width: 1.1rem;
        height: 1.1rem;
        border-radius: 4px;
        border: 1.5px solid var(--color-card-edge-2);
        background: var(--color-card);
        cursor: pointer;
        position: relative;
        transition: background 150ms, border-color 150ms;
        display: inline-flex;
        align-items: center;
        justify-content: center;
        vertical-align: middle;
        outline: none;
    }
    :global(.confirm-check:focus),
    :global(.confirm-check:focus-visible) {
        outline: none;
        box-shadow: none;
    }
    :global(.confirm-check:checked) {
        background: var(--color-font-good);
        border-color: var(--color-font-good);
    }
    :global(.confirm-check:checked::after) {
        content: '';
        position: absolute;
        left: 50%;
        top: 46%;
        width: 32%;
        height: 58%;
        border: 2px solid #000;
        border-top: none;
        border-left: none;
        transform: translate(-50%, -50%) rotate(45deg);
    }
    :global(.confirm-check:disabled) {
        opacity: 0.25;
        cursor: not-allowed;
    }
</style>
