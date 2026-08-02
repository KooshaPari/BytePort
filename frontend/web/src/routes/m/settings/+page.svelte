<script lang="ts">
	/**
	 * Mobile Settings — grouped list of toggles/links.
	 * Mirrors deploys/ touch patterns with sections.
	 *
	 * PILLAR-TAXONOMY-v2.md v2.2 §L43
	 */
	import { t, locale } from '$lib/i18n';
	import LocaleSwitcher from '$lib/components/LocaleSwitcher.svelte';

	let pushEnabled = $state(true);
	let darkMode = $state(false);
	let biometricLock = $state(false);
	let analytics = $state(false);
</script>

<header class="px-4 pt-4 pb-3 bg-base-200">
	<h1 class="text-xl font-semibold">{$t('nav.settings')}</h1>
</header>

<main class="px-4 py-3 pb-24 space-y-4">
	<!-- Section: Account -->
	<section aria-labelledby="sect-account">
		<h2 id="sect-account" class="px-1 mb-2 text-xs uppercase tracking-wide opacity-60">
			{$t('settings.account')}
		</h2>
		<div class="card card-compact bg-base-200 shadow-sm">
			<div class="card-body divide-y divide-base-300">
				<a href="/m/settings/profile" class="flex items-center justify-between py-3 px-1">
					<span>{$t('settings.profile')}</span>
					<span aria-hidden="true" class="opacity-50">›</span>
				</a>
				<a href="/m/settings/billing" class="flex items-center justify-between py-3 px-1">
					<span>{$t('settings.billing')}</span>
					<span aria-hidden="true" class="opacity-50">›</span>
				</a>
			</div>
		</div>
	</section>

	<!-- Section: Notifications -->
	<section aria-labelledby="sect-notif">
		<h2 id="sect-notif" class="px-1 mb-2 text-xs uppercase tracking-wide opacity-60">
			{$t('settings.notifications')}
		</h2>
		<div class="card card-compact bg-base-200 shadow-sm">
			<div class="card-body divide-y divide-base-300">
				<label class="flex items-center justify-between py-3 px-1">
					<span>{$t('settings.push')}</span>
					<input
						type="checkbox"
						bind:checked={pushEnabled}
						class="toggle toggle-primary"
						aria-label={$t('settings.push')}
					/>
				</label>
				<label class="flex items-center justify-between py-3 px-1">
					<span>{$t('settings.analytics')}</span>
					<input
						type="checkbox"
						bind:checked={analytics}
						class="toggle toggle-primary"
						aria-label={$t('settings.analytics')}
					/>
				</label>
			</div>
		</div>
	</section>

	<!-- Section: Appearance -->
	<section aria-labelledby="sect-appear">
		<h2 id="sect-appear" class="px-1 mb-2 text-xs uppercase tracking-wide opacity-60">
			{$t('settings.appearance')}
		</h2>
		<div class="card card-compact bg-base-200 shadow-sm">
			<div class="card-body divide-y divide-base-300">
				<label class="flex items-center justify-between py-3 px-1">
					<span>{$t('settings.dark')}</span>
					<input
						type="checkbox"
						bind:checked={darkMode}
						class="toggle toggle-primary"
						aria-label={$t('settings.dark')}
					/>
				</label>
				<div class="flex items-center justify-between py-3 px-1">
					<span>{$t('settings.language')}</span>
					<LocaleSwitcher />
				</div>
			</div>
		</div>
	</section>

	<!-- Section: Security -->
	<section aria-labelledby="sect-sec">
		<h2 id="sect-sec" class="px-1 mb-2 text-xs uppercase tracking-wide opacity-60">
			{$t('settings.security')}
		</h2>
		<div class="card card-compact bg-base-200 shadow-sm">
			<div class="card-body divide-y divide-base-300">
				<label class="flex items-center justify-between py-3 px-1">
					<span>{$t('settings.biometric')}</span>
					<input
						type="checkbox"
						bind:checked={biometricLock}
						class="toggle toggle-primary"
						aria-label={$t('settings.biometric')}
					/>
				</label>
				<a href="/m/settings/keys" class="flex items-center justify-between py-3 px-1">
					<span>{$t('settings.apiKeys')}</span>
					<span aria-hidden="true" class="opacity-50">›</span>
				</a>
			</div>
		</div>
	</section>

	<p class="text-center text-xs opacity-50 pt-2">byteport · locale={$locale}</p>
</main>
