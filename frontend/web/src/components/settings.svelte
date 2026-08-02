<script lang="ts">
	import Icon from '@iconify/svelte';
	import * as Card from '$lib/components/ui/card';
	import * as Button from '$lib/components/ui/button';

	export let onClose: () => void = () => {};

	const settingsSections = [
		{
			id: 'general',
			label: 'General',
			icon: 'mdi:cog',
			description: 'General application settings'
		},
		{
			id: 'integrations',
			label: 'Integrations',
			icon: 'mdi:plug',
			description: 'Manage connected services'
		},
		{
			id: 'notifications',
			label: 'Notifications',
			icon: 'mdi:bell',
			description: 'Control notification preferences'
		},
		{
			id: 'security',
			label: 'Security',
			icon: 'mdi:shield',
			description: 'Manage security settings'
		},
		{
			id: 'about',
			label: 'About',
			icon: 'mdi:information',
			description: 'Application information'
		}
	];

	let selectedSection = 'general';
</script>

<div class="w-full max-w-2xl space-y-4">
	<Card.Root>
		<Card.Header>
			<Card.Title>Settings</Card.Title>
			<Card.Description>Manage your application preferences</Card.Description>
		</Card.Header>
		<Card.Content class="space-y-4">
			<div class="grid grid-cols-1 md:grid-cols-2 gap-3">
				{#each settingsSections as section (section.id)}
					<button
						on:click={() => (selectedSection = section.id)}
						class={`p-4 rounded-lg border transition-all text-left ${
							selectedSection === section.id
								? 'border-blue-500 bg-blue-500/10'
								: 'border-gray-600 hover:border-gray-500'
						}`}
					>
						<div class="flex items-center gap-3 mb-2">
							<Icon icon={section.icon} class="text-xl text-white" />
							<span class="font-semibold text-white">{section.label}</span>
						</div>
						<p class="text-sm text-gray-400">{section.description}</p>
					</button>
				{/each}
			</div>

			<div class="mt-6 pt-4 border-t border-gray-600">
				{#if selectedSection === 'general'}
					<div class="space-y-4">
						<h3 class="font-semibold text-white">General Settings</h3>
						<p class="text-gray-400">Configure basic application behavior</p>
					</div>
				{:else if selectedSection === 'integrations'}
					<div class="space-y-4">
						<h3 class="font-semibold text-white">Integrations</h3>
						<p class="text-gray-400">
							Connect external services like GitHub, AWS, etc.
						</p>
					</div>
				{:else if selectedSection === 'notifications'}
					<div class="space-y-4">
						<h3 class="font-semibold text-white">Notification Preferences</h3>
						<p class="text-gray-400">Control how and when you receive notifications</p>
					</div>
				{:else if selectedSection === 'security'}
					<div class="space-y-4">
						<h3 class="font-semibold text-white">Security Settings</h3>
						<p class="text-gray-400">Manage your account security and privacy</p>
					</div>
				{:else if selectedSection === 'about'}
					<div class="space-y-4">
						<h3 class="font-semibold text-white">About BytePort</h3>
						<p class="text-gray-400">Application version and information</p>
						<div class="bg-gray-900 p-3 rounded-lg text-sm text-gray-400">
							Version 0.1.0 - Cross-platform data transport application
						</div>
					</div>
				{/if}
			</div>
		</Card.Content>
	</Card.Root>

	<div class="flex justify-end gap-2">
		<Button.Root variant="outline" on:click={onClose}>Close</Button.Root>
	</div>
</div>

<style>
	:global(.text-dark-onSurfaceVariant) {
		color: rgba(255, 255, 255, 0.7);
	}
</style>
