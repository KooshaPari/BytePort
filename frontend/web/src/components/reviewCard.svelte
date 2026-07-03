<script lang="ts">
	import * as Card from '$lib/components/ui/card';
	import type { Project } from '$lib/utils';
	import Icon from '@iconify/svelte';

	export let project: Project;

	const getTypeIcon = (type: string) => {
		const icons: Record<string, string> = {
			web: 'mdi:web',
			api: 'mdi:api',
			monorepo: 'mdi:folder-multiple',
			library: 'mdi:package',
			other: 'mdi:folder'
		};
		return icons[type] || 'mdi:folder';
	};

	const getPlatformIcon = (platform: string) => {
		const icons: Record<string, string> = {
			aws: 'mdi:aws',
			gcp: 'mdi:google-cloud',
			azure: 'mdi:microsoft-azure',
			local: 'mdi:server'
		};
		return icons[platform] || 'mdi:server';
	};
</script>

<Card.Root class="w-full max-w-md">
	<Card.Header>
		<Card.Title>Review Project Details</Card.Title>
		<Card.Description>Verify all details before deployment</Card.Description>
	</Card.Header>
	<Card.Content class="space-y-4">
		<div class="review-item border-b border-gray-600 pb-3">
			<div class="flex items-center gap-2 mb-1">
				<Icon icon="mdi:text" class="text-white" />
				<span class="text-xs text-gray-400">Project Name</span>
			</div>
			<p class="text-white font-medium">{project.name}</p>
		</div>

		<div class="review-item border-b border-gray-600 pb-3">
			<div class="flex items-center gap-2 mb-1">
				<Icon icon="mdi:description" class="text-white" />
				<span class="text-xs text-gray-400">Description</span>
			</div>
			<p class="text-white">{project.description || 'No description provided'}</p>
		</div>

		<div class="review-item border-b border-gray-600 pb-3">
			<div class="flex items-center gap-2 mb-1">
				<Icon icon={getTypeIcon(project.Type)} class="text-white" />
				<span class="text-xs text-gray-400">Project Type</span>
			</div>
			<p class="text-white font-medium capitalize">{project.Type}</p>
		</div>

		<div class="review-item border-b border-gray-600 pb-3">
			<div class="flex items-center gap-2 mb-1">
				<Icon icon={getPlatformIcon(project.Platform)} class="text-white" />
				<span class="text-xs text-gray-400">Platform</span>
			</div>
			<p class="text-white font-medium capitalize">{project.Platform}</p>
		</div>

		{#if project.Repository}
			<div class="review-item border-b border-gray-600 pb-3">
				<div class="flex items-center gap-2 mb-1">
					<Icon icon="mdi:github" class="text-white" />
					<span class="text-xs text-gray-400">Repository</span>
				</div>
				<p class="text-white font-medium">{project.Repository.full_name}</p>
				<a
					href={project.Repository.html_url}
					target="_blank"
					rel="noopener noreferrer"
					class="text-blue-400 hover:text-blue-300 text-sm mt-1"
				>
					{project.Repository.html_url}
				</a>
			</div>
		{/if}

		{#if project.access_url}
			<div class="review-item">
				<div class="flex items-center gap-2 mb-1">
					<Icon icon="mdi:link" class="text-white" />
					<span class="text-xs text-gray-400">Access URL</span>
				</div>
				<a
					href={project.access_url}
					target="_blank"
					rel="noopener noreferrer"
					class="text-blue-400 hover:text-blue-300 break-all"
				>
					{project.access_url}
				</a>
			</div>
		{/if}
	</Card.Content>
</Card.Root>

<style>
	:global(.text-dark-onSurfaceVariant) {
		color: rgba(255, 255, 255, 0.7);
	}
</style>
