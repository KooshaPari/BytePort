<script lang="ts">
	/**
	 * Overview: what the account has, at a glance.
	 *
	 * The 2023 dashboard rendered a 202px-wide sidebar inside the page, a 20%-tall
	 * decorative "Hello." banner, an unclickable "+" tile, and two horizontal
	 * overflow strips of 192x256px cards. Projects came from a local
	 * `getBaseUrl()` copy that called `platform()` from `@tauri-apps/plugin-os`
	 * unguarded, which throws in this shell: the promise rejected, nothing was
	 * ever awaited with a catch, and the strips stayed empty with no explanation.
	 *
	 * Navigation chrome now lives in the shell (`src/routes/home/+layout.svelte`),
	 * `src/components/projectData` owns the fetch, and all three list states (loading,
	 * failed, empty) are rendered explicitly.
	 */
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import ChevronRight from 'lucide-svelte/icons/chevron-right';
	import Box from 'lucide-svelte/icons/box';
	import Server from 'lucide-svelte/icons/server';
	import Activity from 'lucide-svelte/icons/activity';
	import CircleAlert from 'lucide-svelte/icons/circle-alert';
	import RefreshCw from 'lucide-svelte/icons/refresh-cw';
	import Plus from 'lucide-svelte/icons/plus';
	import type { Snippet } from 'svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import EmptyState from '$lib/components/ui/EmptyState.svelte';
	import AddProjectDialog from '../../components/addProjectDialog.svelte';
	import ProjectPopup from '../../components/projectPopup.svelte';
	import {
		fetchInstances,
		fetchProjects,
		watchAuth,
		type InstanceRow,
		type ProjectRow
	} from '../../components/projectData';

	let projects = $state<ProjectRow[]>([]);
	let instances = $state<InstanceRow[]>([]);
	let phase = $state<'loading' | 'ready' | 'error'>('loading');
	let errorMessage = $state('');
	let composerOpen = $state(false);
	let detailOpen = $state(false);
	let selected = $state<ProjectRow | null>(null);

	async function load() {
		phase = 'loading';
		errorMessage = '';
		try {
			const [nextProjects, nextInstances] = await Promise.all([fetchProjects(), fetchInstances()]);
			projects = nextProjects;
			instances = nextInstances;
			phase = 'ready';
		} catch (error) {
			phase = 'error';
			errorMessage =
				error instanceof Error && error.message
					? error.message
					: 'The backend did not return projects or instances.';
		}
	}

	onMount(() => watchAuth(goto, () => void load()));

	const running = $derived(instances.filter((instance) => instance.statusTone === 'primary').length);
	const attention = $derived(
		instances.filter((instance) => instance.statusTone === 'danger').length
	);

	const stats = $derived([
		{ label: 'Projects', value: projects.length, icon: Box },
		{ label: 'Instances', value: instances.length, icon: Server },
		{ label: 'Running', value: running, icon: Activity },
		{ label: 'Needs attention', value: attention, icon: CircleAlert }
	]);

	function openDetail(project: ProjectRow) {
		selected = project;
		detailOpen = true;
	}
</script>

