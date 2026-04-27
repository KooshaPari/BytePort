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

	// read in user store and set Client
	const unsubscribe = user.subscribe((value) => {
		client = value.data;
	});
	onMount(async () => {
		userRepos = await fetchUserRepositories();
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

			console.log('data:', data);

			return data;
		} catch (error) {
			console.error('Failed to fetch repositories:', error);
			return [];
		}
	}

	async function testPart(test: String) {
		let reps = JSON.parse(test as string) as Repository[];

		console.log('reps:', reps);
		return reps;
	}
</script>

<Command.Root>
	<Command.Input placeholder="Type a command or search..." />
	<Command.List>
		<Command.Empty>No results found.</Command.Empty>
		<Command.Group heading="Repositories">
			{#each userRepos as repo}
				<Command.Item va>
					<Avatar.Root>
						<Avatar.Image src={repo.owner.avatar_url} alt={repo.full_name} />
						<Avatar.Fallback>CN</Avatar.Fallback>
					</Avatar.Root>
					<Icon class="me-4 ms-2" icon="fa:github" />
					{repo.name}
					{#if repo.private}
						<Icon class="me-4 ms-2" icon="fa:lock" />
					{:else}
						<Icon class="me-4 ms-2" icon="fa:globe" />
					{/if}
				</Command.Item>
			{/each}
		</Command.Group>
	</Command.List>
</Command.Root>
