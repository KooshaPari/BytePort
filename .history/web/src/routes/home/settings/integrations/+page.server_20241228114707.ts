import type { Actions } from './$types.js';
import { user, initializeUser } from '$lib/../stores/user';

import { fail } from '@sveltejs/kit';
import { formSchema } from './schema';
import type { PageServerLoad } from './$types.js';
import { zod } from 'sveltekit-superforms/adapters';
import type { User } from '$lib/../stores/user';
import { superValidate, message } from 'sveltekit-superforms';

type FormSchema = {
	github: string;
	aws: { accessKey: string; secretKey: string };
	openai: string;
	demo: { endpoint: string; apiKey: string };
	modal: 'openai' | 'local';
};
 
 
export const actions = {
	default: async ({ request }) => {
		const form = await superValidate(request, zod(formSchema));

		console.log('POST', form);

		if (!form.valid) return fail(400, { form });

		return message(form, 'Updated Information');
	}
} satisfies Actions;
