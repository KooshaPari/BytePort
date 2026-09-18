<script lang="ts">
	/**
	 * Text input with an optional label and inline error.
	 *
	 * The old views styled inputs ad hoc and relied on `border-input`, which was
	 * resolving to invalid CSS, so fields had no visible boundary at all.
	 */
	import type { HTMLInputAttributes } from 'svelte/elements';

	let {
		label = '',
		error = '',
		hint = '',
		class: klass = '',
		...rest
	}: {
		label?: string;
		error?: string;
		hint?: string;
		class?: string;
	} & HTMLInputAttributes = $props();

	const id = `field-${Math.random().toString(36).slice(2, 9)}`;
</script>

<div class="flex flex-col gap-1.5 {klass}">
	{#if label}
		<label for={id} class="text-[12px] font-medium tracking-wide text-dark-onSurfaceVariant">
			{label}
		</label>
	{/if}
	<input
		{id}
		class="h-9 w-full rounded-md border bg-dark-surfaceContainerLowest px-3 text-sm
			text-dark-onSurface placeholder:text-dark-onSurfaceVariant/60
			transition-colors focus:outline-none focus:ring-2 focus:ring-ring/60
			{error ? 'border-dark-error' : 'border-border focus:border-dark-primary'}"
		aria-invalid={Boolean(error)}
		{...rest}
	/>
	{#if error}
		<p class="text-[12px] text-dark-error">{error}</p>
	{:else if hint}
		<p class="text-[12px] text-dark-onSurfaceVariant">{hint}</p>
	{/if}
</div>
