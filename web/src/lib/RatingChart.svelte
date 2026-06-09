<script lang="ts">
    import { onMount, onDestroy } from 'svelte';
    import type uPlotType from 'uplot';

    export interface HistoryPoint {
        date: string;
        pre_rating_overall: number | string;
        post_rating_overall: number | string;
        winner: boolean;
    }

    let { history }: { history: HistoryPoint[] } = $props();

    // ── Period selector ──────────────────────────────────────────────────────
    const PERIODS = ['1M', '3M', '6M', '1Y', 'All'] as const;
    type Period = (typeof PERIODS)[number];
    let period = $state<Period>('All');

    // ── DOM refs ─────────────────────────────────────────────────────────────
    let wrap: HTMLDivElement | undefined;

    // ── Tooltip state ────────────────────────────────────────────────────────
    let tip = $state({
        visible: false,
        x:       0,
        rating:  0,
        delta:   0,
        date:    '',
        win:     null as boolean | null,
    });

    // ── Chart instance ───────────────────────────────────────────────────────
    let uP: typeof uPlotType | null = null;
    let chart: uPlotType | null = null;

    // ── Data helpers ─────────────────────────────────────────────────────────
    function getCutoff(p: Period): number {
        if (p === 'All') return 0;
        const days: Record<string, number> = { '1M': 30, '3M': 90, '6M': 180, '1Y': 365 };
        return Date.now() - days[p] * 86_400_000;
    }

    interface Series {
        ts:    number[];
        rating: number[];
        delta: number[];
        wins:  (boolean | null)[];
    }

    function buildSeries(p: Period): Series | null {
        const from     = getCutoff(p);
        const filtered = history.filter(h => new Date(h.date).getTime() >= from);
        if (filtered.length < 1) return null;

        const ts:     number[]            = [];
        const rating: number[]            = [];
        const delta:  number[]            = [];
        const wins:   (boolean | null)[]  = [];

        // Anchor: pre-rating of first game
        ts.push(new Date(filtered[0].date).getTime() / 1000 - 1);
        rating.push(Number(filtered[0].pre_rating_overall));
        delta.push(0);
        wins.push(null);

        for (const h of filtered) {
            const pre  = Number(h.pre_rating_overall);
            const post = Number(h.post_rating_overall);
            ts.push(new Date(h.date).getTime() / 1000);
            rating.push(post);
            delta.push(Math.round(post - pre));
            wins.push(h.winner);
        }

        return { ts, rating, delta, wins };
    }

    // ── Chart lifecycle ───────────────────────────────────────────────────────
    function rebuild() {
        if (!uP || !wrap) return;
        chart?.destroy();
        chart = null;
        tip = { ...tip, visible: false };

        const s = buildSeries(period);
        if (!s || s.ts.length < 2) return;

        const { ts, rating, delta, wins } = s;
        const W = wrap.getBoundingClientRect().width || wrap.clientWidth;
        if (W < 1) return;

        chart = new uP!(
            {
                width:   W,
                height:  180,
                padding: [12, 0, 0, 0],
                cursor: {
                    drag: { x: false, y: false },
                    points: {
                        size:   ((_u: any, _si: any) => 10) as any,
                        fill:   ((u: any, _si: any) => {
                            const w = wins[u.cursor.idx];
                            return w === null ? 'transparent'
                                 : w          ? '#6ab355'
                                             : '#ef5a34';
                        }) as any,
                        stroke: (() => '#08080a') as any,
                        width:  2,
                    },
                },
                hooks: {
                    setCursor: [
                        (u) => {
                            const idx = u.cursor.idx;
                            if (idx == null) {
                                tip = { ...tip, visible: false };
                                return;
                            }
                            const d = new Date((u.data[0][idx] as number) * 1000);
                            tip = {
                                visible: true,
                                // cursor.left is px from left edge of the u-over canvas
                                x:       Math.round(u.cursor.left ?? 0),
                                rating:  Math.round(u.data[1][idx] as number),
                                delta:   delta[idx] ?? 0,
                                date:    d.toLocaleDateString('en-GB', {
                                    day: 'numeric', month: 'short', year: 'numeric',
                                }),
                                win: wins[idx] ?? null,
                            };
                        },
                    ],
                },
                scales: {
                    x: { time: true },
                    y: {},
                },
                axes: [
                    {
                        // x-axis
                        stroke: '#555',
                        grid:  { stroke: '#1e1e1e', width: 1 },
                        ticks: { stroke: '#1e1e1e', width: 1, size: 4 },
                        font:  '11px Quicksand, sans-serif',
                        gap:   6,
                    },
                    {
                        // y-axis on the right
                        side:   1,
                        stroke: '#555',
                        grid:  { stroke: '#1e1e1e', width: 1 },
                        ticks: { stroke: '#1e1e1e', width: 1, size: 4 },
                        font:  '11px Quicksand, sans-serif',
                        gap:   6,
                        size:  48,
                        values: (_u, vals) => vals.map(v => v == null ? '' : String(Math.round(v as number))),
                    },
                ],
                series: [
                    {},
                    {
                        stroke: '#e5c96b',
                        fill:   'rgba(229,201,107,0.055)',
                        width:  2,
                        points: { show: false },
                    },
                ],
                legend: { show: false },
            },
            [ts, rating],
            wrap,
        );
    }

    // ── Period button handler ─────────────────────────────────────────────────
    function setPeriod(p: Period) {
        period = p;
        rebuild();
    }

    // ── Mount / teardown ──────────────────────────────────────────────────────
    onMount(() => {
        // Dynamic import keeps uplot out of the SSR bundle
        Promise.all([
            import('uplot'),
            import('uplot/dist/uPlot.min.css'),
        ]).then(([mod]) => {
            uP = mod.default;
            rebuild();
        });

        const ro = new ResizeObserver(rebuild);
        ro.observe(wrap!);
        return () => { ro.disconnect(); chart?.destroy(); };
    });

    onDestroy(() => chart?.destroy());
