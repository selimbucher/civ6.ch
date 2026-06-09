<script lang="ts">
    import { CloudUpload, FileCheckCorner, Microscope, Search } from '@lucide/svelte';

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
            // Keep native form submission in sync with drag&drop selection.
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
    let { form } = $props();
</script>

<div class="flex justify-center items-center w-full h-full fixed top-0 flex-col gap-2">
    <form
        method="POST"
        enctype="multipart/form-data"
        class="flex flex-col w-fit bg-card p-6 rounded-3xl shadow-md shadow-darken"
    >

        <!-- Drop zone -->
        <button
            type="button"
            aria-label="Drop zone"
            class="z-1 p-24 px-60 pb-35 flex flex-col items-center justify-center gap-3 rounded-2xl border-2 border-dashed transition-colors duration-100 ease
                   {dragging ? 'border-primary bg-primary/10' : 'border-card-edge bg-card/50'}
                   {file ? 'border-font-good' : 'hover:bg-zebra-2'}
                   p-12 cursor-pointer transition-colors duration-150"
            ondragover={(e) => { e.preventDefault(); dragging = true; }}
            ondragleave={() => dragging = false}
            ondrop={onDrop}
            onclick={() => inputEl?.click()}
        >

            <CloudUpload
                strokeWidth={1.5}
                class="h-20 w-20 {file ? 'text-font-good' : 'text-font-dimer'} transition-colors duration-150"
            />
            
            {#if file}
                <span class="text-font-clear text-xl font-semibold">{file.name}</span>
                <span class="text-font-dimer text">{(file.size / 1024 / 1204).toFixed(1)} MB — click to change</span>
            {:else}
                <span class="text-font-dim text-xl">Drop your <span class="font-semibold">.Civ6Save</span> file here</span>
                <span class="text-font-dimer text">or click to browse</span>
            {/if}
            
        </button>

        <input
            bind:this={inputEl}
            type="file"
            name="save"
            accept=".Civ6Save"
            class="hidden"
            onchange={onInput}
        />

        <button
            type="submit"
            disabled={!file}
            class="absolute left-[50%] top-[61%] pl-11 flex justify-center items-center rounded-full tracking-wide text border px-4 py-2 font-semibold transition-colors duration-300 w-fit m-auto
                   {file
                     ? 'z-2 border-transparent bg-font-clear text-background transition-colors cursor-pointer hover:text-transparent'
                     : 'opacity-0 select-none'}"
            style="translate: -50%;"
        >
            <Microscope class="h-5 w-5 inline-block absolute left-4 {file ? 'text-background magic-button' : ''}"/> Analyse
        </button>
        
    </form>
    {#if form?.result}
        <pre>{JSON.stringify(form.result, null, 2)}</pre>
    {/if}
    {#if form?.error}
        <span class="text-font-bad text-lg">{form.error}</span>
    {/if}
</div>