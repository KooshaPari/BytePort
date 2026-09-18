<script lang="ts">
	/**
	 * Monitor. Fleet status at a glance.
	 *
	 * The previous version opened with two imports of `../+layout.svelte` bound
	 * to the names `Project` and `VMInstance`, so the page typed its data
	 * against a layout component instead of the real API shapes. It then
	 * rendered every row as a look-alike card with an empty body, no status, no
	 * counts, and no loading, error or empty treatment. A monitoring view could
	 * therefore not distinguish "backend down" from "account empty" from
	 * "request still in flight", which is exactly the question it exists to
	 * answer.
	 *
	 * This version reads GET /instances and GET /projects through the shared
	 * api helpers, renders a status summary plus two dense tables, and states
	 * its own loading, error and empty conditions explicitly.
	 */
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { user, initializeUser } from '../../../stores/user';
	import { ApiError, apiFetch, getApiBaseUrl } from '$lib/api';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import EmptyState from '$lib/components/ui/EmptyState.svelte';

	// --- wire shapes ---------------------------------------------------------
	// The Go models carry no JSON tags on most fields, so keys arrive as Go
	// field names (`Name`, `Status`). Older frontend types assumed lower case.
	// Accept both: a monitoring page must not silently show blanks because a
	// key changed case.
	interface RawResource {
		Name?: string;
		name?: string;
		Status?: string;
		status?: string;
	}

	interface RawInstance {
		UUID?: string;
		uuid?: string;
		Name?: string;
		name?: string;
		Status?: string;
		status?: string;
		Resources?: RawResource[] | null;
	}

	interface RawProject {
		ID?: string;
		id?: string;
		UUID?: string;
		uuid?: string;
		Name?: string;
		name?: string;
		Platform?: string;
		platform?: string;
		Type?: string;
		type?: string;
		AccessURL?: string;
		access_url?: string;
		LastUpdated?: string;
		last_updated?: string;
		UpdatedAt?: string;
		CreatedAt?: string;
	}

	interface InstanceRow {
		id: string;
		name: string;
		status: string;
		resources: number;
		unhealthy: number;
	}

	interface ProjectRow {
		id: string;
		name: string;
		platform: string;
		type: string;
		accessUrl: string;
		updatedAt: string;
	}

	type Tone = 'neutral' | 'primary' | 'info' | 'warning' | 'danger';
	type LoadState = 'loading' | 'ready' | 'error';

	// --- state ---------------------------------------------------------------
	let loadState = $state<LoadState>('loading');
	let errorDetail = $state('');
	let instances = $state<InstanceRow[]>([]);
	let projects = $state<ProjectRow[]>([]);
	let loadedAt = $state<Date | null>(null);

	// --- status vocabulary ---------------------------------------------------
	// Matched by exact value, not substring: "unhealthy" contains "healthy".
	const ACTIVE = new Set([
		'running',
		'active',
		'healthy',
		'ready',
		'started',
		'deployed',
		'available'
	]);
	const PENDING = new Set([
		'pending',
		'starting',
		'provisioning',
		'deploying',
		'creating',
		'initializing',
		'updating'
	]);
	const STOPPED = new Set([
		'stopped',
		'terminated',
		'deleted',
		'inactive',
		'paused',
		'suspended'
	]);
	const FAILED = new Set(['failed', 'error', 'unhealthy', 'degraded', 'crashed', 'unreachable']);

	function normalizeStatus(value: string | undefined): string {
		return (value ?? '').trim().toLowerCase();
	}

	function toneForStatus(status: string): Tone {
		if (ACTIVE.has(status)) return 'primary';
		if (PENDING.has(status)) return 'warning';
		if (FAILED.has(status)) return 'danger';
		if (STOPPED.has(status)) return 'neutral';
		return 'neutral';
	}

	function labelForStatus(status: string): string {
		return status.length > 0 ? status : 'unknown';
	}

	// --- formatting ----------------------------------------------------------
	function shortId(value: string): string {
		if (!value) return 'n/a';
		return value.length > 12 ? `${value.slice(0, 8)}...` : value;
	}

	function formatTimestamp(value: string | undefined): string {
		if (!value) return 'Unknown';
		const parsed = new Date(value);
		if (Number.isNaN(parsed.getTime())) return 'Unknown';
		return new Intl.DateTimeFormat(undefined, {
			dateStyle: 'medium',
			timeStyle: 'short'
		}).format(parsed);
	}

	function countByState(rows: InstanceRow[], tone: Tone): number {
		return rows.filter((row) => toneForStatus(normalizeStatus(row.status)) === tone).length;
	}

	// --- shared class strings ------------------------------------------------
	// Hoisted so the table stays readable and so the two tables cannot drift
	// apart on column widths or header styling.
	const gridColumns =
		'grid grid-cols-[minmax(0,2fr)_130px_120px_minmax(0,1fr)] items-center gap-4';
	const gridProjectColumns =
		'grid grid-cols-[minmax(0,2fr)_110px_100px_minmax(0,1fr)_90px] items-center gap-4';
	const panel = 'overflow-hidden rounded-lg border border-border bg-dark-surfaceContainer';
	const thClass =
		'flex h-9 items-center text-[11px] font-medium tracking-wide text-dark-onSurfaceVariant uppercase';
	const tdClass = 'flex h-11 items-center';
	const skeleton = 'animate-pulse rounded bg-dark-surfaceVariant/70';

	function tileValueClass(key: string, value: number): string {
		const base = 'mt-1.5 text-[24px] font-semibold leading-none tabular-nums';
		if (key === 'failed' && value > 0) return `${base} text-dark-error`;
		if (key === 'active' && value > 0) return `${base} text-dark-primary`;
		return `${base} text-dark-onSurface`;
	}

	// --- derived views -------------------------------------------------------
	const summary = $derived([
		{ key: 'total', label: 'Instances', value: instances.length },
		{ key: 'active', label: 'Active', value: countByState(instances, 'primary') },
		{ key: 'pending', label: 'In progress', value: countByState(instances, 'warning') },
		{ key: 'failed', label: 'Needs attention', value: countByState(instances, 'danger') }
	]);

	const degradedCount = $derived(instances.reduce((total, row) => total + row.unhealthy, 0));

	// --- data ----------------------------------------------------------------
	function toInstanceRow(raw: RawInstance): InstanceRow {
		const resources = Array.isArray(raw.Resources) ? raw.Resources : [];
		const unhealthy = resources.filter((resource) => {
			const status = normalizeStatus(resource?.Status ?? resource?.status);
			return status.length > 0 && !ACTIVE.has(status);
		}).length;
		return {
			id: raw.UUID ?? raw.uuid ?? '',
			name: raw.Name ?? raw.name ?? 'Unnamed instance',
			status: labelForStatus(normalizeStatus(raw.Status ?? raw.status)),
			resources: resources.length,
			unhealthy
		};
	}

	function toProjectRow(raw: RawProject): ProjectRow {
		return {
			id: raw.UUID ?? raw.uuid ?? raw.ID ?? raw.id ?? '',
			name: raw.Name ?? raw.name ?? 'Untitled project',
			platform: raw.Platform ?? raw.platform ?? '',
			type: raw.Type ?? raw.type ?? '',
			accessUrl: raw.AccessURL ?? raw.access_url ?? '',
			updatedAt: raw.LastUpdated ?? raw.last_updated ?? raw.UpdatedAt ?? raw.CreatedAt ?? ''
		};
	}

	async function load() {
		loadState = 'loading';
		errorDetail = '';
		try {
			const [rawInstances, rawProjects] = await Promise.all([
				apiFetch<RawInstance[]>('/instances'),
				apiFetch<RawProject[]>('/projects')
			]);
			instances = (rawInstances ?? []).map(toInstanceRow);
			projects = (rawProjects ?? []).map(toProjectRow);
			loadedAt = new Date();
			loadState = 'ready';
		} catch (error) {
			// Reachability is reported once by the shell, so this stays specific
			// to the two requests this page makes.
			errorDetail =
				error instanceof ApiError
					? `The BytePort API answered with HTTP ${error.status}.`
					: 'No response from the BytePort API.';
			loadState = 'error';
		}
	}

	onMount(() => {
		let started = false;
		const unsubscribe = user.subscribe((value) => {
			if (value.status === 'unauthenticated') {
				goto('/login');
				return;
			}
			if (value.status === 'authenticated' && !started) {
				started = true;
				void load();
			}
		});
		void initializeUser(getApiBaseUrl());
		return unsubscribe;
	});
