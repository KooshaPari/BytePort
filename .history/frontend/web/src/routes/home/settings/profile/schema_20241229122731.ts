import { z } from 'zod';

 

export const formSchema = z.object({
	Name: z.string({
		required_error: 'Required.'
	}),
	Email: z.string({
		required_error: 'Required.'
	}),
	Password: z.object({
		required_error: 'Required.'
	}),
	 
});

export type FormSchema = typeof formSchema;
