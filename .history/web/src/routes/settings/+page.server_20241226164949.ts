import type { Actions, PageServerLoad } from './$types.js';
import { superValidate, message } from 'sveltekit-superforms';
 
import { fail } from '@sveltejs/kit';
import { formSchema } from '../../components/schema.js';
import { zod } from 'sveltekit-superforms/adapters';

export const load: PageServerLoad = async () => {
	const form = await superValidate(zod(formSchema));
	return { form };
};

export const actions = {
	default: async ({ request }) => {
		const form = await superValidate(request, zod(formSchema));

		console.log('POST', form);

		if (!form.valid) return fail(400, { form });

		return message(form, 'Updated Information');
	}
} satisfies Actions;
