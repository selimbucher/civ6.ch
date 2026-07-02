<script lang="ts">
    import { enhance } from '$app/forms';
    import bg from '$lib/assets/hero.webp';
    import { LogIn, User, Lock } from '@lucide/svelte';
    let { form } = $props();
    let loading = $state(false);
</script>

<!-- faded hero backdrop tying the login to the landing page -->
<div class="pointer-events-none fixed inset-0 -z-10 overflow-hidden">
    <div
        class="absolute inset-0 opacity-20 blur-[2px]"
        style="background-image: url('{bg}'); background-size: cover; background-position: 60% 25%;"
    ></div>
    <div class="absolute inset-0 bg-gradient-to-b from-background/70 via-background/85 to-background"></div>
    <div
        class="absolute left-1/2 top-1/2 h-[36rem] w-[36rem] -translate-x-1/2 -translate-y-1/2 rounded-full"
        style="background: radial-gradient(circle, hsl(49.88deg 73% 50% / 12%) 0%, transparent 70%)"
    ></div>
</div>

<div class="flex flex-1 items-center justify-center px-4 py-10">
    <div class="w-full max-w-sm">
        <!-- crest -->
        <div class="mb-7 flex flex-col items-center gap-1.5">
            <h1
                class="font-fancy text-4xl font-bold text-primary"
                style="text-shadow: 2px 2px 0px var(--color-primary-shadow);"
            >
                civ6.ch
            </h1>
            <p class="font-fancy text-[0.7rem] uppercase tracking-[0.35em] text-font-dimer">
                Swiss Civilization VI League
            </p>
        </div>

        <div class="overflow-hidden rounded-2xl border border-card-edge bg-card shadow-md shadow-darken">
            <div class="h-[3px] bg-gradient-primary"></div>
            <div class="flex flex-col gap-5 p-7">
                <div class="flex flex-col gap-1 text-center">
                    <h2 class="font-fancy text-xl font-semibold text-font-clear">Welcome back, Leader</h2>
                    <p class="text-xs text-font-dimer">Sign in to defend your rating.</p>
                </div>

                {#if form?.error}
                    <div class="rounded-lg border border-font-bad/30 bg-font-bad/5 px-3 py-2 text-center text-sm text-font-bad">
                        {form.error}
                    </div>
                {/if}

                <form
                    method="POST"
                    use:enhance={() => {
                        loading = true;
                        return async ({ update }) => {
                            await update();
                            loading = false;
                        };
                    }}
                    class="flex flex-col gap-3"
                >
                    <label class="flex flex-col gap-1.5">
                        <span class="text-[0.7rem] uppercase tracking-wider text-font-dimest">Username or Email</span>
                        <div class="relative">
                            <User class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-font-dimest" />
                            <input
                                name="username"
                                type="text"
                                autocomplete="username"
                                placeholder="hannibal_barca"
                                class="w-full rounded-lg border border-card-edge bg-card-2 py-2.5 pl-9 pr-3 text-sm text-font-clear outline-none transition-colors duration-150 placeholder:text-font-dimest focus:border-primary/40"
                            />
                        </div>
                    </label>

                    <label class="flex flex-col gap-1.5">
                        <span class="text-[0.7rem] uppercase tracking-wider text-font-dimest">Password</span>
                        <div class="relative">
                            <Lock class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-font-dimest" />
                            <input
                                name="password"
                                type="password"
                                autocomplete="current-password"
                                placeholder="••••••••"
                                class="w-full rounded-lg border border-card-edge bg-card-2 py-2.5 pl-9 pr-3 text-sm text-font-clear outline-none transition-colors duration-150 placeholder:text-font-dimest focus:border-primary/40"
                            />
                        </div>
                    </label>

                    <button
                        type="submit"
                        disabled={loading}
                        class="mt-1 flex items-center justify-center gap-2 rounded-lg bg-gradient-primary py-2.5 font-fancy text-sm font-bold tracking-wider text-black transition-all duration-150 hover:brightness-125 disabled:opacity-60"
                    >
                        <LogIn class="h-4 w-4" />
                        {loading ? 'Entering…' : 'Enter the Arena'}
                    </button>
                </form>
            </div>
        </div>

        <p class="mt-5 text-center text-xs text-font-dimest">
            Membership is by conquest only — ask a regular for an account.
        </p>
    </div>
</div>
