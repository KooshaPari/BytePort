<script lang="ts">
	import * as Dialog from '$lib/components/ui/dialog';
	import * as Button from '$lib/components/ui/button';
	import { goto } from '$app/navigation';
	import { populateLists, getBaseUrl } from '$lib/utils';

	import * as Accordion from '$lib/components/ui/accordion';
	import Icon from '@iconify/svelte';
	import ProjectForm from './projectForm.svelte';
	import type { Project, Instance } from '$lib/utils';
	import { onMount, onDestroy } from 'svelte';
	import ReviewCard from './reviewCard.svelte';
	import GitSearch from './gitSearch.svelte';
	import { platform } from '@tauri-apps/plugin-os';
	import { Description } from 'formsnap';
	import { user, initializeUser } from '../stores/user';
	import type { User } from '../stores/user';
	import { json } from '@sveltejs/kit';
	const SERVER_URL = 'http://localhost:8081';
	let projheadTxt: string = 'Welcome.';
	let stage = 1;
	let client: User | null = null;
	let projHeadDescr: string = "Let's Begin First Time Setup";
	const unsubscribe = user.subscribe((value) => {
		// Handle pending state
		if (value.status === 'pending') {
			console.log('User state pending...');
			return; // Wait for initialization to complete
		}

		// Redirect if unauthenticated
		if (value.status !== 'authenticated') {
			console.log('User unauthenticated, redirecting...');
			goto('/login');
		} else {
			console.log('Authenticated user:', value.data);
			// Perform actions for authenticated user
			client = value.data; // Assign the authenticated user to `client`
			populateLists(); // Populate user-specific lists or data
		}
	});

	let props = $props();
	let project: Project = props.project;
	onMount(async () => {
		let project: Project = props.project;
		const baseUrl = await getBaseUrl();
		try {
			project.Deployments = await Promise.resolve(JSON.parse(project.DeploymentsJSON as string));
		} catch (error) {
			console.error('Error parsing deployments:', error);
			project.Deployments = null;
		}
		console.log('PROJECTS: ', project);
		console.log('PROJECT DEPLOYMENTS: ', project.Deployments);

		// Ensure the user initialization is complete
		await initializeUser(baseUrl);
	});
</script>

<Dialog.Root>
	<Dialog.Trigger
		><Icon
			class="active:text-dark-surfaceVariant hover:bg-dark-onPrimaryContainer active:bg-dark-onPrimary bg-dark-primary text-dark-onPrimary h-max w-max rounded-full p-5 text-4xl transition-all hover:-translate-y-2 hover:scale-105 active:translate-y-1 active:scale-100"
			icon="ic:baseline-add"
		/></Dialog.Trigger
	>

	<Dialog.Content class="flex flex-col items-center justify-center space-y-4"
		><Dialog.Header>
			<Dialog.Title>{project.Name}</Dialog.Title>

			<Dialog.Description><h1>{project.Description}</h1></Dialog.Description>
		</Dialog.Header>
		<!--Access Window, Details view, resource view, status -->
		<div id="vmView">
			<div id="cardLeft">
				<!--Access Window-->
			</div>
			<div id="cardRight">
				<!--Access URL, Status, Info, ProjectURL, Project Info, Resources Accordian-->
				<h1>Name: {project.Name}</h1>
				<h2>Description: {project.Description}</h2>

				<Button.Root on:click={() => (window.location.href = project.AccessURL)}>Access</Button.Root
				>
				<div id="resources">
					<Accordion.Root>
						<Accordion.Item value="item-1">
							<Accordion.Trigger>Resources</Accordion.Trigger>
							<Accordion.Content>
								{#if project.Deployments}
									<script>
										console.log('PROJECT DEPLOYMENTS: ');
									</script>
									{#each project.Deployments as instance}
										<script>
											console.log('INSTANCE: ');
										</script>
										{#each instance.Resources[0] as resource}
											<script>
												console.log('RESOURCE: ');
											</script>
											<div id="resource" class="bg-dark-surfaceContainerHigh mb-2 rounded-lg p-4">
												<h1 class="text-dark-onSurface text-xl font-bold">{resource.Name}</h1>
												<h2 class="text-dark-onSurfaceVariant text-lg">Type: {resource.Type}</h2>
												<h3 class="text-md text-dark-secondary">ARN: {resource.ARN}</h3>
												<h3
													class="text-md
												{resource.Status === 'Running'
														? 'text-green-500'
														: resource.Status === 'Stopped'
															? 'text-red-500'
															: 'text-yellow-500'}"
												>
													Status: {resource.Status}
												</h3>
												<h1>TEST</h1>
											</div>
										{/each}
									{/each}
								{:else}
									<p class="text-dark-onSurfaceVariant">No deployments found</p>
								{/if}
							</Accordion.Content>
						</Accordion.Item>
					</Accordion.Root>
				</div>
			</div>
		</div>
		<Dialog.Footer></Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
