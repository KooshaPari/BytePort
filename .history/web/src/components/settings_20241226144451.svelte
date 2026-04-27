<script lang="ts" context="module">
	import { z } from 'zod';

	const modals = [
		{ label: 'ByteLLama', value: 'local' },
		{ label: 'OpenAI-4o', value: 'openai' }
	] as const;

	type Modal = (typeof modals)[number]['value'];

	export const formSchema = z.object({
		name: z
			.string({
				required_error: 'Required.'
			})
			.min(2, 'Name must be at least 2 characters.')
			.max(30, 'Name must not be longer than 30 characters'),
		// Hack: https://github.com/colinhacks/zod/issues/2280
		language: z.enum(modals.map((modal) => modal.value) as [Modal, ...Modal[]]),
		dob: z
			.string()
			.datetime({ message: '' })
			// we're setting it optional so the user can clear the date and we don't run into
			// type issues, but we refine it to make sure it's not undefined
			.optional()
			.refine((date) => date !== '', 'Please select a valid date.'),
		modal: z.enum(['inter', 'manrope', 'system'], {
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
	import {
		DateFormatter,
		type DateValue,
		getLocalTimeZone,
		parseDate
	} from '@internationalized/date';
	import * as Form from '$lib/registry/ui/form/index.js';
	import * as Popover from '$lib/registry/ui/popover/index.js';
	import * as Command from '$lib/registry/ui/command/index.js';

	import { Input } from '$lib/registry/ui/input/index.js';
	import { buttonVariants } from '$lib/registry/ui/button/index.js';
	import { cn } from '$lib/utils.js';
	import { browser } from '$app/environment';
	const { data, name } = $props<{
		data: SuperValidated<Infer<typeof formSchema>>;
		item: string;
	}>();

	const form = superForm(data, {
		validators: zodClient(formSchema)
	});
	const { form: formData, enhance, validate } = form;
</script>

<form method="POST" class="space-y-8" use:enhance>
	<Form.Field name="name" {form}>
		<Form.Control let:attrs>
			<Form.Label>Name</Form.Label>
			<Input {...attrs} bind:value={$formData.name} />
		</Form.Control>
		<Form.FieldErrors />
	</Form.Field>
	<Form.Field {form} name="aws" class="flex flex-col">
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
	</Form.Field>

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
					<Command.Input placeholder="Search modal..." />
					<Command.Empty>No modal found.</Command.Empty>
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
								{language.label}
							</Command.Item>
						{/each}
					</Command.List>
				</Command.Root>
			</Popover.Content>
		</Popover.Root>
	</Form.Field>

	<Form.Button>Save {name}</Form.Button>
</form>

{#if browser}
	<SuperDebug data={$formData} />
{/if}
