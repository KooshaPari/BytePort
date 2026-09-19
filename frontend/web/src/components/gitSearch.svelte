<script lang="ts">
	/**
	 * Repository picker for step 1 of the new-project wizard.
	 *
	 * The 2023 version wrapped `cmdk`'s command list and rendered whatever came
	 * back, with no error state: when `GET /api/github/repositories` failed (or the
	 * session had expired) the list was simply empty and the wizard had no way
	 * forward. It also had a `setRepo`/`dispatch` pair that logged "dispatching"
	 * and did nothing.
	 *
	 * This version keeps the same endpoint and the same localStorage cache, and
	 * adds the three states it was missing: loading, failed-with-retry, and
	 * genuinely empty.
	 */
	import Search from 'lucide-svelte/icons/search';
	import RefreshCw from 'lucide-svelte/icons/refresh-cw';
	import FolderGit2 from 'lucide-svelte/icons/folder-git-2';
	import Check from 'lucide-svelte/icons/check';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import EmptyState from '$lib/components/ui/EmptyState.svelte';
	import { ApiError, apiFetch } from '$lib/api';
	import type { Repository } from '$lib/git';

	let {
		selected = null,
		disabled = false,
		onselect
	}: {
		selected?: Repository | null;
		disabled?: boolean;
		onselect: (repo: Repository) => void;
	} = $props();

	const CACHE_KEY = 'user_repositories';
	const CACHE_DURATION = 1000 * 60 * 60;

	let repos = $state<Repository[]>([]);
	let phase = $state<'loading' | 'ready' | 'error'>('loading');
	let errorMessage = $state('');
	let query = $state('');
	let highlight = $state(0);
	let failedAvatars = $state(new Set<string>());

	function readCache(): Repository[] | null {
		try {
			const cached = localStorage.getItem(CACHE_KEY);
			if (!cached) return null;
			const { timestamp, data } = JSON.parse(cached) as { timestamp: number; data: unknown };
			if (Date.now() - timestamp > CACHE_DURATION) return null;
			return Array.isArray(data) ? (data as Repository[]) : null;
		} catch {
			// A corrupt cache entry must not break the picker.
			return null;
		}
	}

	function writeCache(value: Repository[]) {
		try {
			localStorage.setItem(CACHE_KEY, JSON.stringify({ timestamp: Date.now(), data: value }));
		} catch {
			// Storage can be unavailable (private mode, quota); the list still works.
		}
	}

	async function load({ useCache = true }: { useCache?: boolean } = {}) {
		if (useCache) {
			const cached = readCache();
			if (cached && cached.length > 0) {
				repos = cached;
				phase = 'ready';
				return;
			}
		}

		phase = 'loading';
		errorMessage = '';
		try {
			const data = await apiFetch<Repository[]>('/api/github/repositories');
			const list = Array.isArray(data) ? data : [];
			repos = list;
			writeCache(list);
			phase = 'ready';
		} catch (error) {
			phase = 'error';
			errorMessage =
				error instanceof ApiError && error.status === 401
					? 'Your session expired. Sign in again to list repositories.'
					: 'Could not reach the BytePort backend to list your repositories.';
		}
	}

	$effect(() => {
		void load();
	});

	const filtered = $derived.by(() => {
		const needle = query.trim().toLowerCase();
		if (!needle) return repos;
		return repos.filter((repo) =>
			[repo.full_name, repo.name, repo.language ?? '', repo.description ?? '']
				.join(' ')
				.toLowerCase()
				.includes(needle)
		);
	});

	$effect(() => {
		// Keep the keyboard highlight inside the filtered list.
		if (highlight > filtered.length - 1) highlight = Math.max(filtered.length - 1, 0);
	});

	function initials(repo: Repository): string {
		const source = repo.owner?.login || repo.name || '?';
		return source.slice(0, 2).toUpperCase();
	}

	function onkeydown(event: KeyboardEvent) {
		if (disabled || filtered.length === 0) return;
		if (event.key === 'ArrowDown') {
			event.preventDefault();
			highlight = Math.min(highlight + 1, filtered.length - 1);
		} else if (event.key === 'ArrowUp') {
			event.preventDefault();
			highlight = Math.max(highlight - 1, 0);
		} else if (event.key === 'Enter') {
			event.preventDefault();
			const repo = filtered[highlight];
			if (repo) onselect(repo);
		}
	}
	function readValue(event: Event): string {
		return (event.currentTarget as HTMLInputElement).value;
	}
