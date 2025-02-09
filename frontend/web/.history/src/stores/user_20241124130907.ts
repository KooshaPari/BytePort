import { writable } from 'svelte/store';

// Define the User interface
export interface User {
	// Personal Details
	UUID: string;
	Name: string;
	Email: string;
	Encrypted_Token: string;

	// API Details
	aws: {
		accessKeyId: string;
		secretAccessKey: string;
	};
	openAI: {
		apiKey: string;
	};
	portfolio: {
		rootEndpoint: string;
		apiKey: string;
	};
	git: {
		repoUrl: string;
		authMethod: string; // e.g., "ssh" or "token"
		authKey: string;
		targetDirectory: string;
	};
}
type UserStore = {
	status: 'pending' | 'authenticated' | 'unauthenticated';
	data: User | null;
};

// Create a writable store with the default user
export const user = writable<UserStore>({
	status: 'pending',
	data: null
});