</script>

<div class="space-y-5">
	<div class="flex items-center justify-end gap-3">
		{#if loadedAt}
			<span class="text-dark-onSurfaceVariant text-[12px]">
				Updated
				{loadedAt.toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit' })}
			</span>
		{/if}
		<Button
			variant="secondary"
			size="sm"
			loading={loadState === 'loading'}
			onclick={() => void load()}
		>
			{loadState === 'loading' ? 'Refreshing' : 'Refresh'}
		</Button>
	</div>

	<div class="space-y-5">
		{#if loadState === 'error'}
			<div
				class="border-dark-error/40 bg-dark-errorContainer/40 flex items-start justify-between gap-4 rounded-lg border p-4"
				role="alert"
			>
				<div class="min-w-0">
					<p class="text-[13px] font-medium">Monitoring data unavailable</p>
					<p class="text-dark-onSurfaceVariant mt-0.5 text-[12px] leading-relaxed">
						{errorDetail}
					</p>
				</div>
				<Button variant="secondary" size="sm" onclick={() => void load()}>Retry</Button>
			</div>
		{/if}

		<section aria-label="Status summary">
			<div class="grid grid-cols-2 gap-3 lg:grid-cols-4">
				{#each summary as tile (tile.key)}
					<Card padding="md">
						<p
							class="text-dark-onSurfaceVariant text-[11px] font-medium tracking-wide uppercase"
						>
							{tile.label}
						</p>
						<p class={tileValueClass(tile.key, tile.value)}>{tile.value}</p>
					</Card>
				{/each}
			</div>
			{#if degradedCount > 0}
				<p class="text-dark-error mt-2 text-[12px]">
					{degradedCount} attached resource{degradedCount === 1 ? '' : 's'} reported a non-active
					status.
				</p>
			{/if}
		</section>

		<section class={panel}>
			<header class="flex items-center justify-between gap-3 px-4 py-3">
				<div class="min-w-0">
					<h2 class="text-[13px] font-semibold">Instances</h2>
					<p class="text-dark-onSurfaceVariant text-[12px]">
						{instances.length} instance{instances.length === 1 ? '' : 's'} from GET /instances
					</p>
				</div>
			</header>

			{#if loadState === 'loading' && instances.length === 0}
				<div class="border-border border-t px-4 py-3">
					{#each [0, 1, 2] as row (row)}
						<div class="flex h-11 items-center gap-4">
							<div class="{skeleton} h-3 w-40"></div>
							<div class="{skeleton} h-4 w-20 rounded-full"></div>
							<div class="{skeleton} h-3 w-24"></div>
						</div>
					{/each}
				</div>
			{:else if instances.length === 0}
				<div class="border-border border-t">
					<EmptyState
						title="No instances yet"
						description="Deploy a project and its instances appear here with live status."
					>
						{#snippet actions()}
							<Button
								variant="primary"
								size="sm"
								onclick={() => goto('/home/projects')}
							>
								Go to projects
							</Button>
						{/snippet}
					</EmptyState>
				</div>
			{:else}
				<div class="{gridColumns} border-border bg-dark-surfaceContainerLow border-y px-4">
					<span class={thClass}>Instance</span>
					<span class={thClass}>State</span>
					<span class={thClass}>Resources</span>
					<span class={thClass}>Instance ID</span>
				</div>
				{#each instances as row (row.id || row.name)}
					<div
						class="{gridColumns} border-border/60 hover:bg-dark-surfaceContainerHigh border-b px-4 last:border-b-0"
					>
						<span class="{tdClass} truncate text-[13px]" title={row.name}
							>{row.name}</span
						>
						<span class={tdClass}>
							<Badge tone={toneForStatus(row.status)} dot>{row.status}</Badge>
						</span>
						<span class="{tdClass} text-[13px] tabular-nums">
							{#if row.resources === 0}
								<span class="text-dark-onSurfaceVariant">None</span>
							{:else if row.unhealthy > 0}
								<span class="text-dark-error"
									>{row.resources} ({row.unhealthy} not ok)</span
								>
							{:else}
								<span>{row.resources}</span>
							{/if}
						</span>
						<span
							class="{tdClass} text-dark-onSurfaceVariant truncate font-mono text-[12px]"
							title={row.id}
						>
							{shortId(row.id)}
						</span>
					</div>
				{/each}
			{/if}
		</section>

		<section class={panel}>
			<header class="flex items-center justify-between gap-3 px-4 py-3">
				<div class="min-w-0">
					<h2 class="text-[13px] font-semibold">Deployments</h2>
					<p class="text-dark-onSurfaceVariant text-[12px]">
						{projects.length} project{projects.length === 1 ? '' : 's'} from GET /projects
					</p>
				</div>
			</header>

			{#if loadState === 'loading' && projects.length === 0}
				<div class="border-border border-t px-4 py-3">
					{#each [0, 1, 2] as row (row)}
						<div class="flex h-11 items-center gap-4">
							<div class="{skeleton} h-3 w-44"></div>
							<div class="{skeleton} h-3 w-24"></div>
						</div>
					{/each}
				</div>
			{:else if projects.length === 0}
				<div class="border-border border-t">
					<EmptyState
						title="No deployments yet"
						description="Projects you deploy are listed here with their platform, type and last update."
					>
						{#snippet actions()}
							<Button
								variant="primary"
								size="sm"
								onclick={() => goto('/home/projects')}
							>
								Go to projects
							</Button>
						{/snippet}
					</EmptyState>
				</div>
			{:else}
				<div
					class="{gridProjectColumns} border-border bg-dark-surfaceContainerLow border-y px-4"
				>
					<span class={thClass}>Project</span>
					<span class={thClass}>Platform</span>
					<span class={thClass}>Type</span>
					<span class={thClass}>Last update</span>
					<span class={thClass}>Access</span>
				</div>
				{#each projects as row (row.id || row.name)}
					<div
						class="{gridProjectColumns} border-border/60 hover:bg-dark-surfaceContainerHigh border-b px-4 last:border-b-0"
					>
						<span class="{tdClass} truncate text-[13px]" title={row.name}
							>{row.name}</span
						>
						<span class="{tdClass} text-dark-onSurfaceVariant truncate text-[13px]">
							{row.platform || 'Unknown'}
						</span>
						<span class="{tdClass} text-dark-onSurfaceVariant truncate text-[13px]">
							{row.type || 'Unknown'}
						</span>
						<span class="{tdClass} text-dark-onSurfaceVariant truncate text-[13px]">
							{formatTimestamp(row.updatedAt)}
						</span>
						<span class={tdClass}>
							{#if row.accessUrl}
								<a
									href={row.accessUrl}
									target="_blank"
									rel="noreferrer"
									class="text-dark-primary truncate text-[13px] hover:underline"
								>
									Open
								</a>
							{:else}
								<span class="text-dark-onSurfaceVariant text-[13px]">None</span>
							{/if}
						</span>
					</div>
				{/each}
			{/if}
		</section>
	</div>
</div>
