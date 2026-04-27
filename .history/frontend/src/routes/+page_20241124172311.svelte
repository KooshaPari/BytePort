<script lang="ts">
	import Icon from '@iconify/svelte';
</script>

<div class="bg-dark-surface flex h-screen w-screen overflow-x-hidden" id="parent">
	<div id="header" class=" w-5/5 bg-dark-surfaceContainerLow h-1/5 flex-col justify-between ps-2.5">
		<div id="headerNav" class="h-3/5 pt-2.5">
			<div class="flex justify-end pe-2.5" id="navRight">
				<Icon
					class="hover:text-dark-primary active:text-dark-surfaceBright mx-1 h-6 w-6 cursor-pointer text-white"
					icon="ic:baseline-notifications"
				/>
				<Icon
					class="hover:text-dark-primary active:text-dark-surfaceBright mx-1 h-6 w-6 cursor-pointer text-white"
					icon="ic:baseline-account-circle"
				/>
			</div>
		</div>
		<div id="headerContent" class="h-2/5 text-4xl text-white">Hello.</div>
	</div>
	<div id="body" class="w-4/5">
		<div id="mainBody"></div>
		<div id="footer"></div>
	</div>
</div>

<style>
</style>
