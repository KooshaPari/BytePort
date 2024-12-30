import { z } from 'zod';

 

export const formSchema = z.object({
	name: z.string({
		required_error: 'Required.'
	}),
	email: z.string({
		required_error: 'Required.'
	}),
	Password: z.object({
		Password: z.string({
			required_error: 'Required.'
		}),
		ConfirmPassword: z.string({
			required_error: 'Required.'
		})
	}),
	 
});

export type FormSchema = typeof formSchema;
