<script lang="ts">
    import type { PageData } from './$types';
    import { Map, Trash2, Save, Trophy, RefreshCw } from '@lucide/svelte';

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
    import foodIcon       from '$lib/assets/icons/yields/food.png';
    import productionIcon from '$lib/assets/icons/yields/production.png';
    import populationIcon from '$lib/assets/icons/emblems/population.png';

    let { data, form }: { data: PageData; form: any } = $props();
    const { game, rows, players } = $derived(data);
    // Every game reaching this page is tmp=true (server redirects finalized games).

    const victoryTypes = [
        { value: 'Domination',   icon: dominationv },
        { value: 'Science',      icon: sciencev },
        { value: 'Culture',      icon: culturev },
        { value: 'Score',        icon: scorev },
        { value: 'Religious',    icon: religiousv },
        { value: 'Diplomatic',   icon: diplomaticv },
        { value: 'Capitulation', icon: capitulationv },
    ];

    const leaderAssets = import.meta.glob<{ default: string }>(
        '$lib/assets/icons/leaders/*.webp', { eager: true }
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

    function fmt(val: number | null) { return val != null ? String(val) : '—'; }

    let selectedVictory = $state('');
    let winnerRowId     = $state<number | null>(null);
    let assignments     = $state<Record<number, number>>({});
    let updateFile      = $state<File | null>(null);
    let updateInputEl   = $state<HTMLInputElement>();

    const allAssigned  = $derived(rows.every((r: any) => (assignments[r.id] ?? 0) !== 0));
    const canSave      = $derived(allAssigned);
    const canFinalize  = $derived(allAssigned && !!selectedVictory && winnerRowId !== null);

    function normaliseLabel(v: string | null | undefined): string {
        if (!v) return '—';
        let s = v;
        for (const p of ['MAPSIZE_', 'GAMESPEED_'])
            if (s.startsWith(p)) { s = s.slice(p.length); break; }
        if (s !== s.toUpperCase()) return s;
        return s.split('_').map(w => w ? w[0].toUpperCase() + w.slice(1).toLowerCase() : '').join(' ');
    }
</script>

<div class="mx-12 mb-12 flex flex-col gap-3">

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

    <!-- ── Header: map + game info ──────────────────────────────────────── -->
    <div class="flex gap-3">

        <!-- Map -->
        <div class="w-72 shrink-0 rounded-2xl border border-card-edge bg-card shadow-md shadow-darken overflow-hidden">
            {#if game.has_map}
                <img src="/files/maps/{game.id}" alt="" class="w-full h-full object-cover" />
            {:else}
                <div class="flex h-full min-h-44 flex-col items-center justify-center gap-2 text-font-dimest">
                    <Map strokeWidth={1} class="h-8 w-8" />
                    <span class="text-sm">Map not available</span>
                </div>
            {/if}
        </div>

        <!-- Info + victory picker -->
        <div class="flex-1 rounded-2xl border border-card-edge bg-card p-6 shadow-md shadow-darken flex flex-col justify-between gap-4">
            <div class="flex flex-col gap-1">
                <div class="flex items-center gap-2">
                    <span class="font-fancy text-xl font-semibold">
                        Confirm Game
                    </span>
                </div>
                <div class="text-font-dimer text-sm flex flex-wrap gap-x-3 gap-y-1 mt-1">
                    {#if game.map}<span>{game.map}</span>{/if}
                    {#if game.map_size}<span>{normaliseLabel(game.map_size)}</span>{/if}
                    {#if game.game_speed}<span>{normaliseLabel(game.game_speed)}</span>{/if}
                    {#if game.turns}<span>Turn {game.turns}</span>{/if}
                </div>
            </div>

            <!-- Victory type (only needed for Finalize) -->
            <div>
                <p class="text-xs text-font-dimest mb-2 font-fancy tracking-wide uppercase">
                    Victory type <span class="text-font-dimest/60">(required to finalize)</span>
                </p>
                <div class="flex gap-2">
                    {#each victoryTypes as vt}
                        <label class="cursor-pointer">
                            <input type="radio" name="victory_type_display" value={vt.value}
                                class="sr-only" bind:group={selectedVictory} />
                            <img src={vt.icon} alt={vt.value} title={vt.value}
                                class="h-10 w-10 rounded-full transition-all duration-150
                                       {selectedVictory === vt.value
                                         ? 'opacity-100'
                                         : 'opacity-30 hover:opacity-65'}" />
                        </label>
                    {/each}
                </div>
            </div>
        </div>
    </div>

    <!-- ── Player table ──────────────────────────────────────────────────── -->
    <form method="POST" action="?/confirm" id="main-form"
          class="rounded-2xl border border-card-edge bg-card shadow-md shadow-darken overflow-hidden">

        <!-- Hidden victory type for submission -->
        <input type="hidden" name="victory_type" value={selectedVictory} />

        <table class="w-full">
            <thead>
                <tr class="border-b border-card-edge">
                    <th class="text-left pl-5 py-2.5 font-fancy text-[10px] tracking-widest uppercase text-font-dimest w-52">Leader</th>
                    <th class="text-left px-3 py-2.5 font-fancy text-[10px] tracking-widest uppercase text-font-dimest">Pseudo</th>
                    <th class="text-right pr-3 py-2.5 font-fancy text-[10px] tracking-widest uppercase text-font-dimest">Score</th>
                    <th class="text-right pr-3 py-2.5 font-fancy text-[10px] tracking-widest uppercase text-font-dimest">Science</th>
                    <th class="text-right pr-3 py-2.5 font-fancy text-[10px] tracking-widest uppercase text-font-dimest">Culture</th>
                    <th class="text-right pr-3 py-2.5 font-fancy text-[10px] tracking-widest uppercase text-font-dimest">Gold</th>
                    <th class="text-right pr-3 py-2.5 font-fancy text-[10px] tracking-widest uppercase text-font-dimest">Faith</th>
                    <th class="text-left px-4 py-2.5 font-fancy text-[10px] tracking-widest uppercase text-font-dimest">Assign to</th>
                    <th class="text-center pr-4 py-2.5 font-fancy text-[10px] tracking-widest uppercase text-font-dimest">Winner</th>
                </tr>
            </thead>
            <tbody>
                {#each rows as row}
                    <tr class="not-last:border-b border-card-edge hover:bg-select transition-colors duration-100">
                        <!-- Leader portrait + name -->
                        <td class="pl-4 py-3">
                            <div class="flex items-center gap-2.5">
                                <div class="h-9 w-9 rounded-full bg-card-edge overflow-hidden shrink-0">
                                    {#if leaderPortrait(row.leader)}
                                        <img src={leaderPortrait(row.leader)} alt=""
                                             class="h-full w-full object-cover"
                                             onerror={(e) => ((e.currentTarget as HTMLImageElement).style.display='none')} />
                                    {/if}
                                </div>
                                <span class="text-sm text-font-dim">{row.leader ?? '—'}</span>
                            </div>
                        </td>
                        <td class="px-3 py-3 text-sm text-font-dimer">{row.pseudo_name ?? '—'}</td>
                        <!-- Yields -->
                        <td class="pr-3 py-3"><span class="flex items-center justify-end gap-1"><img src={scoreIcon}   alt="" class="h-4.5 shrink-0" /><span class="text-score   text-sm font-bold tabular-nums">{fmt(row.score)}</span></span></td>
                        <td class="pr-3 py-3"><span class="flex items-center justify-end gap-1"><img src={scienceIcon} alt="" class="h-5   shrink-0" /><span class="text-science text-sm font-bold tabular-nums">{fmt(row.science)}</span></span></td>
                        <td class="pr-3 py-3"><span class="flex items-center justify-end gap-1"><img src={cultureIcon} alt="" class="h-5   shrink-0" /><span class="text-culture text-sm font-bold tabular-nums">{fmt(row.culture)}</span></span></td>
                        <td class="pr-3 py-3"><span class="flex items-center justify-end gap-1"><img src={goldIcon}    alt="" class="h-5   shrink-0" /><span class="text-gold    text-sm font-bold tabular-nums">{fmt(row.gold)}</span></span></td>
                        <td class="pr-3 py-3"><span class="flex items-center justify-end gap-1"><img src={faithIcon}   alt="" class="h-5   shrink-0" /><span class="text-faith   text-sm font-bold tabular-nums">{fmt(row.faith)}</span></span></td>
                        <!-- Assign + Winner -->
                        <td class="px-4 py-3">
                            <select name="row_{row.id}"
                                class="rounded-lg bg-card border border-card-edge px-2 py-1 text-sm outline-none w-36 text-font-dim"
                                onchange={(e) => assignments[row.id] = parseInt((e.currentTarget as HTMLSelectElement).value)}>
                                <option value="0">Select Player...</option>
                                {#each players as p}
                                    {@const takenByOther = Object.entries(assignments).some(
                                        ([rid, pid]) => pid === p.id && parseInt(rid) !== row.id
                                    )}
                                    <option value={p.id} disabled={takenByOther}>{p.name}</option>
                                {/each}
                            </select>
                        </td>
                        <td class="pr-4 py-3 text-center">
                            <input type="checkbox" name="winner_{row.id}"
                                class="confirm-check"
                                checked={winnerRowId === row.id}
                                onchange={(e) => {
                                    if ((e.currentTarget as HTMLInputElement).checked) winnerRowId = row.id;
                                    else if (winnerRowId === row.id) winnerRowId = null;
                                }} />
                        </td>
                    </tr>
                {/each}
            </tbody>
        </table>
    </form>

    <!-- ── Update save (ongoing games only) ─────────────────────────────── -->
    {#if true}
        <form method="POST" action="?/update" enctype="multipart/form-data"
              class="rounded-2xl border border-card-edge bg-card p-5 shadow-md shadow-darken">
            <p class="font-fancy text-xs tracking-widest uppercase text-font-dimest mb-3">Update Save</p>
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
    {/if}

    <!-- ── Actions ───────────────────────────────────────────────────────── -->
    <div class="flex justify-between items-center gap-3">

        <!-- Destructive left side -->
        <button type="submit" form="main-form" formaction="?/cancel"
            class="flex items-center gap-1.5 rounded-lg border border-card-edge px-3 py-2 text-sm text-font-dimer
                   hover:border-font-bad/40 hover:text-font-bad transition-colors duration-150 cursor-pointer">
            <Trash2 class="h-3.5 w-3.5" strokeWidth={1.5} /> Delete
        </button>

        <!-- Primary actions right side -->
        <div class="flex gap-2">
            <a href="/matches"
                class="flex items-center gap-1.5 rounded-lg border border-card-edge px-4 py-2 text-sm font-semibold text-font-dim hover:bg-select transition-colors duration-150">
                <Save class="h-3.5 w-3.5" strokeWidth={1.5} /> Save Progress
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
        background: var(--color-primary);
        border-color: var(--color-primary);
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
