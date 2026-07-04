<script lang="ts">
  /**
   * Settings/language panel — exposes LocaleSwitcher alongside theme + a11y controls.
   */
  import LocaleSwitcher from '$lib/components/LocaleSwitcher.svelte';
  import { t } from '$lib/i18n';
</script>

<section aria-labelledby="settings-heading" class="space-y-6">
  <h1 id="settings-heading" class="text-2xl font-semibold">{$t('settings.title')}</h1>

  <div class="card bg-base-200 p-4 space-y-2">
    <h2 class="text-sm font-medium uppercase tracking-wide opacity-70">
      {$t('settings.language')}
    </h2>
    <div class="flex items-center justify-between">
      <span id="lang-help" class="text-sm opacity-80">
        {$t('settings.language')}
      </span>
      <LocaleSwitcher />
    </div>
  </div>
</section>