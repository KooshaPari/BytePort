<script lang="ts">
	import * as Form from '$lib/components/ui/form';
	import { Input } from '$lib/components/ui/input';
	import * as Select from '$lib/components/ui/select';
	import type { Repository } from '$lib/git';

	const projectTypes = [
		{ value: 'web', label: 'Web Application' },
		{ value: 'api', label: 'API Service' },
		{ value: 'monorepo', label: 'Monorepo' },
		{ value: 'library', label: 'Library' },
		{ value: 'other', label: 'Other' }
	];

	const platformOptions = [
		{ value: 'aws', label: 'AWS' },
		{ value: 'gcp', label: 'Google Cloud' },
		{ value: 'azure', label: 'Azure' },
		{ value: 'local', label: 'Local' }
	];

	let selectedType = '';
	let selectedPlatform = '';
	let projectName = '';
	let projectDescription = '';
</script>

<form id="form-stage-2" class="space-y-4 w-full max-w-md">
	<div class="form-group">
		<label for="name" class="text-sm font-medium text-white block mb-2">Project Name</label>
		<Input
			id="name"
			name="name"
			type="text"
			placeholder="Enter project name"
			bind:value={projectName}
			required
			class="w-full"
		/>
	</div>

	<div class="form-group">
		<label for="description" class="text-sm font-medium text-white block mb-2">Description</label>
		<textarea
			id="description"
			name="description"
			placeholder="Describe your project"
			bind:value={projectDescription}
			rows="3"
			class="w-full px-3 py-2 border border-gray-600 rounded-md bg-gray-900 text-white placeholder-gray-500"
		></textarea>
	</div>

	<div class="form-group">
		<label for="type" class="text-sm font-medium text-white block mb-2">Project Type</label>
		<select
			id="type"
			name="type"
			bind:value={selectedType}
			required
			class="w-full px-3 py-2 border border-gray-600 rounded-md bg-gray-900 text-white"
		>
			<option value="" disabled selected>Select a project type</option>
			{#each projectTypes as type}
				<option value={type.value}>{type.label}</option>
			{/each}
		</select>
	</div>

	<div class="form-group">
		<label for="platform" class="text-sm font-medium text-white block mb-2">Platform</label>
		<select
			id="platform"
			name="platform"
			bind:value={selectedPlatform}
			required
			class="w-full px-3 py-2 border border-gray-600 rounded-md bg-gray-900 text-white"
		>
			<option value="" disabled selected>Select a platform</option>
			{#each platformOptions as platform}
				<option value={platform.value}>{platform.label}</option>
			{/each}
		</select>
	</div>
</form>

<style>
	:global(.text-dark-onSurfaceVariant) {
		color: rgba(255, 255, 255, 0.7);
	}
</style>
