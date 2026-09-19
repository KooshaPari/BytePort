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
export function setUser(authenticated: boolean, userData: User | null = null) {
	if (authenticated && userData) {
		console.log('setting store: ', userData);
		user.set({
			status: 'authenticated',
			data: userData
		});
	} else {
		console.log('setting store fail: ', userData);
		user.set({ status: 'unauthenticated', data: null });
	}
}
export async function initializeUser() {
	try {
		console.log('Initializing user');
		const data = await apiFetch<{ User: User }>('/authenticate');
		console.log('JS: ', data);
		const authenticatedUser: User = data.User; // Adjust based on backend response structure
		console.log('Authenticated user:', authenticatedUser);
		setUser(true, authenticatedUser);
	} catch (error) {
		console.error('Error validating token:', error);
		setUser(false);
		localStorage.removeItem('authToken');
	}
}
