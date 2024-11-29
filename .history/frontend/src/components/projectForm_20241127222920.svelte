<script lang="ts">
	import { Input } from '$lib/components/ui/input';
	import Check from 'lucide-svelte/icons/check';
	import ChevronsUpDown from 'lucide-svelte/icons/chevrons-up-down';
	import { tick } from 'svelte';
	import * as Command from '$lib/components/ui/command/index.js';
	import * as Popover from '$lib/components/ui/popover/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { cn } from '$lib/utils.js';
	import Combo from './combo.svelte';
	const platformArray = [
		{ value: 'web', label: 'Web' },
		{ value: 'mobile', label: 'Mobile' },
		{ value: 'desktop', label: 'Desktop' }
	];

	const typeArray = [
		{ value: 'single-page', label: 'Single Page' },
		{ value: 'multi-page', label: 'Multi Page' },
		{ value: 'plugin', label: 'Plugin' },
		{ value: 'service', label: 'Service' }
	];
</script>

<form id="form-stage-2">
	<Input placeholder="Project Name" name="name" />
	<Input placeholder="Description" name="description" />
	<Input placeholder="NVMS Path" name="nvms" />
	<input type="hidden" name="typeValue" bind:value={typeValue} />
	<input type="hidden" name="platformValue" bind:value={platformValue} />
	<Combo items={platformArray} on:change={(e) => (platformValue = e.detail)} placeholder="Platform" />
	<Combo items={typeArray} placeholder="Type" />
</form>
