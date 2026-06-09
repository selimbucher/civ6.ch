<script lang="ts">
  import domination from '$lib/assets/icons/vcondition/domination.png';
  import religious from '$lib/assets/icons/vcondition/religious.png';
  import diplomatic from '$lib/assets/icons/vcondition/diplomatic.png';
  import science from '$lib/assets/icons/vcondition/science.png';
  import culture from '$lib/assets/icons/vcondition/culture.png';
  import score from '$lib/assets/icons/vcondition/score.png';
  import capitulation from '$lib/assets/icons/vcondition/capitulation.png';
  import { Swords, Map, AlertCircle } from '@lucide/svelte';

  let { data } = $props();

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

  const victoryIcons: Record<string, string> = {
    'Domination': domination,
    'Religious': religious,
    'Diplomatic': diplomatic,
    'Science': science,
    'Culture': culture,
    'Score': score,
    'Capitulation': capitulation,
  };

  const victoryColor: Record<string, string> = {
    'Domination':  'var(--color-army-2)',
    'Religious':   'var(--color-faith-2)',
    'Diplomatic':  'var(--color-diplo-2)',
    'Science':     'var(--color-science-2)',
    'Culture':     'var(--color-culture-2)',
    'Score':       'var(--color-score-2)',
    'Capitulation':'var(--color-font-dimer)',
  };

  function getTeams(players: any[]) {
    const teams: Record<number, any[]> = {};
    for (const p of players) {
      if (!teams[p.team]) teams[p.team] = [];
      teams[p.team].push(p);
    }
    return Object.values(teams);
  }

  function formatDelta(pre: number, post: number) {
    const delta = Math.round(post - pre);
    if (delta > 0) return { text: `+${delta}`, cls: 'text-font-good' };
    if (delta < 0) return { text: `${delta}`, cls: 'text-font-bad' };
    return { text: '+0', cls: 'text-font-dim' };
  }

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

  const categoryLabel: Record<string, string> = {
    'ffa': 'FFA',
    'teams': 'Teams',
    '1v1': '1v1',
  };
</script>

