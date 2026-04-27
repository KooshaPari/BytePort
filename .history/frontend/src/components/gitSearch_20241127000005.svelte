<script lang="ts">
	import * as Command from '$lib/components/ui/command';
</script>

<Command.Root>
	<Command.Input placeholder="Type a command or search..." />
	<Command.List>
		<Command.Empty>No results found.</Command.Empty>
		<Command.Group heading="Repositories">
			<Command.Item>
            <
                Calendar
            </Command.Item>
		</Command.Group>
	</Command.List>
</Command.Root>
