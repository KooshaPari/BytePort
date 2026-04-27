import type { Actions, PageServerLoad } from './$types.js';
import { superValidate, message } from 'sveltekit-superforms';
 
import { platform } from '@tauri-apps/plugin-os';
import { fail } from '@sveltejs/kit';
import { formSchema } from './schema';
import { zod } from 'sveltekit-superforms/adapters';
import type { User } from '$lib/../stores/user';
import   { user  } from '$lib/../stores/user';
export const load: PageServerLoad = async () => {
	 
	const baseUrl = await getBaseUrl();
    // read user from store
    const user
	const resp = await getCurrent(client, baseUrl);

	const form = await superValidate(resp, zod(formSchema));
	 
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

 
type UserCreds = {
	github: string;
	aws: { accessKey: string; secretKey: string };
	openai: string;
	demo: { endpoint: string; apiKey: string };
	modal: string;
};
async function getCurrent(user: User | null, baseUrl: string): Promise<UserCreds> {
	if (!user) {
		return {
			github: '',
			aws: { accessKey: '', secretKey: '' },
			openai: '',
			demo: { endpoint: '', apiKey: '' },
			modal: ''
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
		modal: ''
	};
}
