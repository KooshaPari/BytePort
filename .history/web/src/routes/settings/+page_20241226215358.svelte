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

	// Initialize form with default values
	const defaultFormData: FormSchema = {
		github: '',
		aws: { accessKey: '', secretKey: '' },
		openai: '',
		demo: { endpoint: '', apiKey: '' },
		modal: 'local'
	};

	// Create a writable store for the form
	const nform = writable<SuperValidated<FormSchema>>();
	let initialized = false;

	async function initializeForm() {
		if (initialized) return;

		try {
			const validatedForm = await superValidate(defaultFormData, zod(formSchema));
			nform.set(validatedForm);
			initialized = true;
		} catch (error) {
			console.error('Failed to initialize form:', error);
		}
	}

	// Initialize immediately
	initializeForm();
	const menuItemsMap = new Map<string, string>([
		['Home', '/home '],
		['Profile', 'profile'],
		['Integrations', 'integrations'],
		['Settings', '/settings']
	]);
	// edit personal info, delete acc
	// edit, add or delete API INFO
	const getBaseUrl = async () => {
		if ((window as any).__TAURI_INTERNALS__) {
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

	let client: User | null = null;
	let baseUrl: string;
	// convert to map name, '/name'

	const unsubscribe = user.subscribe(async (value) => {
		if (value.status === 'pending') {
			console.log('User state pending...');
			return;
		}

		if (value.status !== 'authenticated') {
			console.log('User unauthenticated, redirecting...');
			goto('/login');
		} else {
			console.log('AUTH');
			console.log('Authenticated user:', value.data);
			client = value.data;

			// Update form with user data if available
			if (client && baseUrl) {
				const userCreds = await getCurrent(client, baseUrl);
				console.log('CR');
				console.log('CR: ', userCreds);
				const validatedForm = await superValidate(userCreds, zod(formSchema));
				nform.set(validatedForm);
			}
		}
	});

	type UserCreds = {
		github: string;
		aws: { accessKey: string; secretKey: string };
		openai: string;
		demo: { endpoint: string; apiKey: string };
		modal: 'openai' | 'local';
	};
	async function getCurrent(user: User | null, baseUrl: string): Promise<UserCreds> {
		if (!user) {
			return {
				github: '',
				aws: { accessKey: '', secretKey: '' },
				openai: '',
				demo: { endpoint: '', apiKey: '' },
				modal: 'local'
			};
		} else {
			// request baseurl/:id/creds GET with auth credentials only
			const response = await fetch(`${baseUrl}/user/${user.UUID}/creds`, {
				method: 'GET',
				credentials: 'include'
			});
			if (response.ok) {
				const resp = await response.json();

				return {
					github: resp.Git.Token,
					aws: { accessKey: resp.AwsCreds.AccessKeyID, secretKey: resp.AwsCreds.SecretAccessKey },
					openai: resp.OpenAICreds.APIKey,
					demo: { endpoint: resp.Portfolio.RootEndpoint, apiKey: resp.Portfolio.APIKey },
					modal: resp.OpenAICreds.APIKey === '' ? 'openai' : 'local'
				};
			}
		}
		return {
			github: '',
			aws: { accessKey: '', secretKey: '' },
			openai: '',
			demo: { endpoint: '', apiKey: '' },
			modal: ''
		};
	}

	let resp: UserCreds;
	onMount(() => {
		const initialize = async () => {
			try {
				baseUrl = await getBaseUrl();
				console.log('Base URL:', baseUrl);
				await initializeUser(baseUrl);

				if (client) {
					const userCreds = await getCurrent(client, baseUrl);
					const validatedForm = await superValidate(userCreds, zod(formSchema));
					console.log('C:', userCreds);
					nform.set(validatedForm);
				}
			} catch (error) {
				console.error('Error in onMount:', error);
			}
		};

		initialize();
		return () => unsubscribe();
	});
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
			<SetComp id="bodyCont" item={'integrations'} {nform} />
		</div>
		<div id="footer"></div>
	</div>
</div>

<style>
</style>
