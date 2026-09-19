<script lang="ts">
	/**
	 * Integration settings.
	 *
	 * Was one unlabelled horizontal row of controls (GitHub button, two AWS
	 * inputs, a provider/model combobox stack, two portfolio fields) with no
	 * section headings, no descriptions, and a save button whose result was
	 * never read. It also flipped its own "Linked" label the moment the GitHub
	 * popup opened, and offered an unlink button that called no endpoint.
	 *
	 * Now: four labelled sections with one-line descriptions and explicit
	 * saving, saved and failed feedback. GitHub state comes from the backend
	 * and refreshes after the link popup closes.
	 *
	 * Endpoints unchanged: GET /user/:id/creds, POST /link, GET /link.
	 */
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { user, initializeUser } from '../../../../stores/user';
	import { ApiError, apiFetch, apiUrl } from '$lib/api';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import { modals, providers, type SelectOption } from './models';
	import { formSchema } from './schema';

	interface ProviderModel {
		modal: string;
		apiKey: string;
	}

	// Field names are read defensively: the Go models mix tagged and untagged
	// fields, so one value can arrive as `api_key` or as `APIKey`.
	type Json = Record<string, unknown>;

	function asRecord(value: unknown): Json {
		return value !== null && typeof value === 'object' ? (value as Json) : {};
	}

	function firstString(...values: unknown[]): string {
		for (const value of values) {
			if (typeof value === 'string') return value;
		}
		return '';
	}

	/** Merge the tagged and untagged spellings of one nested object. */
	function pick(root: Json, ...keys: string[]): Json {
		return keys.reduce<Json>((merged, key) => ({ ...merged, ...asRecord(root[key]) }), {});
	}

	let githubToken = $state('');
	let awsAccessKey = $state('');
	let awsSecretKey = $state('');
	let provider = $state('');
	let providerModels = $state<Record<string, ProviderModel>>({});
	let portfolioEndpoint = $state('');
	let portfolioKey = $state('');

	let snapshot = $state('');
	let loading = $state(true);
	let saving = $state(false);
	let loadError = $state('');
	let fieldErrors = $state<Record<string, string>>({});
	let feedback = $state<{ kind: 'saved' | 'error'; message: string } | null>(null);
	let linkPending = $state(false);

	const githubLinked = $derived(githubToken.trim().length > 0);

	function currentSnapshot(): string {
		return JSON.stringify({
			githubToken,
			awsAccessKey,
			awsSecretKey,
			provider,
			providerModels,
			portfolioEndpoint,
			portfolioKey
		});
	}

	const dirty = $derived(snapshot !== '' && currentSnapshot() !== snapshot);

	// One place decides what the header reports, so the label and its colour
	// cannot disagree.
	const statusLabel = $derived(
		loading
			? 'Loading'
			: saving
				? 'Saving'
				: feedback?.kind === 'saved'
					? feedback.message
					: dirty
						? 'Unsaved changes'
						: ''
	);
	const statusClass = $derived(
		`text-[12px] ${!loading && !saving && feedback?.kind === 'saved' ? 'text-dark-primary' : 'text-dark-onSurfaceVariant'}`
	);

	const providerOptions = $derived<SelectOption[]>(
		provider.length > 0 && !providers.some((entry) => entry.value === provider)
			? [{ label: `${provider} (saved)`, value: provider }, ...providers]
			: [...providers]
	);

	const modelOptions = $derived.by<SelectOption[]>(() => {
		const known = modals
			.filter((modal) => modal.provider === provider)
			.map((modal) => ({ label: modal.label, value: modal.value }));
		const current = providerModels[provider]?.modal ?? '';
		if (current.length === 0 || known.some((entry) => entry.value === current)) return known;
		return [{ label: `${current} (saved)`, value: current }, ...known];
	});

	async function loadCreds() {
		const uuid = $user.data?.uuid;
		if (!uuid) {
			loading = false;
			loadError = 'No signed-in user, so credentials could not be loaded.';
			return;
		}
		loading = true;
		loadError = '';
		try {
			const body = asRecord(await apiFetch<unknown>(`/user/${uuid}/creds`));
			const aws = pick(body, 'AwsCreds', 'awsCreds');
			const git = pick(body, 'Git', 'git');
			const llm = pick(body, 'LLMConfig', 'llmConfig');
			const portfolio = pick(body, 'Portfolio', 'portfolio');

			githubToken = firstString(git.Token, git.token);
			awsAccessKey = firstString(aws.AccessKeyID, aws.accessKeyId);
			awsSecretKey = firstString(aws.SecretAccessKey, aws.secretAccessKey);
			provider = firstString(llm.provider, llm.Provider);

			const loaded: Record<string, ProviderModel> = {};
			for (const [key, value] of Object.entries(pick(llm, 'providers', 'Providers'))) {
				const entry = asRecord(value);
				loaded[key] = {
					modal: firstString(entry.modal, entry.Modal),
					apiKey: firstString(entry.api_key, entry.apiKey, entry.APIKey)
				};
			}
			providerModels = loaded;
			portfolioEndpoint = firstString(portfolio.RootEndpoint, portfolio.rootEndpoint);
			portfolioKey = firstString(portfolio.APIKey, portfolio.apiKey);
			snapshot = currentSnapshot();
		} catch (error) {
			loadError =
				error instanceof ApiError
					? `Loading credentials failed with HTTP ${error.status}.`
					: 'Could not reach the BytePort backend to load credentials.';
		} finally {
			loading = false;
		}
	}

	function clearFeedback() {
		if (feedback) feedback = null;
	}

	function selectProvider(next: string) {
		provider = next;
		if (next.length > 0 && !providerModels[next]) {
			providerModels = { ...providerModels, [next]: { modal: '', apiKey: '' } };
		}
		clearFeedback();
	}

	function patchProvider(patch: Partial<ProviderModel>) {
		if (provider.length === 0) return;
		const existing = providerModels[provider] ?? { modal: '', apiKey: '' };
		providerModels = { ...providerModels, [provider]: { ...existing, ...patch } };
		clearFeedback();
	}

	async function save() {
		clearFeedback();
		fieldErrors = {};

		const parsed = formSchema.safeParse({
			awsAccessKey,
			awsSecretKey,
			provider,
			portfolioEndpoint,
			portfolioKey
		});
		if (!parsed.success) {
			const errors: Record<string, string> = {};
			for (const issue of parsed.error.issues) {
				const path = issue.path.join('.');
				if (!errors[path]) errors[path] = issue.message;
			}
			fieldErrors = errors;
			feedback = { kind: 'error', message: 'Check the highlighted fields.' };
			return;
		}

		saving = true;
		try {
			// Mirrors UserLink from src/stores/user.ts. `api_key` travels beside
			// `apiKey` because the backend's AIProvider struct tags the field
			// `json:"api_key"` and Go's decoder will not match `apiKey` to it.
			const payload = {
				UUID: $user.data?.uuid ?? '',
				Name: $user.data?.name ?? '',
				Email: $user.data?.email ?? '',
				awsCreds: { accessKeyId: awsAccessKey, secretAccessKey: awsSecretKey },
				llmConfig: {
					provider,
					providers: Object.fromEntries(
						Object.entries(providerModels).map(([key, entry]) => [
							key,
							{ modal: entry.modal, apiKey: entry.apiKey, api_key: entry.apiKey }
						])
					)
				},
				portfolio: { rootEndpoint: portfolioEndpoint, apiKey: portfolioKey }
			};

			await apiFetch('/link', { method: 'POST', body: JSON.stringify(payload) });
			snapshot = currentSnapshot();
			feedback = { kind: 'saved', message: 'Integrations saved.' };
		} catch (error) {
			let detail = '';
			if (error instanceof ApiError) {
				const body = asRecord(error.body);
				detail = firstString(body.details, body.error, body.message);
			}
			feedback = {
				kind: 'error',
				message: detail ? `Could not save: ${detail}` : 'Could not save your integrations.'
			};
		} finally {
			saving = false;
		}
	}

	function reset() {
		void loadCreds();
		fieldErrors = {};
		clearFeedback();
	}

	/**
	 * Opens the backend's interactive link flow, then re-reads credentials once
	 * the popup closes. The previous page reported success on open, which
	 * claimed a link that may never have completed.
	 */
	function linkGitHub() {
		linkPending = true;
		const popup = window.open(apiUrl('/link'), '_blank', 'width=600,height=600');
		if (!popup) {
			linkPending = false;
			feedback = {
				kind: 'error',
				message: 'The link window was blocked. Allow popups for this app and try again.'
			};
			return;
		}
		const poll = window.setInterval(() => {
			if (!popup.closed) return;
			window.clearInterval(poll);
			linkPending = false;
			void loadCreds();
		}, 700);
	}

	onMount(() => {
		let started = false;
		const unsubscribe = user.subscribe((value) => {
			if (value.status === 'unauthenticated') {
				goto('/login');
				return;
			}
			if (value.status === 'authenticated' && !started) {
				started = true;
				void loadCreds();
			}
		});
		void initializeUser();
		return unsubscribe;
	});

	const sectionLabel =
		'text-[11px] font-medium tracking-wide text-dark-onSurfaceVariant uppercase';
	const sectionHelp = 'mt-0.5 text-[12px] text-dark-onSurfaceVariant';
	const panel = 'rounded-lg border border-border bg-dark-surfaceContainer p-4';
	const labelClass = 'text-[12px] font-medium tracking-wide text-dark-onSurfaceVariant';
	const controlClass =
		'h-9 w-full rounded-md border border-border bg-dark-surfaceContainerLowest px-3 text-sm text-dark-onSurface focus:outline-none focus:ring-2 focus:ring-ring/60';
