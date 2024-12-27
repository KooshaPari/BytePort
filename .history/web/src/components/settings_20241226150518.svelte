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
		aws: z
			.string({
				required_error: 'Required.'
			})
			.min(2, 'Name must be at least 2 characters.')
			.max(30, 'Name must not be longer than 30 characters'),
		openai: z
			.string({
				required_error: 'Required.'
			})
			.min(2, 'Name must be at least 2 characters.')
			.max(30, 'Name must not be longer than 30 characters'),
		// Hack: https://github.com/colinhacks/zod/issues/2280
		demo: z
			.string({
				required_error: 'Required.'
			})
			.min(2, 'Name must be at least 2 characters.')
			.max(30, 'Name must not be longer than 30 characters'),

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
	import SuperDebug, { type Infer, type SuperValidated, superForm } from 'sveltekit-superforms';
	import { zodClient } from 'sveltekit-superforms/adapters';

	import * as Form from '$lib/components/ui/form/index.js';
	import * as Popover from '$lib/components/ui/popover/index.js';
	import * as Command from '$lib/components/ui/command/index.js';

	import { Input } from '$lib/components/ui/input/index.js';
	import { buttonVariants } from '$lib/components/ui/button/index.js';
	import { cn } from '$lib/utils.js';
	import { browser } from '$app/environment';
	export let item: string = '';
	export let id: string = '';
	export let data: SuperValidated<Infer<typeof formSchema>>;

	const form = superForm(data, {
		validators: zodClient(formSchema)
	});
	const { form: formData, enhance, validate } = form;
</script>

<form method="POST" class="space-y-8" use:enhance>
	<!--<Form.Field name="name" {form}>
		<Form.Control let:attrs>
			<Form.Label>Name</Form.Label>
			<Input {...attrs} bind:value={$formData.name} />
		</Form.Control>
		<Form.FieldErrors />
	</Form.Field>
	<Form.Field {form} name="name" class="flex flex-col">
		<Form.Control let:attrs>
			<Form.Label>AWS</Form.Label>
			<Popover.Root>
				<Popover.Trigger
					class={cn(
						buttonVariants({ variant: 'outline' }),
						'w-[240px] justify-start text-left font-normal'
						/*!dobValue && 'text-muted-foreground'*/
					)}
					{...attrs}
				></Popover.Trigger>
				<Popover.Content class="w-auto p-0" align="start"></Popover.Content>
				<input hidden bind:value={$formData.dob} name={attrs.name} />
			</Popover.Root>
		</Form.Control>
		<Form.FieldErrors />
	</Form.Field>-->
	{#if item == 'integrations'}
    <div class="openAICard">
		<Form.Field {form} name="modal" class="flex flex-col">
			<Popover.Root>
				<Form.Control let:attrs>
					<Form.Label>Modals</Form.Label>
					<Popover.Trigger
						role="combobox"
						class={cn(
							buttonVariants({ variant: 'outline' }),
							'w-[200px] justify-between',
							!$formData.modal && 'text-muted-foreground'
						)}
						{...attrs}
					>
						{modals.find((modal) => modal.value === $formData.modal)?.label || 'Select a modal'}
						<CaretSort class="ml-2 size-4 shrink-0 opacity-50" />
					</Popover.Trigger>
					<input hidden value={$formData.modal} name={attrs.name} />
				</Form.Control>
				<Popover.Content class="w-[200px] p-0">
					<Command.Root>
						<Command.List>
							{#each modals as modal}
								<Command.Item
									{...form}
									value={modal.label}
									onSelect={() => {
										$formData.modal = modal.value;
										validate('modal');
									}}
								>
									<Check
										class={cn(
											'mr-2 size-4',
											modal.value === $formData.modal ? 'opacity-100' : 'opacity-0'
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
        
    </div>
	{/if}

	<Form.Button>Save {item}</Form.Button>
</form>

{#if browser}
	<SuperDebug data={$formData} />
{/if}
