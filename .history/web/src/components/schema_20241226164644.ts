import { z } from 'zod';

const modals = [
	{ label: 'ByteLLama', value: 'local' },
	{ label: 'OpenAI-4o', value: 'openai' }
] as const;

type Modal = (typeof modals)[number]['value'];

export const formSchema =object({
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
