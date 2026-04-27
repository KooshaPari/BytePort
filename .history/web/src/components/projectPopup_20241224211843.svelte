<script lang="ts">
	import * as Dialog from '$lib/components/ui/dialog';
	import * as Button from '$lib/components/ui/button';
    import * as Accordion from "$lib/components/ui/accordion";
	import Icon from '@iconify/svelte';
	import ProjectForm from './projectForm.svelte';
	import type { Project, Instance } from '$lib/utils';
	import { onMount, onDestroy } from 'svelte';
	import ReviewCard from './reviewCard.svelte';
	import GitSearch from './gitSearch.svelte';
	import { user, initializeUser } from '../stores/user';
	import type { User } from '../stores/user';
	const SERVER_URL = 'http://localhost:8081';
	let projheadTxt: string = 'Welcome.';
	let stage = 1;
	let client: User | null = null;
	let projHeadDescr: string = "Let's Begin First Time Setup";

	import { platform } from '@tauri-apps/plugin-os';
	import { Description } from 'formsnap';
	const getBaseUrl = async () => {
		if (window.__TAURI_INTERNALS__) {
			const currentPlatform: string = platform();
			console.log(currentPlatform);
			switch (currentPlatform) {
				case 'android':
					return 'http://10.0.2.2:8081';
				case 'windows':
					return 'http://localhost:8081';
				default:
					return 'http://localhost:8081';
			}
		} else {
			return 'http://localhost:8081';
		}
	};
	let props = $props();
	let project = props.project;
	onMount(async () => {
		const baseUrl = await getBaseUrl();
		const unsubscribe = user.subscribe((value) => {
			// Handle pending state
			if (value.status === 'pending') {
				console.log('User state pending...');
				return; // Wait for initialization to complete
			}

			// Redirect if unauthenticated
			if (value.status !== 'authenticated') {
				console.log('User unauthenticated, redirecting...');
				// destroy dialog
				return;
			} else {
				client = value.data; // Assign the authenticated user to `client`
			}
		});

		// Ensure the user initialization is complete
		await initializeUser(baseUrl);

		onDestroy(() => {
			unsubscribe();
		});
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
            <div  id="cardLeft"
        </div>
		<Dialog.Footer></Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
