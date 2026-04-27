<script>
    import * as Command from "$lib/components/ui/command";
</script>
<Command.Item class="my-1" value={repo.name}></Command.Item>
					<Avatar.Image src={repo.owner.avatar_url} alt={repo.full_name} />
					<Avatar.Fallback><Icon class="me-4 ms-2" icon="fa:user" /></Avatar.Fallback>
				</Avatar.Root>
				<Icon class="me-4 ms-2" icon="fa:github" />
				{repo.name}
				{#if repo.private}
					<Icon class="me-4 ms-2" icon="fa:lock" />
				{:else}
					<Icon class="me-4 ms-2" icon="fa:globe" />
				{/if}</Command.Item>