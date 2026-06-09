<script lang="ts">
  import { goto } from '$app/navigation';
  import { Flame, Medal, Trophy } from '@lucide/svelte';

  const tabs = ['FFA', 'Teams', '1v1'];
  const categoryMap: Record<string, string> = {
    'FFA': 'ffa', 'Teams': 'teams', '1v1': '1v1',
  };

  let { data } = $props();
  let active = $state(tabs.find(t => categoryMap[t] === data.category) ?? 'FFA');

  function switchTab(tab: string) {
    active = tab;
    goto(`?category=${categoryMap[tab]}`, { replaceState: true });
  }

  const rankStyle = (i: number) =>
    i === 0 ? 'text-[#FFD700]' :
    i === 1 ? 'text-[#B0B8C8]' :
    i === 2 ? 'text-[#CD7F32]' :
    'text-font-dimest';
</script>

<div class="px-12 pb-12  gap-4 flex justify-center items-center w-full h-full fixed top-0">

  <!-- ── Overall (always visible, left) ──────────────────────────────── -->
  <div class="flex-1 rounded-2xl border border-card-edge bg-card shadow-md shadow-darken overflow-hidden">
    <table class="w-full">
      <thead>
        <tr class="border-b border-card-edge">
          <th class="font-fancy text-[10px] font-semibold tracking-widest uppercase text-font-dimest text-center px-4 py-2.5 w-10"></th>
          <th class="font-fancy text-[10px] font-semibold tracking-widest uppercase text-font-dimest text-left px-3 py-2.5">Player</th>
          <th class="font-fancy text-[10px] font-semibold tracking-widest uppercase text-font-dimest text-right px-3 py-2.5 w-24">Rating</th>
          <th class="font-fancy text-[10px] font-semibold tracking-widest uppercase text-font-dimest text-center px-4 py-2.5 w-16">WR</th>
        </tr>
      </thead>
      <tbody>
        {#each data.overall as player, i}
          {@const wr = player.games > 0 ? Math.round(Number(player.wins) / Number(player.games) * 100) : 0}
          <tr class="not-last:border-b border-card-edge odd:bg-zebra-2 transition-colors duration-100 group">
            <td class="text-center px-4 py-3">
            {#if i > 0}
              <span class="text-font-dimest text-sm flex justify-center items-center ml-3">{i + 1}</span>
            {/if}
            {#if i == 0}
              <span  class="text-sm flex justify-center items-center text-primary drop-shadow-sm drop-shadow-primary ml-3 font-medium" >1</span>
            {/if}
            </td>
            <td class="px-3 py-3">
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
    <div class="flex items-center gap-1.5 px-5 py-3 border-b border-card-edge pb-3.5">
      <span class="font-fancy tracking-wide">Category</span>
      <div class="grow"></div>
      {#each tabs as tab}
        <button
          onclick={() => switchTab(tab)}
          class="font-fancy text-xs px-4 py-1.5 rounded-full transition-colors duration-150 ease-out cursor-pointer
                 {active === tab
                   ? 'bg-gradient-primary text-black font-semibold'
                   : 'text-font-dim hover:text-font-clear hover:bg-select'}"
        >{tab}</button>
      {/each}
    </div>
    <table class="w-full">
      <thead>
        <tr class="border-b border-card-edge">
          <th class="font-fancy text-[10px] font-semibold tracking-widest uppercase text-font-dimest text-center px-4 py-2.5 w-10">#</th>
          <th class="font-fancy text-[10px] font-semibold tracking-widest uppercase text-font-dimest text-left px-3 py-2.5">Player</th>
          <th class="font-fancy text-[10px] font-semibold tracking-widest uppercase text-font-dimest text-right px-3 py-2.5 w-24">Rating</th>
          <th class="font-fancy text-[10px] font-semibold tracking-widest uppercase text-font-dimest text-center px-4 py-2.5 w-16">WR</th>
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

</div>
