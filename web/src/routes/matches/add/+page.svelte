<script lang="ts">
    import { CloudUpload, Microscope } from '@lucide/svelte';

    let dragging = $state(false);
    let file = $state<File | null>(null);
    let inputEl = $state<HTMLInputElement>();

    function onDrop(e: DragEvent) {
        e.preventDefault();
        dragging = false;
        const dropped = e.dataTransfer?.files[0];
        if (dropped?.name?.toLowerCase().endsWith('.civ6save')) {
            file = dropped;
            if (form) form.error = undefined;
            if (inputEl) {
                const dt = new DataTransfer();
                dt.items.add(dropped);
                inputEl.files = dt.files;
            }
        }
    }

    function onInput(e: Event) {
        const selected = (e.currentTarget as HTMLInputElement).files?.[0];
        if (selected) {
            file = selected;
            if (form) form.error = undefined;
        }
    }

    let { form }: { form: any } = $props();
</script>

<!-- Mobile: fills space below header/nav -->
<div class="flex flex-col flex-1 p-3 gap-2 md:hidden">
    <form method="POST" enctype="multipart/form-data" class="flex flex-col flex-1">
        <div
            role="button"
            tabindex="0"
            class="relative flex-1 flex flex-col items-center justify-center gap-3 rounded-2xl border-2 border-dashed transition-colors duration-[250ms] cursor-pointer
                   {dragging ? 'border-primary bg-primary/10' : 'border-card-edge bg-card/50'}
                   {file ? 'border-font-good' : 'hover:bg-zebra-2'}"
            ondragover={(e) => { e.preventDefault(); dragging = true; }}
            ondragleave={() => dragging = false}
            ondrop={onDrop}
            onclick={() => inputEl?.click()}
            onkeydown={(e) => e.key === 'Enter' && inputEl?.click()}
        >
            <CloudUpload strokeWidth={1.5} class="h-20 w-20 {file ? 'text-font-good' : 'text-font-dimer'} transition-colors duration-[250ms]" />
            {#if file}
                <span class="text-font-clear text-xl font-semibold">{file.name}</span>
                <span class="text-font-dimer">{(file.size / 1024 / 1024).toFixed(1)} MB — tap to change</span>
            {:else}
                <span class="text-font-dim text-xl text-center px-6">Tap to select your <span class="font-semibold">.Civ6Save</span> file</span>
            {/if}
            <button
                type="submit"
                disabled={!file}
                onclick={(e) => e.stopPropagation()}
                class="absolute top-3/4 -translate-y-1/2 left-1/2 -translate-x-1/2
                       pl-10 flex items-center rounded-full border px-4 py-2 font-semibold transition-colors duration-300 whitespace-nowrap
                       {file ? 'border-transparent bg-font-clear text-background cursor-pointer hover:text-transparent' : 'opacity-0 select-none pointer-events-none'}"
            >
                <Microscope class="h-5 w-5 absolute left-4 {file ? 'text-background magic-button' : ''}" /> Analyse
            </button>
        </div>
        <input bind:this={inputEl} type="file" name="save" accept=".Civ6Save" class="hidden" onchange={onInput} />
    </form>
    {#if form?.error}
        <span class="text-font-bad text-lg text-center">{form.error}</span>
    {/if}
</div>

<!-- Desktop -->
<div class="mx-auto hidden w-full max-w-3xl flex-1 flex-col gap-5 px-6 py-14 md:flex">
    <div class="flex flex-col gap-1.5 text-center">
        <h1 class="font-fancy text-3xl font-bold text-font-clear">Upload a Game</h1>
        <p class="text-sm text-font-dimer">
            Drop in a <span class="text-font-dim">.Civ6Save</span> and we'll parse the scoreboard, update ratings and hand out
            the achievements you've earned (or disgraced yourself with).
        </p>
    </div>
    <form method="POST" enctype="multipart/form-data" class="flex min-h-[22rem] flex-col flex-1">
        <div
            role="button"
            tabindex="0"
            class="relative flex-1 flex flex-col items-center justify-center gap-3 rounded-2xl border-2 border-dashed transition-colors duration-[250ms] cursor-pointer shadow-md shadow-darken
                   {dragging ? 'border-primary bg-primary/10' : 'border-card-edge bg-card'}
                   {file ? 'border-font-good' : 'hover:bg-card-2'}"
            ondragover={(e) => { e.preventDefault(); dragging = true; }}
            ondragleave={() => dragging = false}
            ondrop={onDrop}
            onclick={() => inputEl?.click()}
            onkeydown={(e) => e.key === 'Enter' && inputEl?.click()}
        >
            <CloudUpload strokeWidth={1.5} class="h-20 w-20 {file ? 'text-font-good' : 'text-font-dimer'} transition-colors duration-[250ms]" />
            {#if file}
                <span class="text-font-clear text-xl font-semibold">{file.name}</span>
                <span class="text-font-dimer">{(file.size / 1024 / 1024).toFixed(1)} MB — click to change</span>
            {:else}
                <span class="text-font-dim text-xl">Drop your <span class="font-semibold">.Civ6Save</span> file here</span>
                <span class="text-font-dimer">or click to browse</span>
            {/if}
            <button
                type="submit"
                disabled={!file}
                onclick={(e) => e.stopPropagation()}
                class="absolute top-3/4 -translate-y-1/2 left-1/2 -translate-x-1/2
                       pl-10 flex items-center rounded-full border px-4 py-2 font-semibold transition-colors duration-300 whitespace-nowrap
                       {file ? 'z-2 border-transparent bg-font-clear text-background cursor-pointer hover:text-transparent' : 'opacity-0 select-none pointer-events-none'}"
            >
                <Microscope class="h-5 w-5 absolute left-4 {file ? 'text-background magic-button' : ''}" /> Analyse
            </button>
        </div>
        <input bind:this={inputEl} type="file" name="save" accept=".Civ6Save" class="hidden" onchange={onInput} />
    </form>
    {#if form?.error}
        <span class="text-center text-lg text-font-bad">{form.error}</span>
    {/if}
    <p class="text-center text-xs text-font-dimest">
        Saves live in <span class="text-font-dimer">Documents\My Games\Sid Meier's Civilization VI\Saves\</span> ·
        we only read the scoreboard, never your strategy.
    </p>
</div>
