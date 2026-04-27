<script lang="ts">
	import { Input } from '$lib/components/ui/input';
	import Check from 'lucide-svelte/icons/check';
	import ChevronsUpDown from 'lucide-svelte/icons/chevrons-up-down';
	import { tick } from 'svelte';
	import * as Command from '$lib/components/ui/command/index.js';
	import * as Popover from '$lib/components/ui/popover/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { cn } from '$lib/utils.js';
	import 
	const platformArray = ['web', 'mobile', 'desktop'];

	const typeArray = ['single-page', 'multi-page', 'plugin', 'service'];

</script>

<form method="POST">
	<Input name="name" />
	<Input name="description" />
	<Input name="nvms" />
	<Combo items={platformArray} placeholder="Platform" />
	<Combo items={typeArray} placeholder="Type" />
</form>
