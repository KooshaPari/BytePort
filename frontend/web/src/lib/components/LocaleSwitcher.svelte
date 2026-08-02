<script lang="ts">
	/**
	 * LocaleSwitcher — dropdown for selecting one of the supported locales.
	 * Persists choice to localStorage via the `locale` store.
	 */
	import {
		locale,
		LOCALES,
		LOCALE_LABELS,
		LOCALE_FLAGS,
		setLocale,
		type Locale
	} from '$lib/i18n';

	let open = $state(false);

	function pick(l: Locale) {
		setLocale(l);
		open = false;
	}

	function handleKey(e: KeyboardEvent) {
		if (e.key === 'Escape') open = false;
	}
</script>

<svelte:window onkeydown={handleKey} />

<div class="relative">
	<button
		type="button"
		class="btn btn-ghost btn-sm gap-1"
		aria-haspopup="listbox"
		aria-expanded={open}
		aria-label="Select language"
		onclick={() => (open = !open)}
	>
		<span aria-hidden="true">{LOCALE_FLAGS[$locale]}</span>
		<span class="hidden sm:inline">{LOCALE_LABELS[$locale]}</span>
		<svg class="h-3 w-3 opacity-60" viewBox="0 0 12 12" fill="none" aria-hidden="true">
			<path
				d="M3 5l3 3 3-3"
				stroke="currentColor"
				stroke-width="1.5"
				stroke-linecap="round"
				stroke-linejoin="round"
			/>
		</svg>
	</button>

	{#if open}
		<ul
			role="listbox"
			aria-label="Languages"
			class="absolute right-0 mt-2 w-44 rounded-md border border-border bg-base-100 shadow-lg py-1 z-50"
		>
			{#each LOCALES as l (l)}
				<li>
					<button
						type="button"
						role="option"
						aria-selected={l === $locale}
						class="flex w-full items-center gap-2 px-3 py-2 text-left text-sm hover:bg-base-200 {l ===
						$locale
							? 'font-semibold'
							: ''}"
						onclick={() => pick(l)}
					>
						<span aria-hidden="true">{LOCALE_FLAGS[l]}</span>
						<span>{LOCALE_LABELS[l]}</span>
						{#if l === $locale}
							<svg
								class="ml-auto h-4 w-4 text-primary"
								viewBox="0 0 16 16"
								fill="none"
								aria-hidden="true"
							>
								<path
									d="M3 8l3 3 7-7"
									stroke="currentColor"
									stroke-width="2"
									stroke-linecap="round"
									stroke-linejoin="round"
								/>
							</svg>
						{/if}
					</button>
				</li>
			{/each}
		</ul>
	{/if}
</div>
