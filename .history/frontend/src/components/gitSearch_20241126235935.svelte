<script lang="ts">
	import * as Command from '$lib/components/ui/command';
</script>

<Command.Root>
	<Command.Input placeholder="Type a command or search..." />
	<Command.List>
		<Command.Empty>No results found.</Command.Empty>
		<Command.Group>
			<Command.Item>Calendar</Command.Item>
			<Command.Item>Search Emoji</Command.Item>
			<Command.Item>Calculator</Command.Item>
		</Command.Group>

	</Command.List>
</Command.Root>
