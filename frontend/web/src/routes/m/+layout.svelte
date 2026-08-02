<script lang="ts">
	/**
	 * Mobile-first layout for /m/* routes.
	 * - Default (< sm): sticky bottom-nav with 4 icons (home, projects, deploys, settings)
	 * - sm+: desktop companion panel as a left rail; bottom-nav hidden
	 * PILLAR-TAXONOMY-v2.md v2.2 §L43 — mobile app scaffolding (responsive PWA)
	 */
	import { page } from '$app/stores';
	import { t } from '$lib/i18n';

	let { children } = $props();

	type NavItem = { href: string; icon: string; label: string };
	const items: NavItem[] = [
		{ href: '/m', icon: 'home', label: 'nav.home' },
		{ href: '/m/projects', icon: 'projects', label: 'nav.projects' },
		{ href: '/m/deploys', icon: 'deploys', label: 'nav.deploys' },
		{ href: '/m/settings', icon: 'settings', label: 'nav.settings' }
	];

	function isActive(href: string, pathname: string): boolean {
		if (href === '/m') return pathname === '/m';
		return pathname === href || pathname.startsWith(href + '/');
	}
</script>

<div class="m-shell">
	<aside class="rail" aria-label="Primary">
		<a class="brand" href="/m" aria-label="BytePort home">
			<img src="/brand/mascot.svg" alt="" />
			<span>BytePort</span>
		</a>
		<nav>
			{#each items as it (it.href)}
				<a
					class="rail-link"
					class:active={isActive(it.href, $page.url.pathname)}
					href={it.href}
				>
					<span class="ico" aria-hidden="true" data-icon={it.icon}></span>
					<span class="lbl">{$t(it.label)}</span>
				</a>
			{/each}
		</nav>
	</aside>

	<main class="content">
		{@render children()}
	</main>

	<nav class="bottom-nav" aria-label="Primary mobile">
		{#each items as it (it.href)}
			<a
				class="bn-item"
				class:active={isActive(it.href, $page.url.pathname)}
				href={it.href}
				aria-current={isActive(it.href, $page.url.pathname) ? 'page' : undefined}
			>
				<span class="bn-ico" aria-hidden="true" data-icon={it.icon}></span>
				<span class="bn-lbl">{$t(it.label)}</span>
			</a>
		{/each}
	</nav>
</div>

<style>
	.m-shell {
		min-height: 100dvh;
		display: flex;
		flex-direction: column;
		background: var(--background, #0e1514);
		color: var(--foreground, #dde4e2);
		padding-bottom: calc(72px + env(safe-area-inset-bottom));
	}
	.content {
		flex: 1;
		width: 100%;
		margin: 0 auto;
		padding: 1rem;
		max-width: 100%;
	}

	/* Bottom nav (mobile) */
	.bottom-nav {
		position: fixed;
		inset: auto 0 0 0;
		z-index: 30;
		display: grid;
		grid-template-columns: repeat(4, 1fr);
		height: calc(64px + env(safe-area-inset-bottom));
		padding-bottom: env(safe-area-inset-bottom);
		background: var(--surface, #101418);
		border-top: 1px solid var(--border, #3f4948);
	}
	.bn-item {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		min-height: 44px;
		min-width: 44px;
		color: var(--muted-foreground, #bec9c7);
		text-decoration: none;
		font-size: 0.7rem;
		gap: 2px;
		transition: color 160ms;
	}
	.bn-item.active {
		color: var(--primary, #80d5cf);
	}
	.bn-ico {
		width: 22px;
		height: 22px;
		background: currentColor;
		-webkit-mask: center/contain no-repeat;
		mask: center/contain no-repeat;
	}
	.bn-ico[data-icon='home'] {
		-webkit-mask-image: url("data:image/svg+xml;utf8,<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='currentColor'><path d='M12 3l9 8h-3v9h-5v-6H11v6H6v-9H3z'/></svg>");
		mask-image: url("data:image/svg+xml;utf8,<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='currentColor'><path d='M12 3l9 8h-3v9h-5v-6H11v6H6v-9H3z'/></svg>");
	}
	.bn-ico[data-icon='projects'] {
		-webkit-mask-image: url("data:image/svg+xml;utf8,<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='currentColor'><path d='M3 5h7v2h11v13H3z'/></svg>");
		mask-image: url("data:image/svg+xml;utf8,<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='currentColor'><path d='M3 5h7v2h11v13H3z'/></svg>");
	}
	.bn-ico[data-icon='deploys'] {
		-webkit-mask-image: url("data:image/svg+xml;utf8,<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='currentColor'><path d='M12 2l4 8h-3v6h-2v-6H8zM5 18h14v2H5z'/></svg>");
		mask-image: url("data:image/svg+xml;utf8,<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='currentColor'><path d='M12 2l4 8h-3v6h-2v-6H8zM5 18h14v2H5z'/></svg>");
	}
	.bn-ico[data-icon='settings'] {
		-webkit-mask-image: url("data:image/svg+xml;utf8,<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='currentColor'><path d='M19.4 13a7.7 7.7 0 0 0 0-2l2-1.5-2-3.4-2.3.9a7.6 7.6 0 0 0-1.7-1L15 3.5h-4l-.4 2.6a7.6 7.6 0 0 0-1.7 1l-2.3-.9-2 3.4 2 1.5a7.7 7.7 0 0 0 0 2l-2 1.5 2 3.4 2.3-.9a7.6 7.6 0 0 0 1.7 1l.4 2.5h4l.4-2.5a7.6 7.6 0 0 0 1.7-1l2.3.9 2-3.4zM12 15.5a3.5 3.5 0 1 1 0-7 3.5 3.5 0 0 1 0 7z'/></svg>");
		mask-image: url("data:image/svg+xml;utf8,<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='currentColor'><path d='M19.4 13a7.7 7.7 0 0 0 0-2l2-1.5-2-3.4-2.3.9a7.6 7.6 0 0 0-1.7-1L15 3.5h-4l-.4 2.6a7.6 7.6 0 0 0-1.7 1l-2.3-.9-2 3.4 2 1.5a7.7 7.7 0 0 0 0 2l-2 1.5 2 3.4 2.3-.9a7.6 7.6 0 0 0 1.7 1l.4 2.5h4l.4-2.5a7.6 7.6 0 0 0 1.7-1l2.3.9 2-3.4zM12 15.5a3.5 3.5 0 1 1 0-7 3.5 3.5 0 0 1 0 7z'/></svg>");
	}

	/* Desktop companion rail (>= sm) */
	.rail {
		display: none;
	}
	.rail .brand {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 1rem;
		text-decoration: none;
		color: var(--foreground, #dde4e2);
		font-weight: 700;
	}
	.rail .brand img {
		width: 36px;
		height: 36px;
	}
	.rail nav {
		display: flex;
		flex-direction: column;
		gap: 4px;
		padding: 0.5rem;
	}
	.rail-link {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		min-height: 44px;
		padding: 0 0.75rem;
		border-radius: 0.5rem;
		text-decoration: none;
		color: var(--muted-foreground, #bec9c7);
		transition:
			background 160ms,
			color 160ms;
	}
	.rail-link:hover {
		background: var(--muted, #101418);
		color: var(--foreground, #dde4e2);
	}
	.rail-link.active {
		background: var(--primary-container, #00504c);
		color: var(--on-primary-container, #9df1eb);
	}
	.ico {
		width: 20px;
		height: 20px;
		flex: 0 0 20px;
	}

	@media (min-width: 640px) {
		.m-shell {
			flex-direction: row;
			padding-bottom: 0;
		}
		.rail {
			display: block;
			width: 220px;
			flex-shrink: 0;
			border-right: 1px solid var(--border, #3f4948);
			background: var(--surface, #101418);
			height: 100dvh;
			position: sticky;
			top: 0;
		}
		.content {
			max-width: calc(100% - 220px);
			padding: 1.5rem 2rem;
		}
		.bottom-nav {
			display: none;
		}
	}
</style>
