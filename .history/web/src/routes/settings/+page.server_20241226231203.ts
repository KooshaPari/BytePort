import type { Actions } from './$types.js';
import { setUser, user, initializeUser } from '$lib/../stores/user';
import type { SuperValidated } from 'sveltekit-superforms';
import { platform } from '@tauri-apps/plugin-os';
import { fail } from '@sveltejs/kit';
import { formSchema } from './schema';
 
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
	if ((window as any).__TAURI_INTERNALS__) {
		const currentPlatform: string = platform();
		console.log(currentPlatform);
		switch (currentPlatform) {
			case 'android':
				return 'http://10.0.2.2:8081';
			case 'windows':
				return 'http://localhost:8081';
			default:
				return 'http://localhost:8081';
		}
	} else {
		return 'http://localhost:8081';
	}
};

let client: User | null = null;
let baseUrl: string;
// convert to map name, '/name'

const unsubscribe = user.subscribe(async (value) => {
	if (value.status === 'pending') {
		console.log('User state pending...');
		return;
	}

	if (value.status !== 'authenticated') {
		console.log('User unauthenticated, redirecting...');
		goto('/login');
	} else {
		console.log('AUTH');
		console.log('Authenticated user:', value.data);
		client = value.data;

		// Update form with user data if available
		if (client && baseUrl) {
			const userCreds = await getCurrent(client, baseUrl);
			console.log('CR');
			console.log('CR: ', userCreds);
		}
	}
});

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
	try {
		let userCreds = FormSchema{
			github: '',
			aws: { accessKey: '', secretKey: '' },
			openai: '',
			demo: { endpoint: '', apiKey: '' },
			modal: 'local'
		};
		baseUrl = await getBaseUrl();
		console.log('Base URL:', baseUrl);
		await initializeUser(baseUrl);

		if (client) {
			 userCreds = await getCurrent(client, baseUrl);
			 
		}
	} catch (error) {
		console.error('Error in onMount:', error);
	}

	return { form: await superValidate(zod(formSchema)) };
};

export const actions = {
	default: async ({ request }) => {
		const form = await superValidate(request, zod(formSchema));

		console.log('POST', form);

		if (!form.valid) return fail(400, { form });

		return message(form, 'Updated Information');
	}
} satisfies Actions;