<!-- Awaiting Confirmation -->
{#if data.unconfirmed && data.unconfirmed.length > 0}
  <div class="mx-12 mb-5">
    <div class="rounded-2xl border border-primary/20 bg-primary/5 shadow-sm shadow-darken overflow-hidden">
      <div class="flex items-center gap-2 px-5 py-3 border-b border-primary/15">
        <AlertCircle class="h-3.5 w-3.5 text-primary" strokeWidth={2} />
        <span class="font-fancy text-xs font-semibold tracking-widest uppercase text-primary/80">Awaiting Confirmation</span>
      </div>
      <div class="flex flex-wrap gap-3 p-4">
        {#each data.unconfirmed as g}
          <a
            href="/matches/confirm/{g.id}"
            class="group flex items-center gap-3 rounded-xl border border-card-edge bg-card px-4 py-3 shadow-sm shadow-darken hover:border-primary/40 hover:bg-select transition-colors duration-150"
          >
            {#if g.has_map}
              <img src="/files/maps/{g.id}" alt="" class="h-10 w-20 rounded-lg object-cover shrink-0" />
            {:else}
              <div class="h-10 w-20 rounded-lg shrink-0 bg-card-2 border border-card-edge flex items-center justify-center">
                <Map strokeWidth={1} class="h-4 w-4 text-font-dimest" />
              </div>
            {/if}
            <div class="flex flex-col gap-1 min-w-0">
              <div class="flex items-center gap-2 text-xs text-font-dimer">
                {#if g.map}<span class="text-font-dim font-medium">{g.map}</span><span>·</span>{/if}
                <span>{#if g.turns}Turn {g.turns} ·{/if} {formatDate(g.date)}</span>
              </div>
              <div class="flex items-center gap-1.5">
                {#each g.players as p}
                  <div class="flex items-center gap-1 shrink-0">
                    <div class="h-5 w-5 rounded-full bg-card-edge overflow-hidden">
                      {#if leaderPortrait(p.leader)}
                        <img src={leaderPortrait(p.leader)!} alt="" class="h-full w-full object-cover" />
                      {/if}
                    </div>
                    {#if p.pseudo_name}
                      <span class="text-xs text-font-dimer">{p.pseudo_name}</span>
                    {/if}
                  </div>
                {/each}
              </div>
            </div>
            <Swords strokeWidth={1.5} class="h-4 w-4 text-font-dimest shrink-0 ml-1 group-hover:text-primary transition-colors duration-150" />
          </a>
        {/each}
      </div>
    </div>
  </div>
{/if}

<!-- Game list -->
<div class="flex flex-col mx-12 mb-12 bg-card border border-card-edge rounded-2xl shadow-md shadow-darken overflow-hidden">
  {#each data.games as game}
    {@const teams = getTeams(game.players)}
    {@const large = game.players.length >= 5}
    <div class="relative flex w-full p-5 not-last:border-b border-card-edge even:bg-zebra-2 hover:bg-select transition-colors duration-200 ease-out gap-5">
      <a href="/matches/view/{game.id}" class="absolute inset-0 z-0" aria-label="View match"></a>

      <!-- Map thumbnail -->
      {#if game.has_map}
        <img src="/files/maps/{game.id}" alt="" class="h-44 w-80 rounded-xl object-cover shrink-0" />
      {:else}
        <div class="rounded-xl h-44 w-80 shrink-0 bg-card-2 border border-card-edge flex flex-col items-center justify-center gap-1.5 text-font-dimest">
          <Map strokeWidth={1} class="h-6 w-6" />
          <span class="text-xs">Map not available</span>
        </div>
      {/if}

      <!-- Content -->
      <div class="flex flex-col flex-1 min-w-0 gap-3 py-1">

        <!-- Header -->
        <div class="flex items-start justify-between">
          <div class="flex items-center gap-3">
            <img src={victoryIcons[game.victory_type] ?? domination} alt="" class="h-11 w-11 shrink-0" />
            <div class="flex flex-col gap-1">
              <span class="font-fancy font-semibold tracking-wide" style="color: {victoryColor[game.victory_type] ?? 'var(--color-font-clear)'};">
                {game.victory_type} Victory
              </span>
              <div class="flex items-center gap-2 text-xs text-font-dimer">
                <span class="font-fancy text-[10px] font-semibold tracking-wider uppercase px-2 py-0.5 rounded-full bg-card-2 border border-card-edge text-font-dim">
                  {categoryLabel[game.category] ?? game.category}
                </span>
                {#if game.map}
                  <span>·</span><span>{game.map}</span>
                {/if}
                {#if game.turns}
                  <span>·</span><span>{game.turns} turns</span>
                {/if}
              </div>
            </div>
          </div>
          <span class="text-xs text-font-dimest mt-1 shrink-0">{formatDate(game.date)}</span>
        </div>

        <!-- Player matchup -->
        <div class="flex-1 flex items-center">
          {#each teams as team, i}
            {#if i > 0}
              <Swords strokeWidth={1.5} class="shrink-0 text-font-dimest {large ? 'mx-2 h-6 w-6' : 'mx-4 h-8 w-8'}" />
            {/if}

            <div class="flex flex-col gap-1.5 {i === 0 ? 'flex-1 items-end' : i === teams.length - 1 ? 'flex-1 items-start' : 'items-center'}">
              {#each team as player}
                {@const d = formatDelta(player.pre_rating, player.post_rating)}
                {@const displayName = teams.length > 2 ? player.name.split(' ').slice(0, 2).join(' ') : player.name}
                <div class="flex items-center gap-2 {i === 0 ? 'flex-row-reverse' : 'flex-row'}">
                  <div class="shrink-0 h-7 w-7 rounded-full bg-card-edge overflow-hidden ring-1 {player.winner ? 'ring-primary/60' : 'ring-transparent'}">
                    {#if leaderPortrait(player.leader)}
                      <img src={leaderPortrait(player.leader)!} alt="" class="h-full w-full object-cover" />
                    {/if}
                  </div>
                  <div class="flex flex-col {i === 0 ? 'items-end' : 'items-start'}">
                    <span class="{large ? 'text-xs' : 'text-sm'} {player.winner ? 'text-font-clear font-semibold' : 'text-font-dim'}">{displayName}</span>
                    <span class="{large ? 'text-[10px]' : 'text-xs'} text-font-dimer tabular-nums">
                      {Math.round(player.pre_rating)} <span class={d.cls}>{d.text}</span>
                    </span>
                  </div>
                </div>
              {/each}
            </div>
          {/each}
        </div>

      </div>
    </div>
  {/each}
</div>
