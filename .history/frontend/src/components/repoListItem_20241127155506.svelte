<script>
	import * as Command from '$lib/components/ui/command';
	import * as Avatar from '$lib/components/ui/avatar';
	import Icon from 'svelte-icons-pack/Icon.svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { Root } from '$lib/components/ui/button';
</script>

<Command.Item class="my-1" value={repo.name}>
	<Avatar.Root class="me-4 ms-2" delayMs={1000}>
		<Avatar.Image src={repo.owner.avatar_url} alt={repo.full_name} />
		<Avatar.Fallback><Icon class="me-4 ms-2" icon="fa:user" /></Avatar.Fallback>
	</Avatar.Root>
	<Icon class="me-4 ms-2" icon="fa:github" />
	{repo.name}
	{#if repo.private}
		<Icon class="me-4 ms-2" icon="fa:lock" />
	{:else}
		<Icon class="me-4 ms-2" icon="fa:globe" />
	{/if}</Command.Item
>
