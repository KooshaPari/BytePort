<script lang="ts">
	/**
	 * Step 2 of the new-project wizard: name, description, type, platform.
	 *
	 * The 2023 version was a bare `<form>` of grid-less inputs with placeholder
	 * text instead of labels (`document.querySelector` was then used to call
	 * `reportValidity()`), so nothing was announced to assistive tech and every
	 * field looked identical. Fields now carry real labels, inline errors and an
	 * `aria-invalid` flag the dialog can focus.
	 *
	 * Type and platform are 3-4 fixed choices each, so they are rendered as a
	 * segmented control rather than a dropdown: no popup, one click, and the
	 * options are visible without discovery.
	 */
	import Input from '$lib/components/ui/Input.svelte';

	let {
		name = $bindable(''),
		description = $bindable(''),
		type = $bindable('single-page'),
		platform = $bindable('web'),
		errors = {},
		formId = 'project-details',
		onsubmit
	}: {
		name?: string;
		description?: string;
		type?: string;
		platform?: string;
		errors?: Record<string, string>;
		formId?: string;
		onsubmit?: () => void;
	} = $props();

	const platforms = [
		{ value: 'web', label: 'Web' },
		{ value: 'mobile', label: 'Mobile' },
		{ value: 'desktop', label: 'Desktop' }
	];

	const types = [
		{ value: 'single-page', label: 'Single page' },
		{ value: 'multi-page', label: 'Multi page' },
		{ value: 'plugin', label: 'Plugin' },
		{ value: 'service', label: 'Service' }
	];
	const readValue = (event: Event) => (event.currentTarget as HTMLInputElement).value;
</script>

{#snippet choiceGroup(
	legend: string,
	groupName: string,
	options: { value: string; label: string }[],
	selected: string,
	onselect: (value: string) => void,
	error: string
)}
	<fieldset class="flex flex-col gap-1.5">
		<legend class="text-dark-onSurfaceVariant text-[12px] font-medium tracking-wide">
			{legend}
		</legend>
		<div
			class="bg-dark-surfaceContainerLowest flex w-full flex-wrap gap-1 rounded-md border p-1
				{error ? 'border-dark-error' : 'border-border'}"
			role="radiogroup"
			aria-label={legend}
			aria-invalid={Boolean(error)}
		>
			{#each options as option (option.value)}
				<button
					type="button"
					role="radio"
					name={groupName}
					aria-checked={selected === option.value}
					onclick={() => onselect(option.value)}
					class="h-7 flex-1 rounded px-2.5 text-[12px] font-medium whitespace-nowrap transition-colors
						focus-visible:outline-none
						{selected === option.value
						? 'bg-dark-primaryContainer text-dark-onPrimaryContainer'
						: 'text-dark-onSurfaceVariant hover:bg-dark-surfaceContainerHigh hover:text-dark-onSurface'}"
				>
					{option.label}
				</button>
			{/each}
		</div>
		{#if error}
			<p class="text-dark-error text-[12px]">{error}</p>
		{/if}
	</fieldset>
{/snippet}

<form
	id={formId}
	class="flex flex-col gap-4"
	novalidate
	onsubmit={(event) => {
		event.preventDefault();
		onsubmit?.();
	}}
>
	<Input
		label="Project name"
		name="name"
		placeholder="checkout-service"
		autocomplete="off"
		data-field="name"
		value={name}
		oninput={(event) => (name = readValue(event))}
		error={errors.name ?? ''}
		required
	/>

	<div class="flex flex-col gap-1.5">
		<label
			for="{formId}-description"
			class="text-dark-onSurfaceVariant text-[12px] font-medium tracking-wide"
			>Description</label
		>
		<textarea
			id="{formId}-description"
			name="description"
			rows="3"
			placeholder="What this project deploys"
			bind:value={description}
			aria-invalid={Boolean(errors.description)}
			class="bg-dark-surfaceContainerLowest text-dark-onSurface placeholder:text-dark-onSurfaceVariant/60 focus:border-dark-primary focus:ring-ring/60 w-full resize-none rounded-md
				border px-3 py-2
				text-sm transition-colors focus:ring-2 focus:outline-none
				{errors.description ? 'border-dark-error' : 'border-border'}"></textarea>
		{#if errors.description}
			<p class="text-dark-error text-[12px]">{errors.description}</p>
		{/if}
	</div>

	{@render choiceGroup(
		'Platform',
		'platform',
		platforms,
		platform,
		(next: string) => (platform = next),
		errors.platform ?? ''
	)}

	{@render choiceGroup(
		'Type',
		'type',
		types,
		type,
		(next: string) => (type = next),
		errors.type ?? ''
	)}
</form>
