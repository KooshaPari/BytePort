import { z } from 'zod';

/**
 * Profile form schema.
 *
 * The password pair is optional here on purpose. The backend only rehashes when
 * a non-empty password arrives (backend/byteport/routes/auth.go UpdateUser),
 * and the previous schema required both password fields, which would have
 * blocked a display-name edit until the user invented a new password.
 *
 * Rules kept from the original: name and email are required, and the two
 * password fields must match. Added: a minimum length whenever a password is
 * actually being set.
 */
export const formSchema = z
	.object({
		name: z.string().trim().min(1, 'Name is required.'),
		email: z.string().trim().min(1, 'Email is required.'),
		password: z.string(),
		confirmPassword: z.string()
	})
	.superRefine(({ password, confirmPassword }, ctx) => {
		if (password !== confirmPassword) {
			ctx.addIssue({
				code: 'custom',
				message: 'The passwords did not match.',
				path: ['confirmPassword']
			});
		}
		if (password.length > 0 && password.length < 8) {
			ctx.addIssue({
				code: 'custom',
				message: 'Use at least 8 characters.',
				path: ['password']
			});
		}
	});

export type ProfileFormSchema = typeof formSchema;
