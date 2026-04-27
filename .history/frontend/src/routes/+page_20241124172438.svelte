<script lang="ts">
	import Icon from '@iconify/svelte';
</script>

<div class="bg-dark-surface flex h-screen w-screen overflow-x-hidden" id="parent">
	<div id="header" class=" w-5/5 bg-dark-surfaceContainerLow h-1/5 flex-col justify-between ps-2.5">
		<div id="headerNav" class="h-3/5 w-screen pt-2.5">
			<img class="h-4/5 py-10" alt="BytePort" src="/src/assets/img/byte.png" />
			<div class="flex justify-end pe-2.5" id="navRight">
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
	</div>
	<div id="footer"></div>
</div>

<style>
</style>
