<script lang="ts">
	/**
	 * Projects list.
	 *
	 * Replaces a grid of 192x256px cards that repeated a broken logo image, showed
	 * no status, no target and no actions, and imported the route layout as a
	 * component (`import Project from '../+layout.svelte'`) to use as a *type*.
	 *
	 * This is now a dense table with the columns that actually get scanned: name,
	 * state, target, last deploy. Row actions live behind a menu so the row itself
	 * stays a single click target for opening the project.
	 *
	 * Endpoints are unchanged (`GET /projects`, `POST /deploy`, `POST /terminate`).
	 */
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { DropdownMenu as Menu } from 'bits-ui';
	import EllipsisVertical from 'lucide-svelte/icons/ellipsis-vertical';
	import ExternalLink from 'lucide-svelte/icons/external-link';
	import RefreshCw from 'lucide-svelte/icons/refresh-cw';
	import Plus from 'lucide-svelte/icons/plus';
	import Search from 'lucide-svelte/icons/search';
	import Box from 'lucide-svelte/icons/box';
	import CircleAlert from 'lucide-svelte/icons/circle-alert';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import EmptyState from '$lib/components/ui/EmptyState.svelte';
	import Combo from '../../../components/combo.svelte';
	import AddProjectDialog from '../../../components/addProjectDialog.svelte';
	import ProjectPopup from '../../../components/projectPopup.svelte';
	import {
		fetchProjects,
		formatAbsolute,
		watchAuth,
		type ProjectRow
	} from '../../../components/projectData';

	const ICON_BUTTON =
		'flex h-7 w-7 items-center justify-center rounded-md text-dark-onSurfaceVariant ' +
		'transition-colors hover:bg-dark-surfaceContainerHigh hover:text-dark-onSurface ' +
		'focus-visible:outline-none';

	const MENU_ITEM =
		'flex h-8 cursor-pointer items-center gap-2 rounded px-2 text-[13px] text-dark-onSurface ' +
		'outline-none select-none data-[highlighted]:bg-dark-surfaceContainerHighest ' +
		'data-[disabled]:pointer-events-none data-[disabled]:opacity-40';

	const STATUS_FILTERS = [
		{ value: 'all', label: 'All states' },
		{ value: 'running', label: 'Running' },
		{ value: 'deploying', label: 'Deploying' },
		{ value: 'failed', label: 'Failed' },
		{ value: 'stopped', label: 'Stopped' },
		{ value: 'not-deployed', label: 'Not deployed' }
	];

	let projects = $state<ProjectRow[]>([]);
	let phase = $state<'loading' | 'ready' | 'error'>('loading');
	let errorMessage = $state('');
	let query = $state('');
	let statusFilter = $state('all');
	let composerOpen = $state(false);
	let detailOpen = $state(false);
	let selected = $state<ProjectRow | null>(null);

	async function load() {
		if (phase !== 'ready') phase = 'loading';
		errorMessage = '';
		try {
			projects = await fetchProjects();
			phase = 'ready';
		} catch (error) {
			phase = 'error';
			errorMessage =
				error instanceof Error && error.message
					? error.message
					: 'The backend did not return a projects list.';
		}
	}

	onMount(() => watchAuth(goto, () => void load()));

	const filtered = $derived.by(() => {
		const needle = query.trim().toLowerCase();
		return projects.filter((project) => {
			if (statusFilter !== 'all' && project.statusKey !== statusFilter) return false;
			if (!needle) return true;
			return [project.name, project.target, project.type, project.platform]
				.join(' ')
				.toLowerCase()
				.includes(needle);
		});
	});

	function readValue(event: Event) {
		query = (event.currentTarget as HTMLInputElement).value;
	}

	function openDetail(project: ProjectRow) {
		selected = project;
		detailOpen = true;
	}

	function openAccess(project: ProjectRow) {
		if (!project.accessUrl) return;
		window.open(project.accessUrl, '_blank', 'noopener');
	}
</script>

