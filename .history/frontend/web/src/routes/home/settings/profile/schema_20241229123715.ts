import { z } from 'zod';

 

export const formSchema = z.object({
	name: z.string({
		required_error: 'Required.'
	}),
	email: z.string({
		required_error: 'Required.'
	}),
	password: z.object({
		password: z.string({
			required_error: 'Required.'
		}),
		ConfirmPassword: z.string({
			required_error: 'Required.'
		})
	}),
	 
});

export type FormSchema = typeof formSchema;
