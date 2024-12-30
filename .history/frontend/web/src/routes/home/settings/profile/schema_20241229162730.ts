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
		confirmPassword: z.string({
			required_error: 'Required.'
		})
	}).superRefine(({ confirmPassword, password }, ctx) => {
  if (confirmPassword !== password) {
    ctx.addIssue({
      code: "custom",
      message: "The passwords did not match",
      path: ['confirmPassword']
    });
  }
}),
});

export type FormSchema = typeof formSchema;
