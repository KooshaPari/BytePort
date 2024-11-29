<script lang="ts">
	import { Command as CommandPrimitive } from 'cmdk-sv';
	import { cn } from '$lib/utils.js';

	type $$Props = CommandPrimitive.CommandProps;

	export let value: $$Props['value'] = undefined;

	let className: string | undefined | null = undefined;
	export { className as class };
</script>

<CommandPrimitive.Root
	class={cn(
		' bg-dark-surfaceContainerLow  border-bg-dark-outlineVariant text-popover-foreground flex h-full w-full flex-col overflow-hidden rounded-md border p-2 drop-shadow-md',
		className
	)}
	bind:value
	{...$$restProps}
>
	<slot />
</CommandPrimitive.Root>
