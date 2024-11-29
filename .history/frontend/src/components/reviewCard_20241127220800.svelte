<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Card from '$lib/components/ui/card/index.js';
	import * as Select from '$lib/components/ui/select/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { Avatar } from '$lib/components/ui/avatar/index.js';
	import Icon from '@iconify/svelte';
	let props = $props();
	let repo = props.project;
	const frameworks = [
		{
			value: 'sveltekit',
			label: 'SvelteKit'
		},
		{
			value: 'next',
			label: 'Next.js'
		},
		{
			value: 'astro',
			label: 'Astro'
		},
		{
			value: 'nuxt',
			label: 'Nuxt.js'
		}
	];
</script>

<Card.Root class="w-[350px]">
	<Card.Header>
		<Card.Title>Create project</Card.Title>
		<Card.Description>Deploy your new project in one-click.</Card.Description>
	</Card.Header>
	<Card.Content>
		<form>
			<div class="grid w-full items-center gap-4">
				<div class="flex flex-col space-y-1.5">
					<Label for="name">Name</Label>
				</div>
				<div class="flex flex-col space-y-1.5">
					<Label for="framework">Framework</Label>
					<Avatar.Root class="me-4 ms-2" delayMs={1000}>
						<Avatar.Image src={repo.owner.avatar_url} alt={repo.full_name} />
						<Avatar.Fallback><Icon class="me-4 ms-2" icon="fa:user" /></Avatar.Fallback>
					</Avatar.Root>
					<Icon class="me-4 ms-2" icon="fa:github" />
					<span>{repo.name}</span>
					{#if repo.private}
						<Icon class="me-4 ms-2" icon="fa:lock" />
					{:else}
						<Icon class="me-4 ms-2" icon="fa:globe" />
					{/if}

					<Icon class="me-4 ms-2" icon="fa:star" />
					{repo.stargazers_count}
					<Icon
						class="me-4 ms-2"
						icon="
iconoir:git-fork"
					/>
					{repo.forks_count}
					<Icon class="me-4 ms-2" icon="fa:code" />
					{repo.language}
				</div>
			</div>
		</form>
	</Card.Content>
	<Card.Footer class="flex justify-between">
		<Button variant="outline">Cancel</Button>
		<Button>Deploy</Button>
	</Card.Footer>
</Card.Root>
