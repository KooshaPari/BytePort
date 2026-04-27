<script lang="ts">
	import * as Command from '$lib/components/ui/command';
	import Icon from '@iconify/svelte';
    let userRepos
</script>

<Command.Root>
	<Command.Input placeholder="Type a command or search..." />
	<Command.List>
		<Command.Empty>No results found.</Command.Empty>
		<Command.Group heading="Repositories">
			<Command.Item>
				<Icon class="me-4 ms-2" icon="fa:github" />
				RepoName<Icon class="me-4 ms-2" icon="fa:lock"/></Command.Item
			>
		</Command.Group>
	</Command.List>
</Command.Root>
