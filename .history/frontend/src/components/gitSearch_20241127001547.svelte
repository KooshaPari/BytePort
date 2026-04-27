<script lang="ts">
	import * as Command from '$lib/components/ui/command';
	import Icon from '@iconify/svelte';
	import type { User } from '../stores/user.ts';
	import { user } from '../stores/user';
	let userRepos = [];
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
        
	}
</script>

<Command.Root>
	<Command.Input placeholder="Type a command or search..." />
	<Command.List>
		<Command.Empty>No results found.</Command.Empty>
		<Command.Group heading="Repositories">
			<Command.Item>
				<Icon class="me-4 ms-2" icon="fa:github" />
				RepoName<Icon class="me-4 ms-2" icon="fa:lock" /></Command.Item
			>
		</Command.Group>
	</Command.List>
</Command.Root>
