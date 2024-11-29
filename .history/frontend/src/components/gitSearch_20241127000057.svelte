<script lang="ts">
	import * as Command from '$lib/components/ui/command';
	import  from IconifyIconComponent
	import { Icon } from 'iconify-icon/dist/iconify-icon.js';
</script>

<Command.Root>
	<Command.Input placeholder="Type a command or search..." />
	<Command.List>
		<Command.Empty>No results found.</Command.Empty>
		<Command.Group heading="Repositories">
			<Command.Item>
				<Icon icon="fa:github" />
				RepoName</Command.Item
			>
		</Command.Group>
	</Command.List>
</Command.Root>