</script>

<div class="flex flex-col gap-3">
	<Input
		label="Repository"
		placeholder="Search your GitHub repositories"
		autocomplete="off"
		disabled={disabled || phase !== 'ready'}
		value={query}
		oninput={(event) => (query = readValue(event))}
		{onkeydown}
	/>

	<div
		class="border-border bg-dark-surfaceContainerLowest h-64 overflow-hidden rounded-lg border"
	>
		{#if phase === 'loading'}
			<div
				class="flex h-full flex-col gap-1 p-2"
				aria-busy="true"
				aria-label="Loading repositories"
			>
				{#each [0, 1, 2, 3, 4] as row (row)}
					<div class="flex h-11 items-center gap-2.5 rounded-md px-2.5">
						<div
							class="bg-dark-surfaceContainerHigh h-5 w-5 shrink-0 animate-pulse rounded"
						></div>
						<div
							class="bg-dark-surfaceContainerHigh h-3 flex-1 animate-pulse rounded"
							style="max-width: {60 - row * 8}%"
						></div>
					</div>
				{/each}
			</div>
		{:else if phase === 'error'}
			<div class="flex h-full flex-col items-center justify-center gap-3 px-6 text-center">
				<p class="text-dark-onSurfaceVariant text-[13px]">{errorMessage}</p>
				<Button variant="secondary" size="sm" onclick={() => load({ useCache: false })}>
					<RefreshCw size={13} strokeWidth={1.75} aria-hidden="true" />
					Retry
				</Button>
			</div>
		{:else if repos.length === 0}
			<EmptyState
				title="No repositories found"
				description="Connect GitHub in Settings to deploy from a repository."
			>
				{#snippet icon()}
					<FolderGit2 size={22} strokeWidth={1.5} aria-hidden="true" />
				{/snippet}
			</EmptyState>
		{:else if filtered.length === 0}
			<EmptyState title="No matches" description="No repository matches “{query}”." />
		{:else}
			<ul class="h-full overflow-y-auto p-1">
				{#each filtered as repo, index (repo.id)}
					{@const isSelected = selected?.full_name === repo.full_name}
					<li>
						<button
							type="button"
							aria-pressed={isSelected}
							onclick={() => onselect(repo)}
							onmousemove={() => (highlight = index)}
							class="flex h-11 w-full items-center gap-2.5 rounded-md px-2 text-left transition-colors
								{index === highlight && !isSelected ? 'bg-dark-surfaceContainerHigh' : ''}
								{isSelected ? 'bg-dark-primaryContainer/40' : ''}"
						>
							<span
								class="bg-dark-surfaceContainerHighest text-dark-onSurfaceVariant flex h-5 w-5 shrink-0 items-center justify-center
									overflow-hidden rounded text-[9px] font-semibold"
								aria-hidden="true"
							>
								{#if repo.owner?.avatar_url && !failedAvatars.has(repo.full_name)}
									<img
										src={repo.owner.avatar_url}
										alt=""
										class="h-full w-full object-cover"
										onerror={() => failedAvatars.add(repo.full_name)}
									/>
								{:else}
									{initials(repo)}
								{/if}
							</span>

							<span class="flex min-w-0 flex-1 flex-col">
								<span
									class="text-dark-onSurface truncate text-[13px] leading-tight font-medium"
								>
									{repo.full_name}
								</span>
								<span
									class="text-dark-onSurfaceVariant truncate text-[11px] leading-tight"
								>
									{repo.description || 'No description'}
								</span>
							</span>

							{#if repo.language}
								<span class="text-dark-onSurfaceVariant shrink-0 text-[11px]"
									>{repo.language}</span
								>
							{/if}
							{#if repo.private}
								<Badge tone="neutral">Private</Badge>
							{/if}
							{#if isSelected}
								<Check
									size={14}
									strokeWidth={2}
									class="text-dark-primary shrink-0"
									aria-hidden="true"
								/>
							{/if}
						</button>
					</li>
				{/each}
			</ul>
		{/if}
	</div>

	<p class="text-dark-onSurfaceVariant text-[11px]">
		{repos.length}
		{repos.length === 1 ? 'repository' : 'repositories'} available
	</p>
</div>
