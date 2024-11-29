<script lang="ts">
	import * as Command from '$lib/components/ui/command';
	import Icon from '@iconify/svelte';
	import type { User } from '../stores/user.ts';
	let userRepos = [];
	let Client: User;
    

    // read in user store and set Client
    
    function getRepos() {
        // fetch repos from backend and set userRepos
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
