<script lang="ts">
	import Icon from '@iconify/svelte';
	import type { SuperValidated } from 'sveltekit-superforms';
	import { superValidate, message } from 'sveltekit-superforms';
	import { onMount } from 'svelte';
	import type { User, UserLink } from '$lib/../stores/user';
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
		name: string;
		email: string;
		password: {
			value: string;
			confirm: string;
		}
	};
	let initialized = false;
	 
	const DEFAULT_CREDS: User = {
		uuid: '',
		name: '',
		email: ''
	};
	const mform = superForm(data.form, {
		validators: zodClient(formSchema),
		dataType: 'json',
		onError: ({ result }) => {
			console.error('Form validation failed:', result);
		}
	});
	const { form, errors, enhance } = mform;

 
	let client: User | null = null;
 
	const menuItemsMap = new Map<string, string>([
		['Home', '/home '],
		['Profile', '/home/settings/profile'],
		['Integrations', '/home/settings/integrations'],
		['Settings', '/home/settings/profile']
	]);
	// edit personal info, delete acc
	// edit, add or delete API INFO
	const getBaseUrl = async () => {
		return 'http://localhost:8081';
	};
	 

	async function GitLink() {
		try {
			const baseUrl = await getBaseUrl();
			// First, send user data

			const popup = window.open(`${baseUrl}/link`, '_blank', 'width=600,height=600');

			return true;

			if (!popup) {
				console.error('Failed to open popup window');
				return false;
			}
		} catch (error) {
			console.error('Error during link process:', error);
			return false;
		}
		// check that link response on get and post is 200
	}
	async function subLink() {
		try {
			const baseUrl = await getBaseUrl();
			const userData: UserLink = {
				UUID: client?.uuid || '',
				Name: client?.name || '',
				Email: client?.email || '',
			};
			console.log('User: ', userData);
			const response = await fetch(`${baseUrl}/link`, {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json'
				},
				credentials: 'include',
				body: JSON.stringify(userData)
			});

			if (!response.ok) {
				throw new Error('Failed to initialize link process');
			}
		} catch (error) {
			console.error('Error during link process:', error);
			return false;
		}
	}
	async function getCurrent(user: User, baseUrl: string): Promise<UserCreds> {
		if (!user?.uuid) {
			console.log('No user found');
			return DEFAULT_CREDS;
		}

		try {
			const response = await fetch(`${baseUrl}/user/${user.uuid}/creds`, {
				method: 'GET',
				credentials: 'include'
			});

			if (!response.ok) {
				throw new Error('Failed to fetch credentials');
			}

			const resp = await response.json();
			return {
				github: resp.git.Token === '' ? 'Not Linked' : 'Authenticated',
				aws: {
					accessKey: resp.awsCreds.AccessKeyID,
					secretKey: resp.awsCreds.SecretAccessKey
				},
				llmConfig: {
					provider: resp.llmConfig.provider,
					providers: resp.llmConfig.providers
				},
				demo: {
					endpoint: resp.portfolio.RootEndpoint,
					apiKey: resp.portfolio.APIKey
				}
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
			userCreds = await getCurrent(client, baseUrl);
			console.log('UCP: ', userCreds);

			// Update form with fetched credentials
			form.set({
				github: userCreds.github,
				aws: userCreds.aws,
				llm: userCreds.llmConfig,
				demo: userCreds.demo
			});

			console.log('Form:', $form.llm);
		}
	});

	onMount(() => {
		const baseUrl = 'http://localhost:8081';
		initializeUser(baseUrl);
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
							<Form.Label class="flex gap-1"><Icon icon="mdi:github"></Icon>Github</Form.Label>
							<div class="  flex flex-row gap-2">
								{#if $form.github === 'Not Linked'}
									<Button.Root
										class="   "
										on:click={() => {
											GitLink();
											$form.github = 'Authenticated';
										}}><Icon icon="mdi:github"></Icon>Link</Button.Root
									>
								{:else}
									<Button.Root class="w-[100px]  "><Icon icon="mdi:check"></Icon>Linked</Button.Root
									>
									<Button.Root
										class="bg-dark-secondaryContainer  w-[50px] "
										on:click={() => {
											$form.github = 'Not Linked';
											//validate('github');
										}}
										><Icon icon="mdi:close-circle" class="text-destructive-foreground"></Icon>
									</Button.Root>
								{/if}
							</div>
						</Form.Control>
						<Form.FieldErrors />
					</Form.Field>
					<Form.Field class=" " name="aws" form={mform}>
						<Form.Control let:attrs>
							<Form.Label>AWS Access Key</Form.Label>
							<Input {...attrs} bind:value={$form.aws.accessKey as string} />
							<Form.Label>AWS Secret Key</Form.Label>
							<Input {...attrs} type="password" bind:value={$form.aws.secretKey as string} />
						</Form.Control>
						<Form.FieldErrors />
					</Form.Field>
					<Form.Field form={mform} name="llm" class="flex flex-col justify-center">
						<Popover.Root>
							<Form.Control let:attrs>
								<Form.Label>Provider</Form.Label>
								<Popover.Trigger
									role="combobox"
									class={cn(
										buttonVariants({ variant: 'outline' }),
										'w-[200px] justify-between',
										!$form.llm.provider && 'text-muted-foreground'
									)}
									{...attrs}
								>
									{providers.find((provider) => provider.value === $form.llm.provider)?.label ||
										'Select a provider'}
									<CaretSort class="ml-2 size-4 shrink-0 opacity-50" />
								</Popover.Trigger>
							</Form.Control>
							<Popover.Content class="w-[200px] p-0">
								<Command.Root>
									<Command.List>
										{#each providers as provider}
											<Command.Item
												value={provider.label}
												onSelect={() => {
													$form.llm.provider = provider.value;
													// Initialize provider if it doesn't exist
													if (!$form.llm.providers[provider.value]) {
														$form.llm.providers[provider.value] = {
															modal: '',
															api_key: ''
														};
													}
												}}
											>
												<Check
													class={cn(
														'mr-2 size-4',
														provider.value === $form.llm.provider ? 'opacity-100' : 'opacity-0'
													)}
												/>
												{provider.label}
											</Command.Item>
										{/each}
									</Command.List>
								</Command.Root>
							</Popover.Content>
						</Popover.Root>

						{#if $form.llm.provider != ''}
							<Popover.Root>
								<Form.Control let:attrs>
									<Form.Label>Model</Form.Label>
									<Popover.Trigger
										role="combobox"
										class={cn(buttonVariants({ variant: 'outline' }), 'w-[200px] justify-between')}
										{...attrs}
									>
										{#if $form.llm.provider && $form.llm.providers[$form.llm.provider]}
											{modals.find(
												(modal) =>
													modal.value === $form.llm.providers[$form.llm.provider].modal &&
													modal.provider === $form.llm.provider
											)?.label || 'Select a Model'}
										{:else}
											Select a Model
										{/if}
										<CaretSort class="ml-2 size-4 shrink-0 opacity-50" />
									</Popover.Trigger>
								</Form.Control>
								<Popover.Content class="w-[200px] p-0">
									<Command.Root>
										<Command.List>
											{#each modals.filter((modal) => modal.provider === $form.llm.provider) as modal}
												<Command.Item
													value={modal.label}
													onSelect={() => {
														if ($form.llm.providers[$form.llm.provider]) {
															$form.llm.providers[$form.llm.provider].modal = modal.value;
														}
													}}
												>
													<Check
														class={cn(
															'mr-2 size-4',
															$form.llm.providers[$form.llm.provider]?.modal === modal.value
																? 'opacity-100'
																: 'opacity-0'
														)}
													/>
													{modal.label}
												</Command.Item>
											{/each}
										</Command.List>
									</Command.Root>
								</Popover.Content>
							</Popover.Root>
						{/if}

						{#if $form.llm.provider && $form.llm.provider !== 'local'}
							<Form.Control let:attrs>
								<Form.Label>{$form.llm.provider} API Key</Form.Label>
								<Input
									type="password"
									{...attrs}
									bind:value={$form.llm.providers[$form.llm.provider].api_key}
								/>
							</Form.Control>
						{/if}
					</Form.Field>

					<Form.Field class=" " name="demo" form={mform}>
						<Form.Control let:attrs>
							<Form.Label>Portfolio URL</Form.Label>
							<Input {...attrs} bind:value={$form.demo.endpoint as string} />
							<Form.Label>Portfolio Key</Form.Label>
							<Input {...attrs} type="password" bind:value={$form.demo.apiKey as string} />
						</Form.Control>
						<Form.FieldErrors />
					</Form.Field>
				</div>
				<Form.Button on:click={() => subLink()}>Save Integrations</Form.Button>
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
