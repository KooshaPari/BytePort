<script lang="ts">
	/**
	 * New-project wizard.
	 *
	 * The 2023 version drove three stages off a bare `let stage = 1` counter and a
	 * chain of `switch (stage)` cases that reassigned `stage` before the case body
	 * ran, so "case 3" could execute on the way from 2 to 4 and a failed deploy
	 * still advanced to "Deployment Complete!". It read form values through
	 * `document.querySelector('#form-stage-2')`, so anything not inside that exact
	 * form silently vanished, and it never reported a failed `/deploy`.
	 *
	 * This version keeps the same endpoint and payload shape with an explicit step
	 * union, validation before advancing, a real error step, and focus moved to the
	 * step heading on every transition.
	 */
	import { Dialog as DialogPrimitive } from 'bits-ui';
	import { tick } from 'svelte';
	import X from 'lucide-svelte/icons/x';
	import CircleAlert from 'lucide-svelte/icons/circle-alert';
	import CircleCheck from 'lucide-svelte/icons/circle-check';
	import Button from '$lib/components/ui/Button.svelte';
	import GitSearch from './gitSearch.svelte';
	import ProjectForm from './projectForm.svelte';
	import ReviewCard from './reviewCard.svelte';
	import { deployProject } from './projectData';
	import type { Repository } from '$lib/git';

	let {
		open = $bindable(false),
		oncreated
	}: {
		open?: boolean;
		oncreated?: () => void | Promise<void>;
	} = $props();

	type Step = 'repository' | 'details' | 'review' | 'deploying' | 'done' | 'error';

	const STEPS: { id: Step; label: string; subtitle: string }[] = [
		{ id: 'repository', label: 'Repository', subtitle: 'Choose the repository to deploy.' },
		{ id: 'details', label: 'Details', subtitle: 'Name the project and describe it.' },
		{ id: 'review', label: 'Review', subtitle: 'Confirm what will be deployed.' }
	];

	let step = $state<Step>('repository');
	let repository = $state<Repository | null>(null);
	let name = $state('');
	let description = $state('');
	let type = $state('single-page');
	let platform = $state('web');
	let errors = $state<Record<string, string>>({});
	let deployError = $state('');
	let deployedStatus = $state('');
	let stepBody = $state<HTMLElement | null>(null);

	const stepIndex = $derived(STEPS.findIndex((entry) => entry.id === step));
	const subtitle = $derived(
		STEPS.find((entry) => entry.id === step)?.subtitle ?? 'Deploying your project.'
	);
	const busy = $derived(step === 'deploying');
	const title = $derived(
		step === 'done'
			? 'Deployment started'
			: step === 'error'
				? 'Deployment failed'
				: busy
					? 'Deploying project'
					: 'New project'
	);

	function reset() {
		step = 'repository';
		repository = null;
		name = '';
		description = '';
		type = 'single-page';
		platform = 'web';
		errors = {};
		deployError = '';
		deployedStatus = '';
	}

	// A closed wizard always reopens at step 1 rather than wherever it was left.
	$effect(() => {
		if (!open) reset();
	});

	async function focusStep() {
		await tick();
		// Land on the first thing that needs attention: a field that failed
		// validation, otherwise the first field of the step. Steps without fields
		// (review, deploy, result) keep focus on the button the user just pressed,
		// which is what a keyboard user expects.
		stepBody
			?.querySelector<HTMLElement>('[aria-invalid="true"], input:not([disabled]), textarea')
			?.focus();
	}

	function validateDetails(): boolean {
		const next: Record<string, string> = {};
		const trimmed = name.trim();

		if (!trimmed) next.name = 'A project name is required.';
		else if (trimmed.length > 64) next.name = 'Keep the name to 64 characters or fewer.';
		if (description.length > 280) next.description = 'Keep the description to 280 characters.';

		errors = next;

		if (Object.keys(next).length > 0) {
			void focusStep();
			return false;
		}
		return true;
	}

	async function goTo(next: Step) {
		step = next;
		await focusStep();
	}

	async function selectRepository(repo: Repository) {
		repository = repo;
		// Default the name to the repository so the details step is usually one
		// click, but leave the user free to change it.
		if (!name.trim()) name = repo.name;
		await goTo('details');
	}

	async function submitDetails() {
		if (!validateDetails()) return;
		await goTo('review');
	}

	async function deploy() {
		step = 'deploying';
		deployError = '';
		await focusStep();

		try {
			const result = await deployProject({
				name: name.trim(),
				description: description.trim(),
				type,
				platform,
				repository
			});
			deployedStatus = result.status || 'created';
			await oncreated?.();
			await goTo('done');
		} catch (error) {
			deployError =
				error instanceof Error && error.message
					? error.message
					: 'The deploy request could not be completed.';
			await goTo('error');
		}
	}
</script>

