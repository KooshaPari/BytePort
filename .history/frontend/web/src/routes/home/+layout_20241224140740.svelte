<script lang="ts">
	import '../../app.css';
	import { user } from '../../stores/user';
	let { children } = $props();
	export const prerender = true;
	export const ssr = false;
</script>

{@render children()}
