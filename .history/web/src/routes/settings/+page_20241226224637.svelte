<script lang="ts">
	import Icon from '@iconify/svelte';
	import type { SuperValidated } from 'sveltekit-superforms';
	import { superValidate, message } from 'sveltekit-superforms';
	import { onMount } from 'svelte';
	import type { User } from '$lib/../stores/user';
	import * as Button from '$lib/components/ui/button';
	import { setUser, user, initializeUser } from '$lib/../stores/user';
	import { formSchema } from './schema';
	import { zod } from 'sveltekit-superforms/adapters';
	import { goto } from '$app/navigation';
	import { platform } from '@tauri-apps/plugin-os';
	import { writable } from 'svelte/store';
	import SetComp from '$lib/../components/settings.svelte';
	type FormSchema = {
		github: string;
		aws: { accessKey: string; secretKey: string };
		openai: string;
		demo: { endpoint: string; apiKey: string };
		modal: 'openai' | 'local';
	};

	 export let data;
	const menuItemsMap = new Map<string, string>([
		['Home', '/home '],
		['Profile', 'profile'],
		['Integrations', 'integrations'],
		['Settings', '/settings']
	]);
	// edit personal info, delete acc
	// edit, add or delete API INFO
	
</script>

<div class="bg-dark-surface flex h-screen w-screen overflow-x-hidden" id="mainDashPar">
	<div
		class="h-5/5 flex-ro bg-dark-surfaceContainer w-1/5 items-center justify-center"
		id="sideBar"
	>
		<button on:click={() => goto('/home')}>
			<img class="py-10" alt="BytePort" src="/src/assets/img/byte.png" />
		</button>``
		<div id="sideBarProfileCont"></div>
		<ul class="" id="menuList">
			{#each [...menuItemsMap] as [key, value]}
				<li class=" text-md w-5/5 py-2 text-center text-white">
					<button
						class="hover:bg-dark-surfaceContainerHigh active:bg-dark-surfaceContainer active:text-dark-surfaceBright w-4/5 py-2 text-center transition-all hover:-translate-y-1
						hover:rounded-full active:translate-y-0.5"
						on:click={() => {
							if (!value.includes('home')) {
								const mainBody = document.getElementById('bodyCont');
								if (mainBody) {
									mainBody.setAttribute('item', value);
								}
							}
							goto(value);
						}}
					>
						{key}
					</button>
				</li>
			{/each}
		</ul>
	</div>

	<div id="body" class="w-4/5">
		<div
			id="header"
			class=" w-5/5 bg-dark-surfaceContainerLow h-1/5 flex-col justify-between ps-2.5"
		>
			<div id="headerNav" class="h-3/5 pt-2.5">
				<div class="flex justify-end pe-2.5" id="navRight">
					<Icon
						class="hover:text-dark-primary active:text-dark-surfaceBright mx-1 h-6 w-6 cursor-pointer text-white"
						icon="ic:baseline-notifications"
					/>
					<Icon
						class="hover:text-dark-primary active:text-dark-surfaceBright mx-1 h-6 w-6 cursor-pointer text-white"
						on:click={() => goto('/settings')}
						icon="ic:baseline-account-circle"
					/>
				</div>
			</div>
			<div id="headerContent" class="h-2/5 text-4xl text-white">Settings</div>
		</div>
		<div id="mainBody">
			<SetComp id="bodyCont" item={'integrations'} {data} />
		</div>
		<div id="footer"></div>
	</div>
</div>

<style>
</style>
