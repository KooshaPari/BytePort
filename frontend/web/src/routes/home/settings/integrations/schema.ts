import { z } from 'zod';

/**
 * Integrations form schema.
 *
 * The previous version mirrored the wire payload and marked every credential
 * as required, so a user could not save an AWS key without also having an LLM
 * key and a portfolio endpoint, even though the backend validates each of
 * those independently. The rules here match what the backend actually
 * requires: the two halves of a credential pair must be filled together,
 * everything else is optional.
 *
 * The wire shape is unrelated to this schema and is built explicitly in
 * +page.svelte so the POST /link contract stays visible in one place.
 */
export const formSchema = z
	.object({
		awsAccessKey: z.string().trim(),
		awsSecretKey: z.string().trim(),
		provider: z.string(),
		portfolioEndpoint: z.string().trim(),
		portfolioKey: z.string().trim()
	})
	.superRefine(({ awsAccessKey, awsSecretKey, portfolioEndpoint, portfolioKey }, ctx) => {
		if (awsAccessKey.length > 0 !== awsSecretKey.length > 0) {
			ctx.addIssue({
				code: 'custom',
				message: 'Provide both the access key ID and the secret access key.',
				path: [awsAccessKey.length > 0 ? 'awsSecretKey' : 'awsAccessKey']
			});
		}
		if (portfolioEndpoint.length > 0 !== portfolioKey.length > 0) {
			ctx.addIssue({
				code: 'custom',
				message: 'Provide both the endpoint URL and the API key.',
				path: [portfolioEndpoint.length > 0 ? 'portfolioKey' : 'portfolioEndpoint']
			});
		}
	});

export type IntegrationsFormSchema = typeof formSchema;
