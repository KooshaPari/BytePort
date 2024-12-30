import { z } from 'zod';

 

export const formSchema = z.object({
	github: z.string({
		required_error: 'Required.'
	}),
	aws: z.object({
		accessKey: z.string({
			required_error: 'Required.'
		}),
		secretKey: z.string({
			required_error: 'Required.'
		})
	}),
	 
});

export type FormSchema = typeof formSchema;
