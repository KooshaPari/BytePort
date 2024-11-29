<script lang="ts">
	import * as Command from '$lib/components/ui/command';
	import Icon from '@iconify/svelte';
	import type { User } from '../stores/user.ts';
	import { user } from '../stores/user';
    import { onMount, onDestroy } from 'svelte';
    interace GitRepo {
        name: string;
        description: string;
        url: string;
        language: string;
        stargazers_count: number;
        forks_count: number;
        private: boolean;
    }
	let userRepos: GitRepo = [];
	let client: User | null = null;

	// read in user store and set Client
	const unsubscribe = user.subscribe((value) => {
		client = value.data;
	});
	async function getRepos() {
		// fetch repos from backend and set userRepos
		const response = await fetch('http://localhost:8081/api/github/repos', {
			method: 'GET',
			headers: {
				'Content-Type': 'application/json'
			},
			credentials: 'include'
		});
		const data = await response.json();
		userRepos = data;
		console.log('userRepos:', userRepos);
	}
</script>

<Command.Root>
	<Command.Input placeholder="Type a command or search..." />
	<Command.List>
		<Command.Empty>No results found.</Command.Empty>
		<Command.Group heading="Repositories">
			{#each userRepos as repo}
				<Command.Item>
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
