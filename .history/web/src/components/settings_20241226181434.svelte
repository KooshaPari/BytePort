<script lang="ts" type="module">
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
	import { browser } from '$app/environment';
	import { superValidate, message } from 'sveltekit-superforms';
	import { formSchema } from '../routes/settings/schema';
	import { zod } from 'sveltekit-superforms/adapters';
	import type { z } from 'zod';

	type FormSchema = z.infer<typeof formSchema>;
	const modals = [
		{ label: 'ByteLLama', value: 'local' },
		{ label: 'OpenAI-4o', value: 'openai' }
	] as const;
	export let item: string;
	export let id: string;
	export let data: any; // Replace 'any' with your actual PageData type if available
	type Modal = (typeof modals)[number]['value'];

	import { onMount } from 'svelte';

	const mform: SuperForm<Record<string, unknown>> = superForm(data.form, {
		validators: zodClient(formSchema),

		dataType: 'json',
		onError: ({ result }) => {
			console.error('Form validation failed:', result);
		}
	});
	const { validate, enhance, form } = mform;
	console.log('initialized');
	console.log('Form: ', form);
</script>

<form method="POST" class="space-y-8" use:enhance>
	{#if item == 'integrations'}
		<div class="openAICard align-center flex flex-row gap-3">
			<Form.Field class=" " name="github" form={mform}>
				<Form.Control let:attrs>
					<Form.Label>Github Integration</Form.Label>
					<Input {...attrs} bind:value={$form.github} />
				</Form.Control>
				<Form.FieldErrors />
			</Form.Field>
			<Form.Field class=" " name="aws" form={$formData}>
				<Form.Control let:attrs>
					<Form.Label>AWS Access Key</Form.Label>
					<Input {...attrs} bind:value={$formData.aws.accessKey} />
					<Form.Label>AWS Secret Key</Form.Label>
					<Input {...attrs} bind:value={$form.aws.secretKey} />
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
				<Input {...attrs} bind:value={$form[demo]endpoint} />
				<Form.Label>Portfolio Key</Form.Label>
				<Input {...attrs} bind:value={$form.demo.apiKey} />
			</Form.Control>
			<Form.FieldErrors />
		</Form.Field>
	{/if}

	<Form.Button>Save {item}</Form.Button>
</form>

{#if browser}
	<SuperDebug data={$mform} />
{/if}
