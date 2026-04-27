import type { Actions, PageServerLoad } from './$types.js';
import { superValidate, message } from 'sveltekit-superforms';
import { goto } from '$app/navigation';
import { platform } from '@tauri-apps/plugin-os';
import { fail } from '@sveltejs/kit';
import { formSchema, type FormSchema } from './schema';
import { zod } from 'sveltekit-superforms/adapters';

export const load: PageServerLoad = async () => {
	console.log('Loading settings page');
	const baseUrl = await getBaseUrl();
	await initializeUser(baseUrl);
	resp = await getCurrent(client, baseUrl);

	const form = await superValidate(client, zod(formSchema));
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

const initializeUser = async (baseUrl: string) => {
	// Your user initialization logic here
	const userResponse = await fetch(`${baseUrl}/api/user`);
	if (userResponse.ok) {
		client = await userResponse.json();
	}
};
type userCreds = {
    github: string;
    aws: { accessKey: string; secretKey: string };
    openai: string;
    demo: { endpoint: string; apiKey: string };
    modal: string;
};
async function getCurrent(user: User | null, baseUrl: string): Promise<userCreds> {
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
		const response = await fetch(`${baseUrl}/api/user/${user.id}/creds`, {
			method: 'GET',
			credentials: 'include'
        });
        if (response.ok) {
            const resp = await response.json();
            /* Returns USER OBJ
            EX:
            "user":{"UUID":"1681fbd4-35e7-4008-a5fc-f5feffb9b0f4","Name":"KooshaPari","Email":"kooshapari@gmail.com","Password":"","AwsCreds":{"AccessKeyID":"++u9Lhjy0d+rWPiUZs+tglZYEWc7B7866rIOXizm2nJdtzPv","SecretAccessKey":"xRhl4BpmWeIV0w9Rt1bcHbEtVE5qWVIE0PKg5jxgC9hssZsAjYVPZjjyjAihvzIowFw9h8zBeSA="},"OpenAICreds":{"APIKey":"s+QLwESliw+M6u/8DZChi3mhoOnetMvggG6+QELLIa3TE46qFyO5W4W7sAAG6CPdzX7M6y8uPpmhzUp2oWnohXDTreeNld/leHHbazvbppS0HBQs/lk91ZPoarYbE8vDqnRthHRjU7h2QR1xXcXnEMyEAlg+i1YtxxkpUMPG/9sT7PWP34trV3pbc9r0x8izDijqnXc9MCEOY8I3YoUM2UnODk6bataHl7WmQ19MxeVr856r"},"Portfolio": {
            "RootEndpoint": "RXYtdFs8gAuHFwj25VffGgzxoOAPpwDNisU4ZNTlhMLj2Uk6elM=",
            "APIKey": "S8hV2PY1LZr5H7L6THzCf/DSrNM3AFsy9P0QvdI="
        },"Git":{"Token":"ljXUoO4w4izS3opN0lTWQfNg4j8q+RIYXDs3kP9Us8DAZ0kkWelz+0E0wC/koufFq/CUXGSjuYc=","RefreshToken":"NXgl6gaL4bFy4OfoMUfqdSTboaNrwRqX1aAhSUTnmuX7eJcxmF0MqyR7WSmJK6OrXJ3NBCA7EnHunbIuSj/t9pn4SwhzJWHtruQcBjwKaEz0LPhpU/MTJsbCPwt9zWot","TokenExpiry":"2024-12-17T22:15:52.2406586-08:00","RefreshTokenExpiry":"2025-04-17T03:30:52.2406586-07:00","Repositories":null},"Projects":null,"Instances":null},  */
            return {
                github: resp.Git.Token,
                aws: { accessKey: resp.AwsCreds.AccessKeyID, secretKey: resp.AwsCreds.SecretAccessKey },
                openai: resp.OpenAICreds.APIKey,
                demo: { endpoint: resp.Portfolio.RootEndpoint, apiKey: resp.Portfolio.APIKey },
                modal: resp.OpenAICreds.APIKey === '' ? 'openai' : 'local'
            };
        }
    
}
