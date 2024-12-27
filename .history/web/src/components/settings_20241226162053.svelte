<script lang="ts" context="module">
	import { z } from 'zod';

	const modals = [
		{ label: 'ByteLLama', value: 'local' },
		{ label: 'OpenAI-4o', value: 'openai' }
	] as const;

	type Modal = (typeof modals)[number]['value'];

	export const formSchema = z.object({
		github: z
			.string({
				required_error: 'Required.'
			})
			.min(2, 'Name must be at least 2 characters.')
			.max(30, 'Name must not be longer than 30 characters'),
		aws: z.object({
			accessKey: z
				.string({
					required_error: 'Required.'
				})
				.min(2, 'Name must be at least 2 characters.')
				.max(30, 'Name must not be longer than 30 characters'),
			secretKey: z
				.string({
					required_error: 'Required.'
				})
				.min(2, 'Name must be at least 2 characters.')
				.max(30, 'Name must not be longer than 30 characters')
		}),
		openai: z
			.string({
				required_error: 'Required.'
			})
			.min(2, 'Name must be at least 2 characters.')
			.max(30, 'Name must not be longer than 30 characters'),
		// Hack: https://github.com/colinhacks/zod/issues/2280
		demo: z.object({
			endpoint: z
				.string({
					required_error: 'Required.'
				})
				.min(2, 'Name must be at least 2 characters.')
				.max(30, 'Name must not be longer than 30 characters'),
			apiKey: z
				.string({
					required_error: 'Required.'
				})
				.min(2, 'Name must be at least 2 characters.')
				.max(30, 'Name must not be longer than 30 characters')
		}),

		modal: z.enum(modals.map((modal) => modal.value) as [Modal, ...Modal[]], {
			invalid_type_error: 'Select a modal',
			required_error: 'Please modal a font.'
		})
	});

	export type FormSchema = typeof formSchema;
</script>

<script lang="ts">
	import CaretSort from 'svelte-radix/CaretSort.svelte';
	import Check from 'svelte-radix/Check.svelte';
	import SuperDebug, { type Infer, type SuperValidated } from 'sveltekit-superforms';
	import { superForm } from 'sveltekit-superforms/client';
	import { zodClient } from 'sveltekit-superforms/adapters';
	import * as Form from '$lib/components/ui/form/index.js';
	import * as Popover from '$lib/components/ui/popover/index.js';
	import * as Command from '$lib/components/ui/command/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { buttonVariants } from '$lib/components/ui/button/index.js';
	import { cn } from '$lib/utils.js';
	import { browser } from '$app/environment';

	export let data;
	export let item;

	const { form, validate, enhance } = superForm(
		data ?? {
			github: '',
			aws: { accessKey: '', secretKey: '' },
			openai: '',
			demo: { endpoint: '', apiKey: '' },
			modal: 'local'
		},
		{
			validators: zodClient(formSchema),

			dataType: 'json',
			onError: ({ result }) => {
				console.error('Form validation failed:', result);
			}
		}
	);
	console.log('initialized');
	console.log('Form: ', form);
</script>

<form method="POST" class="space-y-8" use:enhance>
	{#if item == 'integrations' && $data}
		<div class="openAICard align-center flex flex-row gap-3">
			<Form.Field class=" " name="github" form={$data}>
				<Form.Control let:attrs>
					<Form.Label>Github Integration</Form.Label>
					<Input {...attrs} bind:value={$data.github} />
				</Form.Control>
				<Form.FieldErrors />
			</Form.Field>
			<Form.Field class=" " name="aws" form={$data}>
				<Form.Control let:attrs>
					<Form.Label>AWS Access Key</Form.Label>
					<Input {...attrs} bind:value={$data.aws.accessKey} />
					<Form.Label>AWS Secret Key</Form.Label>
					<Input {...attrs} bind:value={$data.aws.secretKey} />
				</Form.Control>
				<Form.FieldErrors />
			</Form.Field>
			<Form.Field form={$data} name="modal" class="flex flex-col justify-center">
				<Popover.Root>
					<Form.Control let:attrs>
						<Form.Label>Modals</Form.Label>
						<Popover.Trigger
							role="combobox"
							class={cn(
								buttonVariants({ variant: 'outline' }),
								'w-[200px] justify-between',
								!$data.modal && 'text-muted-foreground'
							)}
							{...attrs}
						>
							{modals.find((modal) => modal.value === $data.modal)?.label || 'Select a modal'}
							<CaretSort class="ml-2 size-4 shrink-0 opacity-50" />
						</Popover.Trigger>
						<input hidden value={$data.modal} name={attrs.name} />
					</Form.Control>
					<Popover.Content class="w-[200px] p-0">
						<Command.Root>
							<Command.List>
								{#each modals as modal}
									<Command.Item
										{...form}
										value={modal.label}
										onSelect={() => {
											$data.modal = modal.value;
											validate('modal');
										}}
									>
										<Check
											class={cn(
												'mr-2 size-4',
												modal.value === $data.modal ? 'opacity-100' : 'opacity-0'
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
			<Form.Field class=" " name="openai" form={$data}>
				<Form.Control let:attrs>
					<Form.Label>OpenAI Key</Form.Label>
					<Input
						class={cn(
							!$data.modal.includes('openai') && [
								'text-muted-foreground',
								'bg-muted',
								'border-dark-surfaceContainerLow'
							]
						)}
						disabled={!$data.modal.includes('openai')}
						{...attrs}
						bind:value={$data.openai}
					/>
				</Form.Control>
				<Form.FieldErrors />
			</Form.Field>
		</div>
		<Form.Field class=" " name="demo" form={$data}>
			<Form.Control let:attrs>
				<Form.Label>Portfolio URL</Form.Label>
				<Input {...attrs} bind:value={$data.demo.endpoint} />
				<Form.Label>Portfolio Key</Form.Label>
				<Input {...attrs} bind:value={$data.demo.apiKey} />
			</Form.Control>
			<Form.FieldErrors />
		</Form.Field>
	{/if}

	<Form.Button>Save {item}</Form.Button>
</form>

{#if browser}
	<SuperDebug data={$data} />
{/if}