<DialogPrimitive.Root bind:open>
	<DialogPrimitive.Portal>
		<DialogPrimitive.Overlay class="fixed inset-0 z-50 bg-black/65" />
		<DialogPrimitive.Content
			class="border-border bg-dark-surfaceContainer fixed top-1/2 left-1/2 z-50 flex max-h-[85vh]
				w-[min(560px,92vw)] -translate-x-1/2 -translate-y-1/2 flex-col overflow-hidden rounded-lg
				border outline-none"
			onInteractOutside={(event) => {
				if (busy) event.preventDefault();
			}}
		>
			<header class="border-border flex items-start justify-between gap-4 border-b px-5 py-4">
				<div class="flex min-w-0 flex-col gap-1" aria-live="polite">
					<DialogPrimitive.Title
						class="text-dark-onSurface text-[14px] leading-tight font-semibold"
					>
						{title}
					</DialogPrimitive.Title>
					<DialogPrimitive.Description
						class="text-dark-onSurfaceVariant text-[12px] leading-snug"
					>
						{subtitle}
					</DialogPrimitive.Description>
				</div>

				<DialogPrimitive.Close
					class="text-dark-onSurfaceVariant hover:bg-dark-surfaceContainerHigh hover:text-dark-onSurface flex h-7 w-7 shrink-0 items-center
						justify-center rounded-md transition-colors
						focus-visible:outline-none disabled:opacity-40"
					disabled={busy}
				>
					<X size={15} strokeWidth={1.75} aria-hidden="true" />
					<span class="sr-only">Close</span>
				</DialogPrimitive.Close>
			</header>

			{#if stepIndex >= 0}
				<div class="border-border flex items-center gap-3 border-b px-5 py-3">
					<span
						class="text-dark-onSurfaceVariant text-[11px] font-medium tracking-wide uppercase"
					>
						Step {stepIndex + 1} of {STEPS.length}
					</span>
					<div class="flex flex-1 items-center gap-1.5" aria-hidden="true">
						{#each STEPS as entry, index (entry.id)}
							<span
								class="h-0.5 flex-1 rounded-full {index <= stepIndex
									? 'bg-dark-primary'
									: 'bg-dark-surfaceContainerHighest'}"
							></span>
						{/each}
					</div>
				</div>
			{/if}

			<div class="min-h-0 flex-1 overflow-y-auto px-5 py-4" bind:this={stepBody}>
				{#if step === 'repository'}
					<GitSearch selected={repository} onselect={selectRepository} />
				{:else if step === 'details'}
					<ProjectForm
						bind:name
						bind:description
						bind:type
						bind:platform
						{errors}
						formId="project-details"
						onsubmit={submitDetails}
					/>
				{:else if step === 'review'}
					<ReviewCard {name} {description} {type} {platform} {repository} />
				{:else if step === 'deploying'}
					<div
						class="flex flex-col items-center justify-center gap-3 py-12"
						role="status"
					>
						<span
							class="border-dark-primary h-5 w-5 animate-spin rounded-full border-2 border-t-transparent"
							aria-hidden="true"
						></span>
						<p class="text-dark-onSurface text-[13px]">Sending the deploy request</p>
						<p class="text-dark-onSurfaceVariant text-[12px]">
							The backend registers the project, then asks NanoVMS to start the
							sandbox.
						</p>
					</div>
				{:else if step === 'done'}
					<div
						class="flex flex-col items-center justify-center gap-3 py-12"
						role="status"
					>
						<CircleCheck
							size={28}
							strokeWidth={1.5}
							class="text-dark-primary"
							aria-hidden="true"
						/>
						<p class="text-dark-onSurface text-[13px] font-medium">
							{name} is deploying
						</p>
						<p class="text-dark-onSurfaceVariant text-[12px]">
							Sandbox status: {deployedStatus}. It will appear on the projects list
							shortly.
						</p>
					</div>
				{:else}
					<div class="flex flex-col items-center justify-center gap-3 py-12" role="alert">
						<CircleAlert
							size={28}
							strokeWidth={1.5}
							class="text-dark-error"
							aria-hidden="true"
						/>
						<p class="text-dark-onSurface text-[13px] font-medium">
							The project was not deployed
						</p>
						<p
							class="text-dark-onSurfaceVariant max-w-sm text-center text-[12px] break-words"
						>
							{deployError}
						</p>
					</div>
				{/if}
			</div>

			<footer
				class="border-border bg-dark-surfaceContainerLow flex items-center justify-end gap-2 border-t px-5 py-3"
			>
				{#if step === 'repository'}
					<Button variant="ghost" onclick={() => (open = false)}>Cancel</Button>
					<Button
						variant="primary"
						disabled={!repository}
						onclick={() => goTo('details')}
					>
						Continue
					</Button>
				{:else if step === 'details'}
					<Button variant="ghost" onclick={() => goTo('repository')}>Back</Button>
					<Button variant="primary" onclick={submitDetails}>Review</Button>
				{:else if step === 'review'}
					<Button variant="ghost" disabled={busy} onclick={() => goTo('details')}
						>Back</Button
					>
					<Button variant="primary" onclick={deploy}>Deploy project</Button>
				{:else if step === 'deploying'}
					<Button variant="secondary" disabled>Deploying</Button>
				{:else if step === 'done'}
					<Button variant="secondary" onclick={() => (open = false)}>Close</Button>
				{:else}
					<Button variant="ghost" onclick={() => (open = false)}>Close</Button>
					<Button variant="primary" onclick={() => goTo('review')}>Try again</Button>
				{/if}
			</footer>
		</DialogPrimitive.Content>
	</DialogPrimitive.Portal>
</DialogPrimitive.Root>
