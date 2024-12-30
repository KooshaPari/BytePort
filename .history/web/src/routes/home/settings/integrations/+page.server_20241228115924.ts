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

const getBaseUrl = async () => {
	return 'http://localhost:8081';
};

let client: User | null = null;
let baseUrl: string;
// convert to map name, '/name'

type UserCreds = {
	github: string;
	aws: { accessKey: string; secretKey: string };
	openai: string;
	demo: { endpoint: string; apiKey: string };
	modal: 'openai' | 'local';
};
async function getCurrent(user: User | null, baseUrl: string): Promise<UserCreds> {
	if (!user) {
		return {
			github: '',
			aws: { accessKey: '', secretKey: '' },
			openai: '',
			demo: { endpoint: '', apiKey: '' },
			modal: 'local'
		};
	} else {
		// request baseurl/:id/creds GET with auth credentials only
		const response = await fetch(`${baseUrl}/user/${user.UUID}/creds`, {
			method: 'GET',
			credentials: 'include'
		});
		if (response.ok) {
			const resp = await response.json();

			return {
				github: resp.Git.Token,
				aws: { accessKey: resp.AwsCreds.AccessKeyID, secretKey: resp.AwsCreds.SecretAccessKey },
				openai: resp.OpenAICreds.APIKey,
				demo: { endpoint: resp.Portfolio.RootEndpoint, apiKey: resp.Portfolio.APIKey },
				modal: resp.OpenAICreds.APIKey === '' ? 'openai' : 'local'
			};
		}
	}
	return {
		github: '',
		aws: { accessKey: '', secretKey: '' },
		openai: '',
		demo: { endpoint: '', apiKey: '' },
		modal: 'local'
	};
}

export const load: PageServerLoad = async () => {
	let userCreds: FormSchema = {
		github: '',
		aws: { accessKey: '', secretKey: '' },
		openai: '',
		demo: { endpoint: '', apiKey: '' },
		modal: 'local'
	};
	try {
		baseUrl = await getBaseUrl();
		console.log('Base URL:', baseUrl);

		if (client) {
			console.log('CLIENT FOUND');
			userCreds = await getCurrent(client, baseUrl);
		} else {
			console.log('CLIENT NOT FOUND');
		}
	} catch (error) {
		console.error('Error in onMount:', error);
	}

	return { form: await superValidate(userCreds, zod(formSchema)) };
};

export const actions = {
	default: async ({ request }) => {
		const form = await superValidate(request, zod(formSchema));

		console.log('POST', form);

		if (!form.valid) return fail(400, { form });

		return message(form, 'Updated Information');
	}
} satisfies Actions;