</script>

<div class="mx-auto max-w-3xl">
	<form
		class="space-y-6"
		onsubmit={(event) => {
			event.preventDefault();
			void save();
		}}
	>
		<div class="flex items-center justify-end gap-3">
			<span class={statusClass} aria-live="polite">{statusLabel}</span>
			<Button variant="ghost" size="sm" disabled={loading || saving} onclick={reset}
				>Reload</Button
			>
			<Button
				variant="primary"
				size="sm"
				disabled={loading || !dirty}
				loading={saving}
				onclick={() => void save()}
			>
				Save changes
			</Button>
		</div>

		{#snippet notice(message: string)}
			<div
				class="border-dark-error/40 bg-dark-errorContainer/40 rounded-lg border px-4 py-3"
				role="alert"
			>
				<p class="text-dark-onSurface text-[13px]">{message}</p>
			</div>
		{/snippet}

		{#snippet selectField(
			id: string,
			label: string,
			value: string,
			options: SelectOption[],
			placeholder: string,
			onchange: (next: string) => void,
			disabled: boolean
		)}
			<div class="flex flex-col gap-1.5">
				<label class={labelClass} for={id}>{label}</label>
				<select
					{id}
					class={controlClass}
					{value}
					{disabled}
					onchange={(event) => onchange(event.currentTarget.value)}
				>
					<option value="" disabled>{placeholder}</option>
					{#each options as option (option.value)}
						<option value={option.value}>{option.label}</option>
					{/each}
				</select>
			</div>
		{/snippet}

		{#if loadError}
			{@render notice(loadError)}
		{:else if feedback?.kind === 'error'}
			{@render notice(feedback.message)}
		{/if}

		<section class="space-y-2">
			<div class="px-1">
				<h2 class={sectionLabel}>Source control</h2>
				<p class={sectionHelp}>Connect GitHub so BytePort can read your repositories.</p>
			</div>
			<div class={panel}>
				<div class="flex items-center justify-between gap-4">
					<div class="min-w-0">
						<p class="text-[13px] font-medium">GitHub account</p>
						<p class="text-dark-onSurfaceVariant text-[12px]">
							{githubLinked
								? 'A token is stored for this account.'
								: 'No token stored. Linking opens a GitHub window.'}
						</p>
					</div>
					<div class="flex shrink-0 items-center gap-3">
						<Badge tone={githubLinked ? 'primary' : 'neutral'} dot>
							{githubLinked ? 'linked' : 'not linked'}
						</Badge>
						<Button
							variant="secondary"
							size="sm"
							loading={linkPending}
							onclick={linkGitHub}
						>
							{githubLinked ? 'Relink' : 'Link GitHub'}
						</Button>
					</div>
				</div>
			</div>
		</section>

		<section class="space-y-2">
			<div class="px-1">
				<h2 class={sectionLabel}>Cloud provider</h2>
				<p class={sectionHelp}>Validated against AWS when you save, then encrypted.</p>
			</div>
			<div class="{panel} space-y-4">
				<Input
					label="AWS access key ID"
					error={fieldErrors['awsAccessKey'] ?? ''}
					value={awsAccessKey}
					autocomplete="off"
					spellcheck={false}
					oninput={(event) => {
						awsAccessKey = event.currentTarget.value;
						clearFeedback();
					}}
				/>
				<Input
					label="AWS secret access key"
					error={fieldErrors['awsSecretKey'] ?? ''}
					type="password"
					value={awsSecretKey}
					autocomplete="off"
					oninput={(event) => {
						awsSecretKey = event.currentTarget.value;
						clearFeedback();
					}}
				/>
			</div>
		</section>

		<section class="space-y-2">
			<div class="px-1">
				<h2 class={sectionLabel}>AI provider</h2>
				<p class={sectionHelp}>Generates the project templates BytePort deploys.</p>
			</div>
			<div class="{panel} space-y-4">
				{@render selectField(
					'llm-provider',
					'Provider',
					provider,
					providerOptions,
					'Select a provider',
					selectProvider,
					false
				)}
				{@render selectField(
					'llm-model',
					'Model',
					providerModels[provider]?.modal ?? '',
					modelOptions,
					'Select a model',
					(next: string) => patchProvider({ modal: next }),
					provider.length === 0 || modelOptions.length === 0
				)}
				{#if provider.length > 0 && provider !== 'local'}
					<Input
						label="API key"
						hint="Shown because the backend returns it decrypted for this account."
						type="password"
						value={providerModels[provider]?.apiKey ?? ''}
						autocomplete="off"
						oninput={(event) => patchProvider({ apiKey: event.currentTarget.value })}
					/>
				{:else if provider === 'local'}
					<p class="text-dark-onSurfaceVariant text-[12px]">
						ByteLlama models run locally, so no API key is needed.
					</p>
				{/if}
			</div>
		</section>

		<section class="space-y-2">
			<div class="px-1">
				<h2 class={sectionLabel}>Portfolio</h2>
				<p class={sectionHelp}>Where BytePort publishes the generated templates.</p>
			</div>
			<div class="{panel} space-y-4">
				<Input
					label="Endpoint URL"
					error={fieldErrors['portfolioEndpoint'] ?? ''}
					value={portfolioEndpoint}
					autocomplete="off"
					spellcheck={false}
					oninput={(event) => {
						portfolioEndpoint = event.currentTarget.value;
						clearFeedback();
					}}
				/>
				<Input
					label="API key"
					error={fieldErrors['portfolioKey'] ?? ''}
					type="password"
					value={portfolioKey}
					autocomplete="off"
					oninput={(event) => {
						portfolioKey = event.currentTarget.value;
						clearFeedback();
					}}
				/>
			</div>
		</section>
	</form>
</div>
