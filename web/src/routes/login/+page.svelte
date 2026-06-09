<script lang="ts">
    import { enhance } from '$app/forms';
    let { form } = $props();
</script>

<div class="flex justify-center items-center w-full h-full fixed top-0">
    <div class="flex flex-col gap-2 w-100">
        <span class="text-2xl text-center">Member Login</span>

        {#if form?.error}
            <span class="text-font-bad text-sm">{form.error}</span>
        {/if}

        <form method="POST" use:enhance class="flex flex-col gap-1">
            <input
                name="username"
                type="text"
                placeholder="Username or Email"
                class="rounded-lg border-none bg-card p-3 placeholder:text-font-dimer text-base outline-none"
            />
            <input
                name="password"
                type="password"
                placeholder="Password"
                class="rounded-lg border-none focus:border-none bg-card p-3 placeholder:text-font-dimer text-base outline-none"
            />
            <button
                type="submit"
                class="mt-1 border rounded-lg px-4 py-2 font-semibold hover:bg-primary hover:border-transparent hover:text-black transition-colors duration-250 ease-out"
            >
                Login
            </button>
        </form>
    </div>
</div>