{#snippet panel(title: string, count: string, action: Snippet | undefined, body: Snippet)}
	<section class="overflow-hidden rounded-lg border border-border bg-dark-surfaceContainer">
		<header class="flex h-12 items-center justify-between gap-3 border-b border-border px-4">
			<div class="flex items-baseline gap-2">
				<h2 class="text-[13px] font-semibold text-dark-onSurface">{title}</h2>
				<span class="text-[11px] text-dark-onSurfaceVariant">{count}</span>
			</div>
			{#if action}{@render action()}{/if}
		</header>
		{@render body()}
	</section>
{/snippet}

{#snippet skeletonRows(count: number)}
	<div class="flex flex-col" aria-busy="true" aria-label="Loading">
		{#each Array(count) as _, index (index)}
			<div class="flex h-11 items-center gap-3 border-b border-border px-4 last:border-b-0">
				<span class="h-3.5 w-40 animate-pulse rounded bg-dark-surfaceContainerHigh"></span>
				<span class="h-3.5 w-16 animate-pulse rounded bg-dark-surfaceContainerHigh"></span>
				<span class="ms-auto h-3.5 w-24 animate-pulse rounded bg-dark-surfaceContainerHigh"></span>
			</div>
		{/each}
	</div>
{/snippet}

<div class="flex flex-col gap-5">
	<dl class="grid grid-cols-2 gap-3 lg:grid-cols-4">
		{#each stats as stat (stat.label)}
			{@const StatIcon = stat.icon}
			<div class="flex items-center gap-3 rounded-lg border border-border bg-dark-surfaceContainer px-4 py-3">
				<span
					class="flex h-8 w-8 shrink-0 items-center justify-center rounded-md bg-dark-surfaceContainerHigh text-dark-onSurfaceVariant"
					aria-hidden="true"
				>
					<StatIcon size={16} strokeWidth={1.75} />
				</span>
				<div class="flex min-w-0 flex-col">
					<dt class="text-[11px] font-medium tracking-wide text-dark-onSurfaceVariant uppercase">
						{stat.label}
					</dt>
					<dd class="text-[15px] leading-tight font-semibold text-dark-onSurface">
						{phase === 'ready' ? stat.value : '—'}
					</dd>
				</div>
			</div>
		{/each}
	</dl>

	{#if phase === 'error'}
		<div class="flex items-start gap-3 rounded-lg border border-border bg-dark-surfaceContainer p-4" role="alert">
			<CircleAlert size={16} strokeWidth={1.75} class="mt-0.5 shrink-0 text-dark-error" aria-hidden="true" />
			<div class="flex min-w-0 flex-1 flex-col gap-2">
				<p class="text-[13px] font-medium text-dark-onSurface">
					Projects and instances could not be loaded
				</p>
				<p class="text-[12px] break-words text-dark-onSurfaceVariant">{errorMessage}</p>
			</div>
			<Button variant="secondary" size="sm" onclick={load}>
				<RefreshCw size={13} strokeWidth={1.75} aria-hidden="true" />
				Retry
			</Button>
		</div>
	{/if}

	{@render panel(
		'Projects',
		phase === 'ready' ? `${projects.length}` : '',
		projectActions,
		projectBody
	)}

	{@render panel(
		'Instances',
		phase === 'ready' ? `${instances.length}` : '',
		instanceActions,
		instanceBody
	)}
</div>

{#snippet projectActions()}
	<Button variant="primary" size="sm" onclick={() => (composerOpen = true)}>
		<Plus size={13} strokeWidth={2} aria-hidden="true" />
		New project
	</Button>
{/snippet}

{#snippet projectBody()}
	{#if phase === 'loading'}
		{@render skeletonRows(3)}
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
	{:else}
		<ul class="divide-y divide-border">
			{#each projects as project (project.uuid || project.name)}
				<li>
					<button
						type="button"
						onclick={() => openDetail(project)}
						class="flex h-11 w-full items-center gap-3 px-4 text-left transition-colors hover:bg-dark-surfaceContainerHigh"
					>
						<span class="flex min-w-0 flex-1 flex-col">
							<span class="truncate text-[13px] leading-tight font-medium text-dark-onSurface">
								{project.name}
							</span>
							<span class="truncate text-[11px] leading-tight text-dark-onSurfaceVariant">
								{project.target}
							</span>
						</span>
						<span class="hidden shrink-0 text-[12px] text-dark-onSurfaceVariant md:block">
							{project.lastDeployLabel}
						</span>
						<Badge tone={project.statusTone} dot>{project.status}</Badge>
						<ChevronRight
							size={14}
							strokeWidth={1.75}
							class="shrink-0 text-dark-onSurfaceVariant"
							aria-hidden="true"
						/>
					</button>
				</li>
			{/each}
		</ul>
	{/if}
{/snippet}

{#snippet instanceActions()}
	<Button variant="ghost" size="sm" disabled={phase === 'loading'} onclick={load}>
		<RefreshCw size={13} strokeWidth={1.75} aria-hidden="true" />
		Refresh
	</Button>
{/snippet}

{#snippet instanceBody()}
	{#if phase === 'loading'}
		{@render skeletonRows(2)}
	{:else if instances.length === 0}
		<EmptyState
			title="No instances yet"
			description="Instances appear here once a project has been deployed."
		>
			{#snippet icon()}
				<Server size={22} strokeWidth={1.5} aria-hidden="true" />
			{/snippet}
			{#snippet actions()}
				<Button variant="secondary" size="sm" onclick={() => goto('/home/projects')}>
					Go to projects
				</Button>
			{/snippet}
		</EmptyState>
	{:else}
		<ul class="divide-y divide-border">
			{#each instances as instance (instance.uuid || instance.name)}
				<li class="flex h-11 items-center gap-3 px-4">
					<span class="flex min-w-0 flex-1 flex-col">
						<span class="truncate text-[13px] leading-tight font-medium text-dark-onSurface">
							{instance.name}
						</span>
						<span class="truncate text-[11px] leading-tight text-dark-onSurfaceVariant">
							{instance.os}
						</span>
					</span>
					<span class="hidden shrink-0 text-[12px] text-dark-onSurfaceVariant md:block">
						{instance.resources.length}
						{instance.resources.length === 1 ? 'resource' : 'resources'}
					</span>
					<Badge tone={instance.statusTone} dot>{instance.status}</Badge>
				</li>
			{/each}
		</ul>
	{/if}
{/snippet}

<AddProjectDialog bind:open={composerOpen} oncreated={load} />
<ProjectPopup bind:open={detailOpen} project={selected} onchanged={load} />
