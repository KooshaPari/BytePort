<script lang="ts">
	/**
	 * Instances list.
	 *
	 * The 2023 page fetched `/instances` and then rendered nothing at all: the
	 * cards had an empty body div, so instances existed only as an unexplained
	 * image. It also imported the route layout as a component to use as a type,
	 * and redirected to `/login` on the store's initial `pending` state, so a
	 * direct deep link bounced before the session had resolved.
	 *
	 * Instances are variable-sized records (a sandbox plus its AWS resources), so
	 * this is a table with one expandable detail row per instance rather than a
	 * fixed card.
	 */
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import ChevronRight from 'lucide-svelte/icons/chevron-right';
	import RefreshCw from 'lucide-svelte/icons/refresh-cw';
	import Search from 'lucide-svelte/icons/search';
	import Server from 'lucide-svelte/icons/server';
	import CircleAlert from 'lucide-svelte/icons/circle-alert';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import EmptyState from '$lib/components/ui/EmptyState.svelte';
	import Combo from '../../../components/combo.svelte';
	import ProjectPopup from '../../../components/projectPopup.svelte';
	import {
		fetchInstances,
		fetchProjects,
		formatAbsolute,
		watchAuth,
		type InstanceRow,
		type ProjectRow
	} from '../../../components/projectData';

	const STATUS_FILTERS = [
		{ value: 'all', label: 'All states' },
		{ value: 'primary', label: 'Running' },
		{ value: 'info', label: 'In progress' },
		{ value: 'danger', label: 'Failed' },
		{ value: 'neutral', label: 'Stopped' }
	];

	let instances = $state<InstanceRow[]>([]);
	let projects = $state<ProjectRow[]>([]);
	let phase = $state<'loading' | 'ready' | 'error'>('loading');
	let errorMessage = $state('');
	let query = $state('');
	let statusFilter = $state('all');
	let expanded = $state(new Set<string>());
	let detailOpen = $state(false);
	let selected = $state<ProjectRow | null>(null);

	async function load() {
		if (phase !== 'ready') phase = 'loading';
		errorMessage = '';
		try {
			const [nextInstances, nextProjects] = await Promise.all([fetchInstances(), fetchProjects()]);
			instances = nextInstances;
			projects = nextProjects;
			phase = 'ready';
		} catch (error) {
			phase = 'error';
			errorMessage =
				error instanceof Error && error.message
					? error.message
					: 'The backend did not return an instances list.';
		}
	}

	onMount(() => watchAuth(goto, () => void load()));

	const projectsByUuid = $derived(
		new Map(projects.map((project) => [project.uuid, project] as const))
	);

	function projectName(instance: InstanceRow): string {
		return projectsByUuid.get(instance.projectUuid)?.name ?? 'Unassigned';
	}

	const filtered = $derived.by(() => {
		const needle = query.trim().toLowerCase();
		return instances.filter((instance) => {
			if (statusFilter !== 'all' && instance.statusTone !== statusFilter) return false;
			if (!needle) return true;
			return [instance.name, instance.os, instance.status, projectName(instance)]
				.join(' ')
				.toLowerCase()
				.includes(needle);
		});
	});

	function readValue(event: Event) {
		query = (event.currentTarget as HTMLInputElement).value;
	}

	function toggle(key: string) {
		const next = new Set(expanded);
		if (next.has(key)) next.delete(key);
		else next.add(key);
		expanded = next;
	}

	function openProject(instance: InstanceRow) {
		const project = projectsByUuid.get(instance.projectUuid);
		if (!project) return;
		selected = project;
		detailOpen = true;
	}
</script>

