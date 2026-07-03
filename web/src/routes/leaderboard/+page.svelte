<script lang="ts">
  import { goto } from '$app/navigation';
  import { Flame, Medal, Trophy } from '@lucide/svelte';
  import Avatar from '$lib/Avatar.svelte';
  import PageHeader from '$lib/PageHeader.svelte';
  import Tooltip from '$lib/Tooltip.svelte';

  const tabs = ['FFA', 'Teams', '1v1'];
  const categoryMap: Record<string, string> = {
    'FFA': 'ffa', 'Teams': 'teams', '1v1': '1v1',
  };

  let { data } = $props();
  let active = $state(tabs.find(t => categoryMap[t] === data.category) ?? 'FFA');

  const denouncedSet = $derived(new Set<number>(data.denouncedIds ?? []));
  const winrate = (p: any) => (p.games > 0 ? Math.round(Number(p.wins) / Number(p.games) * 100) : 0);

  function switchTab(tab: string) {
    active = tab;
    goto(`?category=${categoryMap[tab]}`, { replaceState: true });
  }

  const rankStyle = (i: number) =>
    i === 0 ? 'text-[#FFD700]' :
    i === 1 ? 'text-[#B0B8C8]' :
    i === 2 ? 'text-[#CD7F32]' :
    'text-font-dimest';

  // order: 2nd left, 1st center, 3rd right
  const PODIUM = [
    { idx: 1, color: '#B0B8C8',              blockH: 56,  portraitPx: 64, letterCls: 'text-2xl' },
    { idx: 0, color: 'var(--color-primary)', blockH: 88,  portraitPx: 80, letterCls: 'text-3xl' },
    { idx: 2, color: '#CD7F32',              blockH: 36,  portraitPx: 52, letterCls: 'text-xl'  },
  ];
</script>

