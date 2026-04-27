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
	let project = props.project;
	onMount(async () => {
		const baseUrl = await getBaseUrl();
		project.Deployments = json.
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
				<h3>Framework: {project.Framework}</h3>
				<Button.Root on:click={() => (window.location.href = project.AccessURL)}>Access</Button.Root
				>
				<div id="resources">
					<Accordion.Root>
						<Accordion.Item value="item-1">
							<Accordion.Trigger>Resources</Accordion.Trigger>
							<Accordion.Content>
								{#each project.Deployments. as resource}
									<h1>test</h1>
									<div id="resource">
										<h1>{resource.Name}</h1>
										<h2>{resource.Type}</h2>
										<h3>{resource.ARN}</h3>
										<h3>{resource.Status}</h3>
									</div>
								{/each}
							</Accordion.Content>
						</Accordion.Item>
					</Accordion.Root>
				</div>
			</div>
		</div>
		<Dialog.Footer></Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
