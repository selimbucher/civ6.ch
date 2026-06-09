<script lang="ts">
    import { Star } from '@lucide/svelte';

    let { data } = $props();

    const difficultyLabel: Record<number, string> = {
        [-1]: 'Secret', 0: 'Common', 1: 'Uncommon', 2: 'Rare', 3: 'Epic',
    };
    const difficultyColor: Record<number, string> = {
        [-1]: 'text-secondary border-secondary/30 bg-secondary/10',
        0:    'text-font-dimer border-card-edge bg-card-2',
        1:    'text-font-good border-font-good/30 bg-font-good/10',
        2:    'text-primary border-primary/30 bg-primary/15',
        3:    'text-font-bad border-font-bad/30 bg-font-bad/10',
    };
    const difficultyGlow: Record<number, string> = {
        [-1]: 'shadow-[0_0_18px_0px_var(--color-secondary)] shadow-secondary/20',
        0:    '',
        1:    '',
        2:    'shadow-[0_0_14px_0px_var(--color-primary-shadow)]',
        3:    'shadow-[0_0_14px_0px_hsl(10,83%,57%,0.25)]',
    };

    const total = $derived(data.achievements.filter((a: any) => !a.disabled).length);
</script>

<div class="mx-4 md:mx-12 mb-12 flex flex-col gap-8">

    <!-- ── Player scoreboard ────────────────────────────────────────────── -->
    <div class="rounded-2xl border border-card-edge bg-card shadow-md shadow-darken overflow-hidden">
        <div class="px-6 py-4 border-b border-card-edge flex items-center justify-between">
            <span class="font-fancy text-xs font-semibold tracking-widest uppercase text-font-dimer">Achievement Points</span>
        </div>
        <!-- Desktop: original columns layout -->
        <div class="hidden md:flex divide-x divide-card-edge">
            {#each data.players as player, i}
                {@const bits = BigInt(player.achievement_bitstring ?? 0)}
                {@const earned = data.achievements.filter((a: any) =>
                    !a.disabled && ((bits >> BigInt(a.id - 1)) & 1n) === 1n
                ).length}
                <a href="/profile/{player.id}"
                   class="flex-1 flex flex-col items-center gap-1.5 px-6 py-4 hover:bg-select transition-colors duration-150">
                    <div class="flex flex-wrap items-center justify-center gap-1.5">
                        {#if i === 0}
                            <Star class="h-3.5 w-3.5 text-primary shrink-0" fill="currentColor" strokeWidth={0} />
                        {/if}
                        <span class="font-semibold text-font-clear text-center">{player.name}</span>
                    </div>
                    <span class="font-fancy text-2xl font-bold text-primary">{player.achievement_points}</span>
                    <span class="text-xs text-font-dimest">{earned} / {total}</span>
                </a>
            {/each}
        </div>

        <!-- Mobile: row list -->
        <div class="md:hidden flex flex-col">
            {#each data.players as player, i}
                {@const bits = BigInt(player.achievement_bitstring ?? 0)}
                {@const earned = data.achievements.filter((a: any) =>
                    !a.disabled && ((bits >> BigInt(a.id - 1)) & 1n) === 1n
                ).length}
                <a href="/profile/{player.id}"
                   class="flex items-center gap-3 px-5 py-3 not-last:border-b border-card-edge hover:bg-select transition-colors duration-150">
                    <span class="w-4 shrink-0 flex justify-center">
                        {#if i === 0}
                            <Star class="h-3.5 w-3.5 text-primary" fill="currentColor" strokeWidth={0} />
                        {:else}
                            <span class="text-xs text-font-dimest">{i + 1}</span>
                        {/if}
                    </span>
                    <span class="flex-1 font-semibold text-font-clear text-sm truncate">{player.name}</span>
                    <span class="text-xs text-font-dimest shrink-0">{earned}/{total}</span>
                    <span class="font-fancy text-xl font-bold text-primary tabular-nums shrink-0 w-8 text-right">{player.achievement_points}</span>
                </a>
            {/each}
        </div>
    </div>

    <!-- ── Achievement grid ─────────────────────────────────────────────── -->
    {#each [-1, 3, 2, 1, 0] as diff}
        {@const group = data.achievements.filter((a: any) => a.difficulty === diff && !a.disabled)}
        {#if group.length > 0}
            <div>
                <div class="flex items-center gap-4 mb-4">
                    <div class="h-px flex-1 bg-card-edge"></div>
                    <span class="font-fancy text-xs tracking-widest uppercase {difficultyColor[diff].split(' ')[0]}">
                        {difficultyLabel[diff]}
                    </span>
                    <div class="h-px flex-1 bg-card-edge"></div>
                </div>

                <div class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-3">
                    {#each group as ach}
                        {@const anyEarner = ach.earners.length > 0}
                        <div class="rounded-2xl border border-card-edge bg-card p-5 shadow-md shadow-darken flex flex-col gap-3
                                    {anyEarner ? difficultyGlow[diff] : 'opacity-60'}">
                            <div class="flex items-start justify-between gap-3">
                                <h3 class="font-fancy font-semibold text-font-clear leading-tight">{ach.name}</h3>
                                <span class="shrink-0 text-[10px] font-semibold px-2 py-0.5 rounded-full border
                                             {difficultyColor[diff]}">
                                    {difficultyLabel[ach.difficulty]}
                                </span>
                            </div>

                            <p class="text-sm text-font-dimer leading-relaxed flex-1">{ach.description}</p>

                            <!-- Earners — clicking the name goes to the game that triggered the achievement -->
                            {#if ach.earners.length > 0}
                                <div class="flex flex-wrap gap-1.5 pt-1 border-t border-card-edge">
                                    {#each ach.earners as earner}
                                        <a href={earner.game_id ? `/matches/view/${earner.game_id}` : `/profile/${earner.player_id}`}
                                           class="text-xs px-2 py-0.5 rounded-full bg-select text-font-dim
                                                  hover:text-font-clear hover:bg-card-3 transition-colors duration-150">
                                            {earner.player_name}
                                        </a>
                                    {/each}
                                </div>
                            {:else}
                                <div class="pt-1 border-t border-card-edge">
                                    <span class="text-xs text-font-dimest italic">Not yet earned</span>
                                </div>
                            {/if}
                        </div>
                    {/each}
                </div>
            </div>
        {/if}
    {/each}

    <!-- Retired -->
    {#if data.achievements.some((a: any) => a.disabled)}
        {@const disabled = data.achievements.filter((a: any) => a.disabled)}
        <div>
            <div class="flex items-center gap-4 mb-4">
                <div class="h-px flex-1 bg-card-edge"></div>
                <span class="font-fancy text-xs tracking-widest uppercase text-font-dimest">Retired</span>
                <div class="h-px flex-1 bg-card-edge"></div>
            </div>
            <div class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-3">
                {#each disabled as ach}
                    <div class="rounded-2xl border border-card-edge bg-card p-5 opacity-40 flex flex-col gap-3">
                        <div class="flex items-start justify-between gap-3">
                            <h3 class="font-fancy font-semibold text-font-dim leading-tight">{ach.name}</h3>
                            <span class="shrink-0 text-[10px] font-semibold px-2 py-0.5 rounded-full border border-card-edge text-font-dimest">
                                Retired
                            </span>
                        </div>
                        <p class="text-sm text-font-dimest leading-relaxed">{ach.description}</p>
                        {#if ach.earners.length > 0}
                            <div class="flex flex-wrap gap-1.5 pt-1 border-t border-card-edge">
                                {#each ach.earners as earner}
                                    <span class="text-xs px-2 py-0.5 rounded-full bg-select text-font-dimest">
                                        {earner.player_name}
                                    </span>
                                {/each}
                            </div>
                        {/if}
                    </div>
                {/each}
            </div>
        </div>
    {/if}

</div>