<div class="px-4 md:px-12 pb-12 pt-2 flex flex-col gap-6 w-full">

  <PageHeader title="Leaderboard" subtitle="The official pecking order — updated after every match, disputed immediately.">
    {#snippet icon()}
      <Trophy strokeWidth={1.5} class="h-10 w-10" />
    {/snippet}
  </PageHeader>

  <!-- ── Podium ───────────────────────────────────────────────────────── -->
  <div class="flex justify-center pt-2">
   <div class="flex items-end justify-center gap-6 border-b border-card-edge px-2">
    {#each PODIUM as { idx, color, blockH, portraitPx, letterCls }}
      {@const player = data.overall[idx]}
      {#if player}
        {@const wr = winrate(player)}
        <div class="flex flex-col items-center gap-2" style="width: 140px">
          <!-- Portrait -->
          <div class="rounded-full overflow-visible shrink-0 flex items-center justify-center bg-card-2"
               style="width:{portraitPx}px; height:{portraitPx}px;
                      box-shadow: 0 0 0 2px var(--color-card), 0 0 0 3.5px color-mix(in srgb, {color} 50%, transparent)">
            <Avatar id={player.id} name={player.name} avatar={player.avatar}
                    denounced={denouncedSet.has(player.id)}
                    wrapClass="w-full h-full rounded-full bg-card-2"
                    letterClass="text-primary {letterCls} font-bold font-fancy select-none" />
          </div>
          <!-- Name + rating + winrate -->
          <div class="text-center flex flex-col items-center">
            <a href="/profile/{player.id}"
               class="font-semibold text-font-dim hover:text-font-clear transition-colors duration-150 text-sm leading-tight">
              {player.name}
            </a>
            <div class="mt-0.5 text-sm text-font-dimer font-medium tabular-nums">{Math.round(Number(player.rating))}</div>
            {#if player.games > 0}
              <div class="mt-1 flex items-center gap-1.5">
                <div class="w-10 h-[3px] rounded-full bg-card-edge overflow-hidden">
                  <div class="h-full rounded-full {wr >= 60 ? 'bg-font-good' : wr >= 40 ? 'bg-primary' : 'bg-font-bad'}"
                       style="width:{wr}%"></div>
                </div>
                <span class="text-[10px] text-font-dimest tabular-nums">{wr}%</span>
              </div>
            {/if}
          </div>
          <!-- Step block -->
          <div class="w-full rounded-t-xl flex items-center justify-center"
               style="height:{blockH}px;
                      background: color-mix(in srgb, {color} 6%, transparent);
                      border: 1px solid color-mix(in srgb, {color} 18%, transparent);
                      border-bottom: none">
            <span class="font-fancy font-black text-2xl" style="color:{color}; opacity:0.65">{idx + 1}</span>
          </div>
        </div>
      {/if}
    {/each}
   </div>
  </div>

  <!-- ── Tables ───────────────────────────────────────────────────────── -->
  <div class="flex flex-col md:flex-row md:items-start gap-4">

  <!-- ── Overall (always visible, left) ──────────────────────────────── -->
  <div class="flex-1 rounded-2xl border border-card-edge bg-card shadow-md shadow-darken overflow-hidden">
    <table class="w-full">
      <thead>
        <tr class="border-b border-card-edge">
          <th class="font-fancy text-[10px] font-semibold tracking-widest uppercase text-font-dimest text-center px-4 py-2.5 w-10"></th>
          <th class="font-fancy text-[10px] font-semibold tracking-widest uppercase text-font-dimest text-left px-3 py-2.5">Player</th>
          <th class="font-fancy text-[10px] font-semibold tracking-widest uppercase text-font-dimest text-right px-3 py-2.5 w-24">Rating</th>
          <th class="font-fancy text-[10px] font-semibold tracking-widest uppercase text-font-dimest text-center px-4 py-2.5 w-16">
            <Tooltip label="Win rate across rated games" placement="bottom" align="right">
              {#snippet children()}<span class="cursor-help">WR</span>{/snippet}
            </Tooltip>
          </th>
        </tr>
      </thead>
      <tbody>
        {#each data.overall.slice(3) as player, i}
          {@const wr = player.games > 0 ? Math.round(Number(player.wins) / Number(player.games) * 100) : 0}
          <tr class="not-last:border-b border-card-edge odd:bg-zebra-2 transition-colors duration-100 group">
            <td class="text-center px-4 py-3">
              <span class="text-font-dimest text-sm flex justify-center items-center ml-3">{i + 4}</span>
            </td>
            <td class="px-3 py-3">
              <div class="flex items-center gap-3">
                <Avatar id={player.id} name={player.name} avatar={player.avatar}
                        denounced={denouncedSet.has(player.id)}
                        wrapClass="h-8 w-8 rounded-full bg-card-2 border border-card-edge shrink-0"
                        letterClass="text-primary text-[10px] font-bold font-fancy select-none" />
                <div class="flex flex-col gap-1">
                  <a href="/profile/{player.id}" class="flex items-center gap-2 w-fit">
                    <span class="tracking-wide text-font-dim group-hover:text-font-clear transition-colors duration-150">
                      {player.name}
                    </span>
                    {#if Number(player.streak) > 1}
                      <span class="flex items-center gap-0.5 text-xs font-semibold text-font-good bg-font-good/10 border border-font-good/20 px-1.5 py-0.5 rounded-full">
                        <Flame class="h-3 w-3" strokeWidth={2} />{player.streak}
                      </span>
                    {/if}
                  </a>
                </div>
              </div>
            </td>
            <td class="text-right px-3 py-3 flex flex-col">
              <span class="text-font-clear font-semibold tabular-nums">{Math.round(Number(player.rating))}</span>
              {#if Number(player.rd) >= 110}
                <span class="text-font-dimest text-xs ml-1">±{Math.round(Number(player.rd))}</span>
              {/if}
            </td>
            <td class="text-right px-4 py-3">
              <div class="flex flex-col items-end gap-1 mr-3 ">
                <span class="text-font-dim tabular-nums text-sm">{player.games > 0 ? wr + '%' : '—'}</span>
                {#if player.games > 0}
                  <div class="w-12 h-[3px] rounded-full bg-card-edge overflow-hidden">
                    <div class="h-full rounded-full transition-all duration-300
                                {wr >= 60 ? 'bg-font-good' : wr >= 40 ? 'bg-primary' : 'bg-font-bad'}"
                         style="width:{wr}%"></div>
                  </div>
                {/if}
              </div>
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>

  <!-- ── Categorical (right) ──────────────────────────────────────────── -->
  <div class="flex-1 rounded-2xl border border-card-edge bg-card shadow-md shadow-darken overflow-hidden">
    <!-- Tab bar -->
    <div class="flex items-center border-b border-card-edge px-1">
      <span class="font-fancy text-sm tracking-wide text-font-dimer px-3">Category</span>
      <div class="grow"></div>
      {#each tabs as tab}
        <button
          onclick={() => switchTab(tab)}
          class="relative px-4 py-3 font-fancy text-sm tracking-wide transition-colors duration-150 cursor-pointer
                 {active === tab ? 'text-font-clear' : 'text-font-dimest hover:text-font-dim'}">
          {tab}
          {#if active === tab}
            <span class="absolute bottom-0 left-3 right-3 h-[2px] rounded-full bg-gradient-primary"></span>
          {/if}
        </button>
      {/each}
    </div>
    <table class="w-full">
      <thead>
        <tr class="border-b border-card-edge">
          <th class="font-fancy text-[10px] font-semibold tracking-widest uppercase text-font-dimest text-center px-4 py-2.5 w-10">#</th>
          <th class="font-fancy text-[10px] font-semibold tracking-widest uppercase text-font-dimest text-left px-3 py-2.5">Player</th>
          <th class="font-fancy text-[10px] font-semibold tracking-widest uppercase text-font-dimest text-right px-3 py-2.5 w-24">Rating</th>
          <th class="font-fancy text-[10px] font-semibold tracking-widest uppercase text-font-dimest text-center px-4 py-2.5 w-16">
            <Tooltip label="Win rate across rated games" placement="bottom" align="right">
              {#snippet children()}<span class="cursor-help">WR</span>{/snippet}
            </Tooltip>
          </th>
        </tr>
      </thead>
      <tbody>
        {#each data.categorical as player, i}
          {@const wr = player.games > 0 ? Math.round(Number(player.wins) / Number(player.games) * 100) : 0}
          <tr class="not-last:border-b border-card-edge odd:bg-zebra-2 transition-colors duration-100 group">
            <td class="text-center px-4 py-3">
              <span class="text-font-dimer text-sm ml-3">{i + 1}</span>
            </td>
            <td class="px-3 py-3">
              <div class="flex items-center gap-3">
                <Avatar id={player.id} name={player.name} avatar={player.avatar}
                        denounced={denouncedSet.has(player.id)}
                        wrapClass="h-8 w-8 rounded-full bg-card-2 border border-card-edge shrink-0"
                        letterClass="text-primary text-[10px] font-bold font-fancy select-none" />
                <a href="/profile/{player.id}" class="flex items-center gap-2 w-fit">
                  <span class="racking-wide text-font-dim group-hover:text-font-clear transition-colors duration-150">
                    {player.name}
                  </span>
                  {#if Number(player.streak) > 1}
                    <span class="flex items-center gap-0.5 text-xs font-semibold text-font-good bg-font-good/10 border border-font-good/20 px-1.5 py-0.5 rounded-full">
                      <Flame class="h-3 w-3" strokeWidth={2} />{player.streak}
                    </span>
                  {/if}
                </a>
              </div>
            </td>
            <td class="text-right px-3 py-3">
              <span class="text-font-clear font-semibold tabular-nums">{Math.round(Number(player.rating))}</span>
              {#if Number(player.rd) >= 110}
                <span class="text-font-dimest text-xs ml-1">±{Math.round(Number(player.rd))}</span>
              {/if}
            </td>
            <td class="text-right px-4 py-3">
              <div class="flex flex-col items-end gap-1 mr-3">
                <span class="text-font-dim tabular-nums text-sm">{player.games > 0 ? wr + '%' : '—'}</span>
                {#if player.games > 0}
                  <div class="w-12 h-[3px] rounded-full bg-card-edge overflow-hidden">
                    <div class="h-full rounded-full transition-all duration-300
                                {wr >= 60 ? 'bg-font-good' : wr >= 40 ? 'bg-primary' : 'bg-font-bad'}"
                         style="width:{wr}%"></div>
                  </div>
                {/if}
              </div>
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>

  </div> <!-- end tables row -->
</div>
