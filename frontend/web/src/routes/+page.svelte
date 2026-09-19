<script lang="ts">
	/**
	 * Entry route.
	 *
	 * The Tauri window loads `index.html`, which resolves here, so this is the
	 * first thing the user sees. It previously rendered a second copy of the
	 * sidebar (duplicating the app shell) and a product image pointing at
	 * `/src/assets/img/byte.png`, a path that is never published in the bundle;
	 * the real files are `Byte.png` / `BytePort.png` under `src/assets/img/`, and
	 * `src/` is not served, so the brand slot showed a broken image.
	 *
	 * It now does one job: initialise the session, then hand off to `/home`,
	 * which is wrapped by the real shell. The brief branded state below is what
	 * shows during that hand-off.
	 */
	import { onMount, onDestroy } from 'svelte';
	import { goto } from '$app/navigation';
	import { user, initializeUser } from '../stores/user';

	let unsubscribe: (() => void) | undefined;

	onMount(() => {
		unsubscribe = user.subscribe((value) => {
			// Wait for initialisation to settle before deciding.
			if (value.status === 'pending') return;

			if (value.status === 'authenticated') {
				goto('/home');
			} else {
				goto('/login');
			}
		});

		// Never throws: the base URL is resolved defensively in $lib/api.
		initializeUser();
	});

	onDestroy(() => unsubscribe?.());
</script>

<main class="bg-dark-background flex h-screen w-screen items-center justify-center">
	<div class="flex flex-col items-center gap-3" role="status" aria-live="polite">
		<span
			class="bg-dark-primaryContainer text-dark-primary flex h-9 w-9 items-center justify-center rounded-lg"
			aria-hidden="true"
		>
			<svg
				viewBox="0 0 16 16"
				class="h-4 w-4"
				fill="none"
				stroke="currentColor"
				stroke-width="1.25"
				stroke-linecap="round"
				stroke-linejoin="round"
			>
				<path d="M8 1.9 13.7 5v6L8 14.1 2.3 11V5z" />
				<path d="M2.3 5 8 8.1 13.7 5" />
				<path d="M8 8.1v6" />
			</svg>
		</span>
		<p class="text-dark-onSurfaceVariant text-[13px] font-medium">Starting BytePort</p>
		<span class="sr-only">Loading</span>
	</div>
</main>
