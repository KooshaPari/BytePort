import type { Actions, PageServerLoad } from './$types.js';
import { superValidate, message } from 'sveltekit-superforms';

import { fail } from '@sveltejs/kit';
import { formSchema } from './schema.';
import { zod } from 'sveltekit-superforms/adapters';

export const load: PageServerLoad = async () => {
	console.log('Loading settings page');
	const form = await superValidate(zod(formSchema));
	console.log('Loaded settings page');
	return { form };
};

export const actions = {
	default: async ({ request }) => {
		const form = await superValidate(request, zod(formSchema));

		console.log('POST', form);

		if (!form.valid) return fail(400, { form });
		console.log('Updated Information');
		return message(form, 'Updated Information');
	}
} satisfies Actions;
