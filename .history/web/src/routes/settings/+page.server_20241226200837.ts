import type { Actions, PageServerLoad } from './$types.js';
import { superValidate, message } from 'sveltekit-superforms';

import { platform } from '@tauri-apps/plugin-os';
import { fail } from '@sveltejs/kit';
import { formSchema } from './schema';
import { zod } from 'sveltekit-superforms/adapters';
import type { User } from '$lib/../stores/user';
import { user } from '$lib/../stores/user';
import { get } from 'svelte/store';
export const load: PageServerLoad = async () => {
	 
	// read user from store

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
let client: User | null = null;

const getBaseUrl = async (): Promise<string> => {
	if (typeof window !== 'undefined' && (window as any).__TAURI_INTERNALS__) {
		const currentPlatform = platform();
		switch (currentPlatform) {
			case 'android':
				return 'http://10.0.2.2:8081';
			case 'windows':
				return 'http://localhost:8081';
			default:
				return 'http://localhost:8081';
		}
	}
	return 'http://localhost:8081';
};
