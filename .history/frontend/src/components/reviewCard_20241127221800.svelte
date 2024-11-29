<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Card from '$lib/components/ui/card/index.js';
	import * as Select from '$lib/components/ui/select/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import * as Avatar from '$lib/components/ui/avatar/index.js';
	import Icon from '@iconify/svelte';
	let props = $props();
	let project = props.project;
	let repo = project.Repository;
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
					<Label for="name">Repository</Label>
				</div>
				<div class="flex flex-col space-y-1.5">
					<Card.Root
						><Card.Content>
							<div class="flex items-center">
								<Avatar.Root class="me-4 ms-2" delayMs={1000}>
									<Avatar.Image src={repo.owner.avatar_url} alt={repo.full_name} />
									<Avatar.Fallback><Icon class="me-4 ms-2" icon="fa:user" /></Avatar.Fallback>
								</Avatar.Root>
								<span>{repo.name}</span>
								{#if repo.private}
									<Icon class="me-4 ms-2" icon="fa:lock" />
								{:else}
									<Icon class="me-4 ms-2" icon="fa:globe" />
								{/if}
							</div>
						</Card.Content></Card.Root
					>
				</div>
                <div class="flex flex-col space-y-1.5">
					<Label for="name">Repository</Label>
				</div>
                <div class="flex flex-col space-y-1.5">
                    <Label for="name">Project Name</Label>
                    <span>{project.Name}</span>
                    
                    <Label for="name">Description</Label>
                    <span>{project.Description}</span>
                    <Label for="name">Framework</Label>
                    <span>{project.Framework}</span>
                </div>
			</div>
		</form>
	</Card.Content>
	<Card.Footer class="flex justify-between">
		<Button variant="outline">Cancel</Button>
		<Button>Deploy</Button>
	</Card.Footer>
</Card.Root>
