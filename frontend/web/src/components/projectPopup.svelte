<script lang="ts">
	/**
	 * Project detail dialog.
	 *
	 * Opened by a row on the overview and projects lists. The 2023 version was the
	 * project card itself: it rendered `window.open` on the access URL, printed
	 * raw ARNs into `h1` headings, used green/red/yellow Tailwind palette classes
	 * that no longer exist in this theme, and its "Terminate(X)" button fired
	 * immediately with no confirmation and no result reported to the user.
	 *
	 * Terminating is destructive, so it is now a two-step confirm with a busy state
	 * and a real error surface. The request itself is unchanged.
	 */
	import { Dialog as DialogPrimitive } from 'bits-ui';
	import X from 'lucide-svelte/icons/x';
	import ExternalLink from 'lucide-svelte/icons/external-link';
	import Server from 'lucide-svelte/icons/server';
	import Button from '$lib/components/ui/Button.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import EmptyState from '$lib/components/ui/EmptyState.svelte';
	import { formatAbsolute, terminateProject, type ProjectRow } from './projectData';

	let {
		project = null,
		open = $bindable(false),
		onchanged
	}: {
		project?: ProjectRow | null;
		open?: boolean;
		onchanged?: () => void | Promise<void>;
	} = $props();

	let confirming = $state(false);
	let terminating = $state(false);
	let error = $state('');

	// A reopened dialog must never still be armed for a destructive action.
	$effect(() => {
		if (!open) {
			confirming = false;
			error = '';
		}
	});

	function openAccess() {
		if (!project?.accessUrl) return;
		window.open(project.accessUrl, '_blank', 'noopener');
	}

	async function terminate() {
		if (!project) return;

		if (!confirming) {
			confirming = true;
			return;
		}

		terminating = true;
		error = '';
		try {
			await terminateProject(project.uuid, project.name);
			await onchanged?.();
			open = false;
		} catch (failure) {
			error =
				failure instanceof Error && failure.message
					? failure.message
					: 'The terminate request failed.';
		} finally {
			terminating = false;
			confirming = false;
		}
	}
</script>

