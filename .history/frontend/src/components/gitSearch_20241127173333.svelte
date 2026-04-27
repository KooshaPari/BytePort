<script lang="ts">
	import * as Command from '$lib/components/ui/command';
	import Icon from '@iconify/svelte';
	import type { User } from '../stores/user.ts';
	import { user } from '../stores/user';
	import { onMount, onDestroy } from 'svelte';
	import type { Repository } from '../lib/git';

	import * as Avatar from '$lib/components/ui/avatar';

	let userRepos: Repository[];

	let client: User | null = null;
	const CACHE_KEY = 'user_repositories';
	const CACHE_DURATION = 1000 * 60 * 60;

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

	// Function to manually refresh repos
	async function refreshRepos() {
		await fetchAndCacheRepos();
	}
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

			console.log('data:', data);

			return data;
		} catch (error) {
			console.error('Failed to fetch repositories:', error);
			return [];
		}
	}
Command.</script>

<Command.Root>
	<Command.Input placeholder="Type a command or search..." />
	<Command.List>
		{#if loading}
			<Command.Loading>Fetching Repos...</Command.Loading>
		{:else}
			<Command.Empty>No results found.</Command.Empty>
			<Command.Group heading="Repositories">
				<Command.Item
					icon={Avatar}
					value="OpenAI"
					description="OpenAI API">OpenAI</Command.Item>
					
			</Command.Group>
		{/if}
	</Command.List>
</Command.Root>
