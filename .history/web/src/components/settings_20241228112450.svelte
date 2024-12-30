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
	import { browser } from '$app/environment';
	import { superValidate, message } from 'sveltekit-superforms';
	import { formSchema } from '../routes/settings/schema';
	import { zod } from 'sveltekit-superforms/adapters';
	import type { z } from 'zod';

	type FormSchema = {
		github: string;
		aws: { accessKey: string; secretKey: string };
		openai: string;
		demo: { endpoint: string; apiKey: string };
		modal: 'openai' | 'local';
	};
	const modals = [
		{ label: 'ByteLLama', value: 'local' },
		{ label: 'OpenAI-4o', value: 'openai' }
	] as const;
	export let item: string;
	export let id: string;

	export let data 
	type Modal = (typeof modals)[number]['value'];
 
	const mform: SuperForm<Record<string, unknown>> = superForm(data.form, {
		validators: zodClient(formSchema),

		dataType: 'json',
		onError: ({ result }) => {
			console.error('Form validation failed:', result);
		}
	});
 
	const { validate, enhance, form } = mform;
 
</script>


