<script lang="ts">
	import { Input } from '$lib/components/ui/input';
	import * as Command from '$lib/components/ui/command';
	import * as Popover from '$lib/components/ui/popover';
	import { buttonVariants } from '$lib/components/ui/button';
	import Icon from '@iconify/svelte';
	import type { Repository } from '$lib/git';
	import { onMount } from 'svelte';

	export let select: (repo: Repository) => void;

	let searchQuery = '';
	let repositories: Repository[] = [];
	let isLoading = false;
	let isOpen = false;
	let selectedRepo: Repository | null = null;

	async function searchRepositories() {
		if (!searchQuery.trim()) {
			repositories = [];
			return;
		}

		isLoading = true;
		try {
			// Call the backend to search for repositories
			// This assumes the backend has a /search endpoint
			const baseUrl = 'http://localhost:8081';
			const response = await fetch(`${baseUrl}/git/search?q=${encodeURIComponent(searchQuery)}`, {
				method: 'GET',
				headers: {
					'Content-Type': 'application/json'
				},
				credentials: 'include'
			});

			if (response.ok) {
				const data = await response.json();
				repositories = data || [];
			} else {
				repositories = [];
				console.error('Failed to search repositories');
			}
		} catch (error) {
			console.error('Error searching repositories:', error);
			repositories = [];
		} finally {
			isLoading = false;
		}
	}

	function handleSelectRepo(repo: Repository) {
		selectedRepo = repo;
		isOpen = false;
		select(repo);
	}

	onMount(async () => {
		// Try to fetch user's repositories on mount
		try {
			const baseUrl = 'http://localhost:8081';
			const response = await fetch(`${baseUrl}/git/repos`, {
				method: 'GET',
				headers: {
					'Content-Type': 'application/json'
				},
				credentials: 'include'
			});

			if (response.ok) {
				const data = await response.json();
				repositories = data || [];
			}
		} catch (error) {
			console.error('Error fetching repositories:', error);
		}
	});
</script>

<div class="w-full max-w-md space-y-2">
	<label for="git-search" class="text-sm font-medium text-white block mb-2">Select Repository</label>

	<Popover.Root bind:open={isOpen}>
		<Popover.Trigger
			role="combobox"
			class={`${buttonVariants({ variant: 'outline' })} w-full justify-between`}
		>
			{#if selectedRepo}
				<span class="flex items-center gap-2">
					<Icon icon="mdi:github" />
					{selectedRepo.name}
				</span>
			{:else}
				<span class="text-gray-400">Search or select a repository...</span>
			{/if}
			<Icon icon="mdi:chevron-down" class="ml-2" />
		</Popover.Trigger>

		<Popover.Content class="w-full p-0" side="bottom" align="start">
			<Command.Root>
				<div class="flex items-center border-b px-3">
					<Icon icon="mdi:magnify" class="mr-2 text-gray-400" />
					<Input
						id="git-search"
						placeholder="Search repositories..."
						bind:value={searchQuery}
						on:input={searchRepositories}
						class="border-0 focus:ring-0"
					/>
				</div>

				<Command.List class="max-h-[200px]">
					{#if isLoading}
						<div class="flex items-center justify-center py-6">
							<Icon icon="mdi:loading" class="animate-spin text-gray-400" />
							<span class="ml-2 text-gray-400">Searching...</span>
						</div>
					{:else if repositories.length === 0}
						<div class="flex flex-col items-center justify-center py-6 text-gray-400">
							<Icon icon="mdi:folder-open" class="text-2xl mb-2" />
							<p>No repositories found</p>
						</div>
					{:else}
						{#each repositories as repo (repo.id)}
							<Command.Item
								value={repo.name}
								onSelect={() => handleSelectRepo(repo)}
								class="flex items-center gap-2 py-2"
							>
								<Icon icon="mdi:repository" />
								<div class="flex-1">
									<p class="font-medium">{repo.full_name}</p>
									<p class="text-xs text-gray-400">
										{repo.description || 'No description'}
									</p>
								</div>
								{#if repo.private}
									<Icon icon="mdi:lock" class="text-gray-400" />
								{:else}
									<Icon icon="mdi:lock-open" class="text-gray-400" />
								{/if}
							</Command.Item>
						{/each}
					{/if}
				</Command.List>
			</Command.Root>
		</Popover.Content>
	</Popover.Root>
</div>

<style>
	:global(.text-dark-onSurfaceVariant) {
		color: rgba(255, 255, 255, 0.7);
	}
</style>