</script>

<!-- ── Period selector (flow, not absolute – gives the chart breathing room) -->
<div class="flex justify-end gap-1 mb-2">
    {#each PERIODS as p}
        <button
            onclick={() => setPeriod(p)}
            class="px-2 py-0.5 rounded text-xs font-semibold font-fancy tracking-wide transition-colors duration-150 cursor-pointer
                   {period === p
                       ? 'bg-primary/15 text-primary border border-primary/25'
                       : 'text-font-dimest hover:text-font-dim border border-transparent'}"
        >{p}</button>
    {/each}
</div>

<!-- ── Chart + tooltip (relative anchor for tooltip positioning) ─────────── -->
<div class="relative">

    <!-- Chart mount point -->
    <!-- svelte-ignore a11y_no_static_element_interactions -->
    <div
        bind:this={wrap}
        class="w-full"
        onmouseleave={() => (tip = { ...tip, visible: false })}
    >
        <!-- Tooltip -->
        {#if tip.visible}
            <div
                class="absolute z-20 pointer-events-none
                       bg-card border border-card-edge rounded-xl px-3 py-2 shadow-lg shadow-darken"
                style="left: {tip.x}px; bottom: calc(100% - 160px); transform: translateX(-50%);"
            >
                <div class="flex items-baseline gap-1.5">
                    <span class="text-base font-bold text-font-clear leading-none">{tip.rating}</span>
                    {#if tip.win !== null}
                        <span class="text-xs font-semibold leading-none
                                     {tip.delta > 0 ? 'text-font-good' : tip.delta < 0 ? 'text-font-bad' : 'text-font-dimer'}">
                            {tip.delta > 0 ? '+' : ''}{tip.delta}
                        </span>
                    {/if}
                </div>
                <div class="text-[11px] text-font-dimer mt-1 whitespace-nowrap">{tip.date}</div>
                {#if tip.win !== null}
                    <div class="text-[10px] mt-0.5 font-semibold {tip.win ? 'text-font-good' : 'text-font-bad'}">
                        {tip.win ? '▲ Win' : '▼ Loss'}
                    </div>
                {/if}
            </div>
        {/if}
    </div>
</div>

<style>
    /*
     * Do NOT override .uplot or .u-wrap widths.
     * uplot sets .u-wrap to exact px via inline style; .uplot has width:min-content
     * which resolves correctly to that pixel value.  Any % override on .u-wrap
     * creates a circular min-content dependency that collapses the canvas to 0.
     * The ResizeObserver in onMount handles responsive resizing instead.
     */
    :global(.u-cursor-x) { border-right: 1px dashed rgba(255,255,255,0.10) !important; }
    :global(.u-cursor-y) { display: none !important; }
    :global(.u-select)   { background: rgba(255,255,255,0.03) !important; }
</style>