<DialogPrimitive.Root bind:open>
	<DialogPrimitive.Portal>
		<DialogPrimitive.Overlay class="fixed inset-0 z-50 bg-black/65" />
		<DialogPrimitive.Content
			class="fixed top-1/2 left-1/2 z-50 flex max-h-[85vh] w-[min(680px,94vw)] -translate-x-1/2
				-translate-y-1/2 flex-col overflow-hidden rounded-lg border border-border
				bg-dark-surfaceContainer outline-none"
		>
			{#if project}
				<header class="flex items-start justify-between gap-4 border-b border-border px-5 py-4">
					<div class="flex min-w-0 flex-col gap-1.5">
						<DialogPrimitive.Title
							class="truncate text-[14px] leading-tight font-semibold text-dark-onSurface"
						>
							{project.name}
						</DialogPrimitive.Title>
						<DialogPrimitive.Description class="flex items-center gap-2">
							<Badge tone={project.statusTone} dot>{project.status}</Badge>
							<span class="truncate text-[12px] text-dark-onSurfaceVariant">{project.target}</span>
						</DialogPrimitive.Description>
					</div>

					<DialogPrimitive.Close
						class="flex h-7 w-7 shrink-0 items-center justify-center rounded-md text-dark-onSurfaceVariant
							transition-colors hover:bg-dark-surfaceContainerHigh hover:text-dark-onSurface
							focus-visible:outline-none"
					>
						<X size={15} strokeWidth={1.75} aria-hidden="true" />
						<span class="sr-only">Close</span>
					</DialogPrimitive.Close>
				</header>

				<div class="min-h-0 flex-1 overflow-y-auto px-5 py-4">
					<dl class="grid grid-cols-2 gap-x-5 gap-y-3">
						<div class="flex min-w-0 flex-col gap-0.5">
							<dt
								class="text-[11px] font-medium tracking-wide text-dark-onSurfaceVariant uppercase"
							>
								Type
							</dt>
							<dd class="truncate text-[13px] text-dark-onSurface">{project.type}</dd>
						</div>
						<div class="flex min-w-0 flex-col gap-0.5">
							<dt
								class="text-[11px] font-medium tracking-wide text-dark-onSurfaceVariant uppercase"
							>
								Platform
							</dt>
							<dd class="truncate text-[13px] text-dark-onSurface">{project.platform}</dd>
						</div>
						<div class="flex min-w-0 flex-col gap-0.5">
							<dt
								class="text-[11px] font-medium tracking-wide text-dark-onSurfaceVariant uppercase"
							>
								Host
							</dt>
							<dd class="truncate text-[13px] text-dark-onSurface">
								{project.host || 'Not assigned'}
							</dd>
						</div>
						<div class="flex min-w-0 flex-col gap-0.5">
							<dt
								class="text-[11px] font-medium tracking-wide text-dark-onSurfaceVariant uppercase"
							>
								Last deploy
							</dt>
							<dd
								class="truncate text-[13px] text-dark-onSurface"
								title={formatAbsolute(project.lastDeployAt)}
							>
								{project.lastDeployLabel}
							</dd>
						</div>
						<div class="col-span-2 flex min-w-0 flex-col gap-0.5">
							<dt
								class="text-[11px] font-medium tracking-wide text-dark-onSurfaceVariant uppercase"
							>
								Repository
							</dt>
							<dd class="flex min-w-0 items-center gap-2">
								{#if project.repoUrl}
									<a
										href={project.repoUrl}
										target="_blank"
										rel="noreferrer"
										class="truncate text-[13px] text-dark-primary hover:underline"
									>
										{project.repoFullName}
									</a>
								{:else}
									<span class="truncate text-[13px] text-dark-onSurface">{project.target}</span>
								{/if}
								{#if project.repoPrivate}
									<Badge tone="neutral">Private</Badge>
								{/if}
							</dd>
						</div>
					</dl>

					{#if project.description}
						<p class="mt-4 text-[13px] leading-relaxed text-dark-onSurfaceVariant">
							{project.description}
						</p>
					{/if}

					<div class="mt-5 flex items-center justify-between gap-3">
						<h3 class="text-[11px] font-medium tracking-wide text-dark-onSurfaceVariant uppercase">
							Deployments
						</h3>
						<span class="text-[11px] text-dark-onSurfaceVariant">
							{project.total}
							{project.total === 1 ? 'instance' : 'instances'}
						</span>
					</div>

					<div class="mt-2 flex flex-col gap-2">
						{#if project.total === 0}
							<EmptyState
								title="No deployments yet"
								description="This project has not been deployed. Use Deploy project from the projects list to create one."
							>
								{#snippet icon()}
									<Server size={22} strokeWidth={1.5} aria-hidden="true" />
								{/snippet}
							</EmptyState>
						{:else}
							{#each project.instances as instance (instance.uuid || instance.name)}
								<div class="overflow-hidden rounded-md border border-border bg-dark-surfaceContainerLow">
									<div class="flex items-center justify-between gap-3 px-3 py-2">
										<span class="truncate text-[13px] font-medium text-dark-onSurface">
											{instance.name}
										</span>
										<Badge tone={instance.statusTone} dot>{instance.status}</Badge>
									</div>

									{#if instance.resources.length > 0}
										<ul class="divide-y divide-border border-t border-border">
											{#each instance.resources as resource (resource.id + resource.name)}
												<li class="flex items-center gap-3 px-3 py-1.5">
													<span
														class="w-24 shrink-0 truncate text-[11px] tracking-wide text-dark-onSurfaceVariant uppercase"
													>
														{resource.service}
													</span>
													<span class="min-w-0 flex-1 truncate text-[12px] text-dark-onSurface">
														{resource.name}
													</span>
													<Badge tone={resource.statusTone}>{resource.status}</Badge>
												</li>
											{/each}
										</ul>
									{:else}
										<p
											class="border-t border-border px-3 py-2 text-[12px] text-dark-onSurfaceVariant"
										>
											No resources reported.
										</p>
									{/if}
								</div>
							{/each}
						{/if}
					</div>

					{#if error}
						<p class="mt-4 rounded-md border border-dark-error/40 bg-dark-errorContainer/30 px-3 py-2 text-[12px] text-dark-onSurface" role="alert">
							{error}
						</p>
					{/if}
				</div>

				<footer
					class="flex items-center justify-between gap-2 border-t border-border bg-dark-surfaceContainerLow px-5 py-3"
				>
					<Button
						variant="ghost"
						disabled={!project.accessUrl}
						onclick={openAccess}
						title={project.accessUrl || 'No access URL assigned yet'}
					>
						<ExternalLink size={13} strokeWidth={1.75} aria-hidden="true" />
						Open access URL
					</Button>

					<div class="flex items-center gap-2">
						<Button variant="ghost" disabled={terminating} onclick={() => (open = false)}>
							Close
						</Button>
						<Button
							variant={confirming ? 'danger' : 'secondary'}
							loading={terminating}
							disabled={terminating}
							onclick={terminate}
						>
							{confirming ? 'Confirm terminate' : 'Terminate'}
						</Button>
					</div>
				</footer>
			{/if}
		</DialogPrimitive.Content>
	</DialogPrimitive.Portal>
</DialogPrimitive.Root>
