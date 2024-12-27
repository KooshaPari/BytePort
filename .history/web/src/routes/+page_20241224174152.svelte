<script lang="ts">
	import Icon from '@iconify/svelte';
	import { goto } from '$app/navigation';
	import { onMount, onDestroy } from 'svelte';
	import type { User } from '../stores/user';
	import { setUser, user, initializeUser } from '../stores/user';
	import { platform } from '@tauri-apps/plugin-os';
	const getBaseUrl = async () => {
		if (window.__TAURI_INTERNALS__) {
		 const currentPlatform: string  =  platform()
			console.log(currentPlatform);
			switch(currentPlatform) {
				case 'android':
					return 'http://10.0.2.2:8081';
				case 'windows':
					return 'http://localhost:8081';
				default:
					return 'http://localhost:8081';
			}}
			else{
				return 'http://localhost:8081';
			}

    };
	
	const getClientUrl =  () => {
		if (window.__TAURI_INTERNALS__) {
		 const currentPlatform: string | null =  platform();
		 if(currentPlatform == null) {
			return 'http://localhost:5173';
		 }
    
			switch(currentPlatform) {
				case 'android':
					return 'http:///10.0.2.2:5173';
				case 'windows':
					return 'http://localhost:5173';
				default:
					return 'http://localhost:5173';
			}}
			else{
				return 'http://localhost:5173';
			}

    };
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
				goto('/login');
			} else {
				console.log('Authenticated user:', value.data);
				// Perform actions for authenticated user
				client = value.data; // Assign the authenticated user to `client`
				populateLists(); // Populate user-specific lists or data
			}
		});

		// Ensure the user initialization is complete
		await initializeUser(baseUrl);

		onDestroy(() => {
			unsubscribe();
		});
	});
</script>

<div class="bg-dark-surface flex h-screen w-screen overflow-x-hidden" id="parent">
	<div id="header" class=" w-5/5 bg-dark-surfaceContainerLow h-1/5 flex-col justify-between ps-2.5">
		<div id="headerNav" class="h-3/5 w-screen pt-2.5">
			<img class="h-4/5 py-10" alt="BytePort" src="/src/assets/img/byte.png" />
			<div class="flex justify-end pe-2.5" id="navRight">
				<Icon
					class="hover:text-dark-primary active:text-dark-surfaceBright mx-1 h-6 w-6 cursor-pointer text-white"
					icon="ic:baseline-account-circle"
				/>
			</div>
		</div>
		<div id="headerContent" class="h-2/5 text-4xl text-white">Hello.</div>
	</div>
	<div id="body" class="w-4/5">
		<div id="mainBody"></div>
	</div>
	<div id="footer"></div>
</div>

<style>
</style>
