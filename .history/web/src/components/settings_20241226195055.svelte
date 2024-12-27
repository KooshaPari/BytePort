<script lang="ts">
	import CaretSort from 'svelte-radix/CaretSort.svelte';
	import Check from 'svelte-radix/Check.svelte';
	import SuperDebug, {
		type Infer,
		type SuperValidated,
		type SuperForm
	} from 'sveltekit-superforms';
	import { superForm } from 'sveltekit-superforms';
	import { zodClient } from 'sveltekit-superforms/adapters';
	import * as Form from '$lib/components/ui/form/index.js';
	import * as Popover from '$lib/components/ui/popover/index.js';
	import * as Command from '$lib/components/ui/command/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { buttonVariants } from '$lib/components/ui/button/index.js';
	import { cn } from '$lib/utils.js';
	import { onMount } from 'svelte';
	import { browser } from '$app/environment';
	import { superValidate, message } from 'sveltekit-superforms';
	import { formSchema } from '../routes/settings/schema';
	import { zod } from 'sveltekit-superforms/adapters';
	import { platform } from '@tauri-apps/plugin-os';
	import type { z } from 'zod';
	import type { User } from '$lib/../stores/user';
	type FormSchema = z.infer<typeof formSchema>;
	const modals = [
		{ label: 'ByteLLama', value: 'local' },
		{ label: 'OpenAI-4o', value: 'openai' }
	] as const;
	export let item: string;
	export let id: string;
	export let user: User;

	export let data: any; // Replace 'any' with your actual PageData type if available
	type Modal = (typeof modals)[number]['value'];
	export const getBaseUrl = async () => {
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
	type UserCreds = {
		github: string;
		aws: { accessKey: string; secretKey: string };
		openai: string;
		demo: { endpoint: string; apiKey: string };
		modal: string;
	};
	let baseUrl
	const resp = getCurrent(user, baseUrl);
	data.form = superValidate(resp, zod(formSchema));

	async function getCurrent(user: User | null, baseUrl: string): Promise<UserCreds> {
		if (!user) {
			return {
				github: '',
				aws: { accessKey: '', secretKey: '' },
				openai: '',
				demo: { endpoint: '', apiKey: '' },
				modal: ''
			};
		} else {
			// request baseurl/:id/creds GET with auth credentials only
			const response = await fetch(`${baseUrl}/user/${user.uuid}/creds`, {
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

	const mform: SuperForm<Record<string, unknown>> = superForm(data.form, {
		validators: zodClient(formSchema),

		dataType: 'json',
		onError: ({ result }) => {
			console.error('Form validation failed:', result);
		}
	});
	const { validate, enhance, form } = mform;
</script>

<form method="POST" class="space-y-8" use:enhance>
	{#if item == 'integrations'}
		<div class="openAICard align-center flex flex-row gap-3">
			<Form.Field class=" " name="github" form={mform}>
				<Form.Control let:attrs>
					<Form.Label>Github Integration</Form.Label>
					<Input {...attrs} bind:value={$form.github as string} />
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
	{/if}

	<Form.Button>Save {item}</Form.Button>
</form>

{#if browser}
	<SuperDebug data={$form} />
{/if}