<div class="flex flex-col gap-4">
	<div class="flex flex-wrap items-end gap-2">
		<div class="w-64">
			<Input
				label="Search"
				placeholder="Name, repository or type"
				autocomplete="off"
				value={query}
				oninput={readValue}
			/>
		</div>

		<Combo
			label="State"
			options={STATUS_FILTERS}
			bind:value={statusFilter}
			class="w-40"
			placeholder="All states"
		/>

		<div class="ms-auto flex items-center gap-2 pb-0.5">
			{#if phase === 'ready'}
				<span class="text-dark-onSurfaceVariant text-[11px]">
					{filtered.length} of {projects.length}
				</span>
			{/if}
			<Button variant="secondary" size="sm" disabled={phase === 'loading'} onclick={load}>
				<RefreshCw size={13} strokeWidth={1.75} aria-hidden="true" />
				Refresh
			</Button>
			<Button variant="primary" size="sm" onclick={() => (composerOpen = true)}>
				<Plus size={13} strokeWidth={2} aria-hidden="true" />
				New project
			</Button>
		</div>
	</div>

	{#if phase === 'error'}
		<div
			class="border-border bg-dark-surfaceContainer flex items-start gap-3 rounded-lg border p-4"
			role="alert"
		>
			<CircleAlert
				size={16}
				strokeWidth={1.75}
				class="text-dark-error mt-0.5 shrink-0"
				aria-hidden="true"
			/>
			<div class="flex min-w-0 flex-1 flex-col gap-1">
				<p class="text-dark-onSurface text-[13px] font-medium">
					Projects could not be loaded
				</p>
				<p class="text-dark-onSurfaceVariant text-[12px] break-words">{errorMessage}</p>
			</div>
			<Button variant="secondary" size="sm" onclick={load}>Retry</Button>
		</div>
	{/if}

	<div class="border-border bg-dark-surfaceContainer overflow-hidden rounded-lg border">
		{#if phase === 'loading'}
			<div aria-busy="true" aria-label="Loading projects">
				{#each Array(5) as _, index (index)}
					<div
						class="border-border flex h-11 items-center gap-3 border-b px-4 last:border-b-0"
					>
						<span class="bg-dark-surfaceContainerHigh h-3.5 w-48 animate-pulse rounded"
						></span>
						<span class="bg-dark-surfaceContainerHigh h-3.5 w-20 animate-pulse rounded"
						></span>
						<span
							class="bg-dark-surfaceContainerHigh ms-auto h-3.5 w-28 animate-pulse rounded"
						></span>
					</div>
				{/each}
			</div>
		{:else if projects.length === 0}
			<EmptyState
				title="No projects yet"
				description="A project links a repository to a deployment. Create one to get started."
			>
				{#snippet icon()}
					<Box size={22} strokeWidth={1.5} aria-hidden="true" />
				{/snippet}
				{#snippet actions()}
					<Button variant="primary" size="sm" onclick={() => (composerOpen = true)}>
						<Plus size={13} strokeWidth={2} aria-hidden="true" />
						New project
					</Button>
				{/snippet}
			</EmptyState>
		{:else if filtered.length === 0}
			<EmptyState
				title="No projects match"
				description="Clear the search or pick a different state to see more projects."
			>
				{#snippet icon()}
					<Search size={22} strokeWidth={1.5} aria-hidden="true" />
				{/snippet}
				{#snippet actions()}
					<Button
						variant="secondary"
						size="sm"
						onclick={() => {
							query = '';
							statusFilter = 'all';
						}}
					>
						Clear filters
					</Button>
				{/snippet}
			</EmptyState>
		{:else}
			<table class="w-full text-left">
				<thead>
					<tr class="border-border bg-dark-surfaceContainerLow h-9 border-b">
						<th
							scope="col"
							class="text-dark-onSurfaceVariant px-4 text-[11px] font-medium tracking-wide uppercase"
						>
							Project
						</th>
						<th
							scope="col"
							class="text-dark-onSurfaceVariant w-36 px-4 text-[11px] font-medium tracking-wide uppercase"
						>
							State
						</th>
						<th
							scope="col"
							class="text-dark-onSurfaceVariant px-4 text-[11px] font-medium tracking-wide uppercase"
						>
							Target
						</th>
						<th
							scope="col"
							class="text-dark-onSurfaceVariant w-32 px-4 text-[11px] font-medium tracking-wide uppercase"
						>
							Last deploy
						</th>
						<th scope="col" class="w-12 px-2">
							<span class="sr-only">Actions</span>
						</th>
					</tr>
				</thead>
				<tbody class="divide-border divide-y">
					{#each filtered as project (project.uuid || project.name)}
						<tr class="hover:bg-dark-surfaceContainerHigh h-11 transition-colors">
							<td class="max-w-0 px-4">
								<button
									type="button"
									onclick={() => openDetail(project)}
									class="flex w-full flex-col text-left focus-visible:outline-none"
								>
									<span
										class="text-dark-onSurface truncate text-[13px] leading-tight font-medium"
									>
										{project.name}
									</span>
									<span
										class="text-dark-onSurfaceVariant truncate text-[11px] leading-tight"
									>
										{project.type} on {project.platform}
									</span>
								</button>
							</td>
							<td class="px-4">
								<Badge tone={project.statusTone} dot>{project.status}</Badge>
							</td>
							<td class="max-w-0 px-4">
								<span class="text-dark-onSurfaceVariant block truncate text-[12px]">
									{project.target}
								</span>
							</td>
							<td
								class="text-dark-onSurfaceVariant px-4 text-[12px] whitespace-nowrap"
							>
								<span title={formatAbsolute(project.lastDeployAt)}
									>{project.lastDeployLabel}</span
								>
							</td>
							<td class="px-2 text-right">
								<Menu.Root>
									<Menu.Trigger
										class={ICON_BUTTON}
										aria-label="Actions for {project.name}"
									>
										<EllipsisVertical
											size={15}
											strokeWidth={1.75}
											aria-hidden="true"
										/>
									</Menu.Trigger>

									<Menu.Content
										align="end"
										sideOffset={4}
										class="border-border bg-dark-surfaceContainerHigh z-50 min-w-[12rem] rounded-lg border p-1"
									>
										<Menu.Item
											class={MENU_ITEM}
											onSelect={() => openDetail(project)}
										>
											Open details
										</Menu.Item>
										<Menu.Item
											class={MENU_ITEM}
											disabled={!project.accessUrl}
											onSelect={() => openAccess(project)}
										>
											<ExternalLink
												size={13}
												strokeWidth={1.75}
												aria-hidden="true"
											/>
											Open access URL
										</Menu.Item>
										<Menu.Separator class="bg-border my-1 h-px" />
										<Menu.Item
											class="{MENU_ITEM} text-dark-error"
											onSelect={() => openDetail(project)}
										>
											Terminate deployment
										</Menu.Item>
									</Menu.Content>
								</Menu.Root>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		{/if}
	</div>
</div>

<AddProjectDialog bind:open={composerOpen} oncreated={load} />
<ProjectPopup bind:open={detailOpen} project={selected} onchanged={load} />
