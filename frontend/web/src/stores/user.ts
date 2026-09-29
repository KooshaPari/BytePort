import { writable } from 'svelte/store';
import { apiFetch } from '../lib/api';

// Define the User interface
export interface UserLink {
	// Personal Details
	UUID: string;
	Name: string;
	Email: string;

	// API Details
	awsCreds: {
		accessKeyId: string;
		secretAccessKey: string;
	};
	llmConfig: {
		provider: string;
		providers: {
			[provider: string]: {
				modal: string;
				apiKey: string;
			};
		};
	};
	portfolio: {
		rootEndpoint: string;
		apiKey: string;
	};
}
export interface User {
	// Personal Details
	uuid: string;
	name: string;
	email: string;
}
type UserStore = {
	status: 'pending' | 'authenticated' | 'unauthenticated';
	data: User | null;
};

export const user = writable<UserStore>({
	status: 'pending',
	data: null
});
export function setAuthenticatedUser(userData: User) {
	console.log('setting store: ', userData);
	user.set({
		status: 'authenticated',
		data: userData
	});
}
export function setUnauthenticatedUser() {
	console.log('setting store fail');
	user.set({ status: 'unauthenticated', data: null });
}
export async function initializeUser() {
	try {
		console.log('Initializing user');
		const data = await apiFetch<{ User: User }>('/authenticate');
		console.log('JS: ', data);
		const authenticatedUser: User = data.User; // Adjust based on backend response structure
		console.log('Authenticated user:', authenticatedUser);
		setAuthenticatedUser(authenticatedUser);
	} catch (error) {
		console.error('Error validating token:', error);
		setUnauthenticatedUser();
		localStorage.removeItem('authToken');
	}
}