<div class="flex flex-col gap-4">
	<div class="flex flex-wrap items-end gap-2">
		<div class="w-64">
			<Input
				label="Search"
				placeholder="Name, project or state"
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
				<span class="text-[11px] text-dark-onSurfaceVariant">
					{filtered.length} of {instances.length}
				</span>
			{/if}
			<Button variant="secondary" size="sm" disabled={phase === 'loading'} onclick={load}>
				<RefreshCw size={13} strokeWidth={1.75} aria-hidden="true" />
				Refresh
			</Button>
		</div>
	</div>

	{#if phase === 'error'}
		<div class="flex items-start gap-3 rounded-lg border border-border bg-dark-surfaceContainer p-4" role="alert">
			<CircleAlert size={16} strokeWidth={1.75} class="mt-0.5 shrink-0 text-dark-error" aria-hidden="true" />
			<div class="flex min-w-0 flex-1 flex-col gap-1">
				<p class="text-[13px] font-medium text-dark-onSurface">Instances could not be loaded</p>
				<p class="text-[12px] break-words text-dark-onSurfaceVariant">{errorMessage}</p>
			</div>
			<Button variant="secondary" size="sm" onclick={load}>Retry</Button>
		</div>
	{/if}

	<div class="overflow-hidden rounded-lg border border-border bg-dark-surfaceContainer">
		{#if phase === 'loading'}
			<div aria-busy="true" aria-label="Loading instances">
				{#each Array(4) as _, index (index)}
					<div class="flex h-11 items-center gap-3 border-b border-border px-4 last:border-b-0">
						<span class="h-3.5 w-40 animate-pulse rounded bg-dark-surfaceContainerHigh"></span>
						<span class="h-3.5 w-20 animate-pulse rounded bg-dark-surfaceContainerHigh"></span>
						<span class="ms-auto h-3.5 w-24 animate-pulse rounded bg-dark-surfaceContainerHigh"></span>
					</div>
				{/each}
			</div>
		{:else if instances.length === 0}
			<EmptyState
				title="No instances yet"
				description="Instances appear here once a project has been deployed. Deploy one from the projects list."
			>
				{#snippet icon()}
					<Server size={22} strokeWidth={1.5} aria-hidden="true" />
				{/snippet}
				{#snippet actions()}
					<Button variant="primary" size="sm" onclick={() => goto('/home/projects')}>
						Go to projects
					</Button>
				{/snippet}
			</EmptyState>
		{:else if filtered.length === 0}
			<EmptyState
				title="No instances match"
				description="Clear the search or pick a different state to see more instances."
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
					<tr class="h-9 border-b border-border bg-dark-surfaceContainerLow">
						<th scope="col" class="w-10 px-2"><span class="sr-only">Expand</span></th>
						<th scope="col" class="px-4 text-[11px] font-medium tracking-wide text-dark-onSurfaceVariant uppercase">
							Instance
						</th>
						<th scope="col" class="w-36 px-4 text-[11px] font-medium tracking-wide text-dark-onSurfaceVariant uppercase">
							State
						</th>
						<th scope="col" class="px-4 text-[11px] font-medium tracking-wide text-dark-onSurfaceVariant uppercase">
							Project
						</th>
						<th scope="col" class="w-28 px-4 text-[11px] font-medium tracking-wide text-dark-onSurfaceVariant uppercase">
							Resources
						</th>
						<th scope="col" class="w-32 px-4 text-[11px] font-medium tracking-wide text-dark-onSurfaceVariant uppercase">
							Updated
						</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-border">
					{#each filtered as instance (instance.uuid || instance.name)}
						{@const key = instance.uuid || instance.name}
						{@const open = expanded.has(key)}
						<tr class="h-11 transition-colors hover:bg-dark-surfaceContainerHigh">
							<td class="px-2">
								<button
									type="button"
									onclick={() => toggle(key)}
									aria-expanded={open}
									aria-label={open ? `Collapse ${instance.name}` : `Expand ${instance.name}`}
									class="flex h-7 w-7 items-center justify-center rounded-md text-dark-onSurfaceVariant
										transition-colors hover:bg-dark-surfaceContainerHighest hover:text-dark-onSurface
										focus-visible:outline-none"
								>
									<ChevronRight
										size={14}
										strokeWidth={1.75}
										class="transition-transform {open ? 'rotate-90' : ''}"
										aria-hidden="true"
									/>
								</button>
							</td>
							<td class="max-w-0 px-4">
								<span class="block truncate text-[13px] leading-tight font-medium text-dark-onSurface">
									{instance.name}
								</span>
								<span class="block truncate text-[11px] leading-tight text-dark-onSurfaceVariant">
									{instance.os}
								</span>
							</td>
							<td class="px-4">
								<Badge tone={instance.statusTone} dot>{instance.status}</Badge>
							</td>
							<td class="max-w-0 px-4">
								{#if projectsByUuid.has(instance.projectUuid)}
									<button
										type="button"
										onclick={() => openProject(instance)}
										class="block max-w-full truncate text-left text-[12px] text-dark-primary hover:underline focus-visible:outline-none"
									>
										{projectName(instance)}
									</button>
								{:else}
									<span class="block truncate text-[12px] text-dark-onSurfaceVariant">
										{projectName(instance)}
									</span>
								{/if}
							</td>
							<td class="px-4 text-[12px] text-dark-onSurfaceVariant">
								{instance.resources.length}
							</td>
							<td class="px-4 text-[12px] whitespace-nowrap text-dark-onSurfaceVariant">
								<span title={formatAbsolute(instance.lastUpdated)}>
									{instance.lastUpdatedLabel}
								</span>
							</td>
						</tr>

						{#if open}
							<tr class="bg-dark-surfaceContainerLow">
								<td></td>
								<td colspan="5" class="px-4 py-3">
									{#if instance.resources.length === 0}
										<p class="text-[12px] text-dark-onSurfaceVariant">
											No resources reported for this instance.
										</p>
									{:else}
										<ul class="divide-y divide-border overflow-hidden rounded-md border border-border">
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
													<span
														class="hidden min-w-0 max-w-[18rem] truncate text-[11px] text-dark-onSurfaceVariant lg:block"
														title={resource.arn}
													>
														{resource.region || resource.type}
													</span>
													<Badge tone={resource.statusTone}>{resource.status}</Badge>
												</li>
											{/each}
										</ul>
									{/if}
								</td>
							</tr>
						{/if}
					{/each}
				</tbody>
			</table>
		{/if}
	</div>
</div>

<ProjectPopup bind:open={detailOpen} project={selected} onchanged={load} />
