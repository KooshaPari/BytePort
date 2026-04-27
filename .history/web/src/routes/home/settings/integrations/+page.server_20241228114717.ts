import type { Actions } from './$types.js';

import { fail } from '@sveltejs/kit';
import { formSchema } from './schema';

import { zod } from 'sveltekit-superforms/adapters';

import { superValidate, message } from 'sveltekit-superforms';

export const actions = {
	default: async ({ request }) => {
		const form = await superValidate(request, zod(formSchema));

		console.log('POST', form);

		if (!form.valid) return fail(400, { form });

		return message(form, 'Updated Information');
	}
} satisfies Actions;
