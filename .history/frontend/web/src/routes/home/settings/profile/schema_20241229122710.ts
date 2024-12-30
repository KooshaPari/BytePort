import { z } from 'zod';

 

export const formSchema = z.object({
	Name: z.string({
		required_error: 'Required.'
	}),
	 
	 
});

export type FormSchema = typeof formSchema;
