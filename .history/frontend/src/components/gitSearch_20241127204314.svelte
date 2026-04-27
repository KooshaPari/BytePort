<script lang="ts">
	import * as Command from '$lib/components/ui/command';
	import Icon from '@iconify/svelte';
	import type { User } from '../stores/user.ts';
	import { user } from '../stores/user';
	import { onMount, onDestroy } from 'svelte';
	import type { Repository } from '../lib/git';
	//import { setStage } from './addProjectDialog.svelte';
	import { createEventDispatcher } from 'svelte';

	import * as Avatar from '$lib/components/ui/avatar';
	type EventDetail = {
		selectedItem: string;
	};

	const dispatch = createEventDispatcher<EventDetail>();
	let userRepos: Repository[] = [];
	let client: User | null = null;
	const CACHE_KEY = 'user_repositories';
	const CACHE_DURATION = 1000 * 60 * 60;
	function setRepo(repo: Repository) {
		console.log('dispatching');
		dispatch('select', { repo: repo });
	}
	// read in user store and set Client
	async function fetchAndCacheRepos() {
		const repos = await fetchUserRepositories();
		userRepos = repos;
		localStorage.setItem(
			CACHE_KEY,
			JSON.stringify({
				timestamp: Date.now(),
				data: repos
			})
		);
	}

	function getCachedRepos() {
		const cached = localStorage.getItem(CACHE_KEY);
		if (cached) {
			const { timestamp, data } = JSON.parse(cached);
			if (Date.now() - timestamp < CACHE_DURATION) {
				return data;
			}
		}
		return null;
	}
	let loading = false;
	onMount(async () => {
		loading = true;
		const cachedRepos = getCachedRepos();
		if (cachedRepos) {
			userRepos = cachedRepos;
		} else {
			await fetchAndCacheRepos();
		}
		loading = false;
	});

	async function fetchUserRepositories(): Promise<Repository[]> {
		try {
			const response = await fetch('http://localhost:8081/api/github/repositories', {
				method: 'GET',
				headers: {
					'Content-Type': 'application/json'
				},
				credentials: 'include'
			});

			if (!response.ok) {
				throw new Error(`Error fetching repositories: ${response.statusText}`);
			}

			const rawData = await response.json();
			console.log('rawData:', rawData);
			let data: Repository[] = rawData as Repository[];

			return data;
		} catch (error) {
			console.error('Failed to fetch repositories:', error);
			return [];
		}
	}
</script>

<Command.Root>
	<Command.Input placeholder="Type a command or search..." />
	<Command.List
		class="scrollbar-thin scrollbar-thumb-gray-700 scrollbar-track-transparent hover:scrollbar-thumb-gray-600"
	>
		{#if loading}
			<Command.Loading progress={0.5}>Fetching Repos...</Command.Loading>
		{:else}
			<Command.Empty>No results found.</Command.Empty>
			<Command.Group heading="Repositories">
				{#each userRepos as repo}
					<Command.Item
						class="cursor-pointer"
						value={repo.name}
						on:select={() => {
							setRepo(repo);
						}}
					>
						<Avatar.Root class="me-4 ms-2" delayMs={1000}>
							<Avatar.Image src={repo.owner.avatar_url} alt={repo.full_name} />
							<Avatar.Fallback><Icon class="me-4 ms-2" icon="fa:user" /></Avatar.Fallback>
						</Avatar.Root>
						<Icon class="me-4 ms-2" icon="fa:github" />
						<span>{repo.name}</span>
						{#if repo.private}
							<Icon class="me-4 ms-2" icon="fa:lock" />
						{:else}
							<Icon class="me-4 ms-2" icon="fa:globe" />
						{/if}

						<Icon class="me-4 ms-2" icon="fa:star" />
						{repo.stargazers_count}
						<Icon
							class="me-4 ms-2"
							icon="
iconoir:git-fork"
						/>
						{repo.forks_count}
						<Icon class="me-4 ms-2" icon="fa:code" />
						{repo.language}
					</Command.Item>
				{/each}
			</Command.Group>
		{/if}
	</Command.List>
</Command.Root>
