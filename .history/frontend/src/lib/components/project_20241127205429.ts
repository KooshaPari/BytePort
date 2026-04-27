import { z } from 'zod';

const projectSchema = z.object({
	projectName: z.string().min(1, 'Project name is required'),
	description: z.string().min(1, 'Description is required'),
	platform: z.enum(
		['web', 'mobile', 'desktop', 'other'],
		"Platform must be one of 'web', 'mobile', 'desktop', or 'other'"
	),
	type: z.enum(['frontend', 'backend', 'fullstack', 'utility'], {
		invalid_message: "Type must be one of 'frontend', 'backend', 'fullstack', or 'utility'"
	}),
	nvmsPath: z.string().url('Invalid URL for NVMS path')
});

export default projectSchema;
