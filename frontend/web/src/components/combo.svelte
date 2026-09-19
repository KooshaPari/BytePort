<script lang="ts">
	/**
	 * Single-select dropdown for list toolbars.
	 *
	 * Hand-rolled rather than built on a popover primitive: it only ever needs
	 * "below the trigger, inside the page", so avoiding a portal and floating
	 * positioning removes any interaction with the focus scope of the dialogs
	 * these toolbars sit beside. Keyboard semantics follow a listbox: Down/Up
	 * move the active row, Enter/Space commit, Escape closes and refocuses the
	 * trigger, Home/End jump.
	 */
	import ChevronDown from 'lucide-svelte/icons/chevron-down';
	import Check from 'lucide-svelte/icons/check';

	export interface ComboOption {
		value: string;
		label: string;
	}

	let {
		options,
		value = $bindable(''),
		placeholder = 'All',
		label = '',
		disabled = false,
		class: klass = ''
	}: {
		options: ComboOption[];
		value?: string;
		placeholder?: string;
		label?: string;
		disabled?: boolean;
		class?: string;
	} = $props();

	const uid = $props.id();
	const listId = `${uid}-list`;

	let open = $state(false);
	let activeIndex = $state(0);
	let root = $state<HTMLDivElement | null>(null);
	let trigger = $state<HTMLButtonElement | null>(null);

	const selected = $derived(options.find((option) => option.value === value) ?? null);
	const activeId = $derived(
		open && options[activeIndex] ? `${uid}-opt-${activeIndex}` : undefined
	);

	function openList() {
		if (disabled) return;
		const current = options.findIndex((option) => option.value === value);
		activeIndex = current >= 0 ? current : 0;
		open = true;
	}

	function closeList(refocus = true) {
		open = false;
		if (refocus) trigger?.focus();
	}

	function commit(index: number) {
		const option = options[index];
		if (!option) return;
		value = option.value;
		closeList();
	}

	function onkeydown(event: KeyboardEvent) {
		if (disabled) return;

		if (event.key === 'ArrowDown' || event.key === 'ArrowUp') {
			event.preventDefault();
			if (!open) {
				openList();
				return;
			}
			const step = event.key === 'ArrowDown' ? 1 : -1;
			activeIndex = Math.min(
				Math.max(activeIndex + step, 0),
				Math.max(options.length - 1, 0)
			);
			return;
		}

		if (event.key === 'Enter' || event.key === ' ') {
			if (!open) {
				event.preventDefault();
				openList();
				return;
			}
			if (event.key === ' ') event.preventDefault();
			commit(activeIndex);
			return;
		}

		if (event.key === 'Escape' && open) {
			event.preventDefault();
			closeList();
			return;
		}

		if (open && event.key === 'Home') {
			event.preventDefault();
			activeIndex = 0;
			return;
		}

		if (open && event.key === 'End') {
			event.preventDefault();
			activeIndex = Math.max(options.length - 1, 0);
		}
	}

	// Close on outside pointer down. Capture phase so a click that lands on
	// another control still dismisses this list first.
	$effect(() => {
		if (!open) return;

		const onPointerDown = (event: MouseEvent) => {
			if (root && !root.contains(event.target as Node)) open = false;
		};

		document.addEventListener('mousedown', onPointerDown, true);
		return () => document.removeEventListener('mousedown', onPointerDown, true);
	});
</script>

<div class="relative {klass}" bind:this={root}>
	{#if label}
		<span
			class="text-dark-onSurfaceVariant mb-1 block text-[11px] font-medium tracking-wide uppercase"
		>
			{label}
		</span>
	{/if}

	<button
		bind:this={trigger}
		type="button"
		{disabled}
		{onkeydown}
		onclick={() => (open ? closeList(false) : openList())}
		role="combobox"
		aria-haspopup="listbox"
		aria-expanded={open}
		aria-controls={listId}
		aria-activedescendant={activeId}
		class="border-border bg-dark-surfaceContainerLowest hover:border-dark-outline hover:bg-dark-surfaceContainerLow focus-visible:border-dark-primary flex h-8 w-full items-center
			justify-between gap-2 rounded-md border
			px-2.5 text-[13px]
			transition-colors focus-visible:outline-none disabled:opacity-50"
	>
		<span class="truncate {selected ? 'text-dark-onSurface' : 'text-dark-onSurfaceVariant'}">
			{selected?.label ?? placeholder}
		</span>
		<ChevronDown
			size={14}
			strokeWidth={1.75}
			class="text-dark-onSurfaceVariant shrink-0"
			aria-hidden="true"
		/>
	</button>

	{#if open}
		<div
			id={listId}
			role="listbox"
			class="border-border bg-dark-surfaceContainerHigh absolute left-0 z-30 mt-1 max-h-64 w-full min-w-[11rem]
				overflow-y-auto rounded-lg border p-1"
		>
			{#each options as option, index (option.value)}
				<button
					id={`${uid}-opt-${index}`}
					type="button"
					role="option"
					aria-selected={option.value === value}
					onmousemove={() => (activeIndex = index)}
					onclick={() => commit(index)}
					class="flex h-8 w-full items-center gap-2 rounded px-2 text-left text-[13px] transition-colors
						{index === activeIndex
						? 'bg-dark-surfaceContainerHighest text-dark-onSurface'
						: 'text-dark-onSurfaceVariant'}"
				>
					<Check
						size={13}
						strokeWidth={2}
						class={option.value === value ? 'text-dark-primary' : 'text-transparent'}
						aria-hidden="true"
					/>
					<span class="truncate">{option.label}</span>
				</button>
			{/each}
		</div>
	{/if}
</div>
