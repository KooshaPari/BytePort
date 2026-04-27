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
	import CaretSort from 'svelte-radix/CaretSort.svelte';
	import Check from 'svelte-radix/Check.svelte';
	import SuperDebug, { type Infer, type SuperForm } from 'sveltekit-superforms';
	import { superForm } from 'sveltekit-superforms';
	import { zodClient } from 'sveltekit-superforms/adapters';
	import * as Form from '$lib/components/ui/form/index.js';
	import * as Popover from '$lib/components/ui/popover/index.js';
	import * as Command from '$lib/components/ui/command/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { buttonVariants } from '$lib/components/ui/button/index.js';
	import { cn } from '$lib/utils.js';
	import { browser } from '$app/environment';

	import { string, type z } from 'zod';

	import { writable } from 'svelte/store';
	import SetComp from '$lib/../components/settings.svelte';
	export let data;
	type FormSchema = {
		github: string;
		aws: { accessKey: string; secretKey: string };
		openai: string;
		demo: { endpoint: string; apiKey: string };
		modal: 'openai' | 'local';
	};
	let initialized = false;
	const modals = [
		{ label: 'ByteLLama', value: 'local' },
		{ label: 'OpenAI-4o', value: 'openai' }
	] as const;
	const DEFAULT_CREDS: UserCreds = {
		github: '',
		aws: { accessKey: '', secretKey: '' },
		llm: { provider: '', providers: new Map<string, AIProvider>() },
		demo: { endpoint: '', apiKey: '' },
		modal: 'local'
	};
	const { form, errors, enhance } = superForm(data.form, {
		validators: zodClient(formSchema),
		dataType: 'json',
		onError: ({ result }) => {
			console.error('Form validation failed:', result);
		}
	});

	// State management
	let userCreds: UserCreds = DEFAULT_CREDS;
	let client: User | null = null;
	type Modal = (typeof modals)[number]['value'];
	const menuItemsMap = new Map<string, string>([
		['Home', '/home '],
		['Profile', '/home/settings/profile'],
		['Integrations', '/home/settings/integrations'],
		['Settings', '/home/settings']
	]);
	// edit personal info, delete acc
	// edit, add or delete API INFO
	const getBaseUrl = async () => {
		return 'http://localhost:8081';
	};
	type AIProvider = {
		provider: string;
		apiKey: string;
	};
	type UserCreds = {
		github: string;
		aws: { accessKey: string; secretKey: string };
		llm: { provider: string; providers: Map<string, AIProvider> };
		demo: { endpoint: string; apiKey: string };
		modal: 'openai' | 'local';
	};

	async function getCurrent(user: User, baseUrl: string): Promise<UserCreds> {
		if (!user?.UUID) {
			console.log('No user found');
			return DEFAULT_CREDS;
		}

		try {
			const response = await fetch(`${baseUrl}/user/${user.UUID}/creds`, {
				method: 'GET',
				credentials: 'include'
			});

			if (!response.ok) {
				throw new Error('Failed to fetch credentials');
			}

			const resp = await response.json();
			return {
				github: resp.Git.Token,
				aws: {
					accessKey: resp.AwsCreds.AccessKeyID,
					secretKey: resp.AwsCreds.SecretAccessKey
				},
				llm: {
					provider: resp.LLM.provider,
					providers: resp.LLM.Providers
				},
				demo: {
					endpoint: resp.Portfolio.RootEndpoint,
					apiKey: resp.Portfolio.APIKey
				},
				modal: resp.OpenAICreds.APIKey === '' ? 'openai' : 'local'
			};
		} catch (error) {
			console.error('Error fetching credentials:', error);
			return DEFAULT_CREDS;
		}
	}

	// Handle user authentication and initialization
	const unsubscribe = user.subscribe(async (value) => {
		if (value.status === 'pending') {
			console.log('User state pending...');
			return;
		}

		if (value.status !== 'authenticated') {
			if (browser) {
				console.log('User unauthenticated, redirecting...');
				goto('/login');
			}
			return;
		}

		client = value.data;
		console.log('Authenticated user:', client);

		if (client) {
			const baseUrl = await getBaseUrl();
			userCreds = await getCurrent(client,baseUrl);
			// Update form with fetched credentials
			form.set({
				github: userCreds.github,
				aws: userCreds.aws,
				openai: userCreds.llm.providers.get('openai')?.apiKey || '',
				demo: userCreds.demo,
				modal: userCreds.modal
			});
		}
	});

	onMount(() => {
		initializeUser(BASE_URL);
		return unsubscribe;
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
						on:click={() => goto('/home/settings')}
						icon="ic:baseline-account-circle"
					/>
				</div>
			</div>
			<div id="headerContent" class="h-2/5 text-4xl text-white">Settings</div>
		</div>
		<div id="mainBody">
			<form method="POST" class="space-y-8" use:enhance>
				<div class="openAICard align-center flex flex-row gap-3">
					<Form.Field class=" " name="github" form={mform}>
						<Form.Control let:attrs>
							<Form.Label>Github Integration</Form.Label>
						</Form.Control>
						<Form.FieldErrors />
					</Form.Field>
					<Form.Field class=" " name="aws" form={mform}>
						<Form.Control let:attrs>
							<Form.Label>AWS Access Key</Form.Label>
							<Input {...attrs} bind:value={$form.aws.accessKey as string} />
							<Form.Label>AWS Secret Key</Form.Label>
							<Input {...attrs} bind:value={$form.aws.secretKey as string} />
						</Form.Control>
						<Form.FieldErrors />
					</Form.Field>
					<Form.Field form={mform} name="modal" class="flex flex-col justify-center">
						<Popover.Root>
							<Form.Control let:attrs>
								<Form.Label>Modals</Form.Label>
								<Popover.Trigger
									role="combobox"
									class={cn(
										buttonVariants({ variant: 'outline' }),
										'w-[200px] justify-between',
										!$form.modal && 'text-muted-foreground'
									)}
									{...attrs}
								>
									{modals.find((modal) => modal.value === $form.modal)?.label || 'Select a modal'}
									<CaretSort class="ml-2 size-4 shrink-0 opacity-50" />
								</Popover.Trigger>
								<input hidden value={$form.modal} name={attrs.name} />
							</Form.Control>
							<Popover.Content class="w-[200px] p-0">
								<Command.Root>
									<Command.List>
										{#each modals as modal}
											<Command.Item
												{...form}
												value={modal.label}
												onSelect={() => {
													$form.modal = modal.value;
													validate('modal');
												}}
											>
												<Check
													class={cn(
														'mr-2 size-4',
														modal.value === $form.modal ? 'opacity-100' : 'opacity-0'
													)}
												/>
												{modal.label}
											</Command.Item>
										{/each}
									</Command.List>
								</Command.Root>
							</Popover.Content>
						</Popover.Root>
					</Form.Field>
					<Form.Field class=" " name="openai" form={mform}>
						<Form.Control let:attrs>
							<Form.Label>OpenAI Key</Form.Label>
							<Input
								class={cn(
									!$form.modal.includes('openai') && [
										'text-muted-foreground',
										'bg-muted',
										'border-dark-surfaceContainerLow'
									]
								)}
								disabled={!$form.modal.includes('openai')}
								{...attrs}
								bind:value={$form.openai}
							/>
						</Form.Control>
						<Form.FieldErrors />
					</Form.Field>
				</div>
				<Form.Field class=" " name="demo" form={mform}>
					<Form.Control let:attrs>
						<Form.Label>Portfolio URL</Form.Label>
						<Input {...attrs} bind:value={$form.demo.endpoint as string} />
						<Form.Label>Portfolio Key</Form.Label>
						<Input {...attrs} bind:value={$form.demo.apiKey as string} />
					</Form.Control>
					<Form.FieldErrors />
				</Form.Field>

				<Form.Button>Save Integrations</Form.Button>
			</form>

			{#if browser}
				<SuperDebug data={$form} />
			{/if}
		</div>
		<div id="footer"></div>
	</div>
</div>

<style>
</style>
