<script lang="ts">
	/**
	 * First-time setup.
	 *
	 * Rewritten from the 2023 view, which had four real defects:
	 *
	 *  1. It called `platform()` from `@tauri-apps/plugin-os` unguarded. That
	 *     plugin is not registered in the Rust shell, so the call threw while
	 *     dereferencing `window.__TAURI_OS_PLUGIN_INTERNALS__`. `getBaseUrl` was
	 *     async, so the rejection was swallowed by `onMount` and `baseUrl` stayed
	 *     `undefined`; the final `fetch(`${baseUrl}/link`)` then hit the literal
	 *     URL "undefined/link". Base URLs now come from `$lib/api`, which never
	 *     throws.
	 *  2. Every `<label for="...">` pointed at an id no element had, because the
	 *     inputs only set `name`. No field was actually labelled.
	 *  3. Advancement was driven by `document.querySelector` against
	 *     `#form-stage-N`, with the stage counter incremented in two places. The
	 *     forms had no submit handler, so Enter did nothing.
	 *  4. The icon-only buttons carried no accessible name.
	 */
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { ApiError, apiFetch, apiUrl, getApiBaseUrl } from '$lib/api';
	import { initializeUser, user } from '../../stores/user';
	import Button from '$lib/components/ui/Button.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Input from '$lib/components/ui/Input.svelte';

	/**
	 * The wire shape `POST /link` actually binds, declared against the Go
	 * structs rather than the shared `UserLink` type.
	 *
	 * `models.AIProvider` tags its key field `api_key` (snake_case) and
	 * `ValidateLink` looks the provider up under the literal key "openai"
	 * (lowercase). `stores/user.ts` models both as `apiKey` / `openAI`, so
	 * sending that type's casing meant the backend validated an EMPTY OpenAI
	 * key and every attempt failed with "Failed to validate OAI credentials".
	 * Go map keys are case-sensitive, so the casing here is load-bearing.
	 */
	interface LinkPayload {
		awsCreds: { accessKeyId: string; secretAccessKey: string };
		llmConfig: {
			provider: string;
			providers: Record<string, { modal: string; api_key: string }>;
		};
		portfolio: { rootEndpoint: string; apiKey: string };
	}

	const TOTAL_STEPS = 4;
	const STEPS = [
		{ title: 'AWS credentials', description: 'BytePort uses these to provision instances.' },
		{ title: 'OpenAI credentials', description: 'Used to generate your portfolio content.' },
		{ title: 'Portfolio endpoint', description: 'Where BytePort publishes your portfolio.' },
		{ title: 'Connect GitHub', description: 'Authorise BytePort to read your repositories.' }
	] as const;
	const STEP_INDICES = Array.from({ length: TOTAL_STEPS }, (_, i) => i + 1);

	type AuthState = 'checking' | 'authenticated' | 'unauthenticated';

	let authState = $state<AuthState>('checking');
	let step = $state(1);
	let linked = $state(false);
	let popupBlocked = $state(false);
	let submitting = $state(false);
	let formError = $state('');
	let fieldErrors = $state<Record<string, string>>({});

	let awsAccessKeyId = $state('');
	let awsSecretAccessKey = $state('');
	let openAiApiKey = $state('');
	let portfolioRootEndpoint = $state('');
	let portfolioApiKey = $state('');

	const currentStep = $derived(STEPS[step - 1]);

	onMount(() => {
		// `initializeUser` reads the session cookie from the backend and drives
		// the shared store; the subscription reacts to the result.
		void initializeUser();

		const unsubscribe = user.subscribe((value) => {
			if (value.status === 'pending') return;

			if (value.status !== 'authenticated') {
				authState = 'unauthenticated';
				void goto('/login');
				return;
			}
			authState = 'authenticated';
		});

		return unsubscribe;
	});

	function validateStep(): boolean {
		const next: Record<string, string> = {};

		if (step === 1) {
			if (!awsAccessKeyId.trim()) next.awsAccessKeyId = 'Enter your AWS access key ID.';
			if (!awsSecretAccessKey.trim())
				next.awsSecretAccessKey = 'Enter your AWS secret access key.';
		} else if (step === 2) {
			// No prefix rule: the backend calls AWS/OpenAI directly to validate,
			// and guessing key formats client-side only risks false rejections.
			if (!openAiApiKey.trim()) next.openAiApiKey = 'Enter your OpenAI API key.';
		} else if (step === 3) {
			const endpoint = portfolioRootEndpoint.trim();
			if (!endpoint) next.portfolioRootEndpoint = 'Enter your portfolio root endpoint.';
			else {
				try {
					const parsed = new URL(endpoint);
					if (parsed.protocol !== 'http:' && parsed.protocol !== 'https:') {
						next.portfolioRootEndpoint = 'Use an http or https URL.';
					}
				} catch {
					next.portfolioRootEndpoint =
						'Enter a full URL, for example https://example.com.';
				}
			}
			if (!portfolioApiKey.trim()) next.portfolioApiKey = 'Enter your portfolio API key.';
		}

		fieldErrors = next;
		return Object.keys(next).length === 0;
	}

	function describeFailure(err: unknown, what: string): string {
		if (err instanceof ApiError) {
			const body = err.body as { message?: string; error?: string; details?: string } | null;
			const text = body?.message ?? body?.error ?? '';
			// The backend's `details` field carries the actionable part, e.g. the
			// reason AWS rejected the key.
			if (text) return `${text}${body?.details ? `: ${body.details}` : ''}`;
			if (err.status === 401) return 'Your session expired. Sign in again.';
			return `${what} failed (HTTP ${err.status}).`;
		}

		if (err instanceof DOMException && err.name === 'AbortError') {
			return 'The backend did not respond in time.';
		}

		return `Could not reach the BytePort backend at ${getApiBaseUrl()}. Check that it is running.`;
	}

	function onPrimary() {
		formError = '';
		if (submitting) return;

		if (step < TOTAL_STEPS) {
			if (validateStep()) step += 1;
			return;
		}
		void connectGithub();
	}

	async function onSubmit(event: SubmitEvent) {
		event.preventDefault();
		onPrimary();
	}

	function goBack() {
		formError = '';
		fieldErrors = {};
		if (step > 1) step -= 1;
	}

	/** Opens the GitHub authorisation window. Returns false if it was blocked. */
	function openLinkWindow(): boolean {
		const popup = window.open(apiUrl('/link'), 'byteport-github-link', 'width=620,height=720');
		return Boolean(popup);
	}

	async function connectGithub() {
		submitting = true;
		formError = '';
		popupBlocked = false;

		const payload: LinkPayload = {
			awsCreds: {
				accessKeyId: awsAccessKeyId.trim(),
				secretAccessKey: awsSecretAccessKey.trim()
			},
			llmConfig: {
				provider: 'openai',
				providers: { openai: { modal: 'gpt-4o', api_key: openAiApiKey.trim() } }
			},
			portfolio: {
				rootEndpoint: portfolioRootEndpoint.trim(),
				apiKey: portfolioApiKey.trim()
			}
		};

		try {
			// Validates and encrypts every credential server-side, so failures
			// (a bad AWS key, for example) are reported before the user is sent
			// to GitHub.
			await apiFetch<{ message: string }>('/link', {
				method: 'POST',
				body: JSON.stringify(payload)
			});
		} catch (err) {
			formError = describeFailure(err, 'Credential validation');
			submitting = false;
			return;
		}

		// The popup must be opened from the click that started this handler, and
		// that click has already been consumed by the await above, so browsers
		// may block it. Report that instead of failing silently.
		popupBlocked = !openLinkWindow();
		linked = true;
		submitting = false;
	}
</script>

<main class="bg-dark-background h-screen w-full overflow-y-auto">
	<div class="flex min-h-full items-center justify-center px-4 py-10">
		<div class="w-full max-w-[420px]">
			<div class="mb-6 flex flex-col items-center gap-3">
				<div
					class="border-border bg-dark-surfaceContainer flex h-10 w-10 items-center justify-center rounded-[10px] border"
					aria-hidden="true"
				>
					<span class="text-dark-primary text-[13px] font-semibold tracking-tight"
						>BP</span
					>
				</div>
				<div class="text-center">
					<p
						class="text-dark-onSurfaceVariant text-[11px] font-medium tracking-wide uppercase"
					>
						First time setup
					</p>
					<h1 class="text-dark-onSurface mt-1 text-[15px] font-semibold tracking-tight">
						{linked ? 'Finish connecting GitHub' : currentStep.title}
					</h1>
					<p class="text-dark-onSurfaceVariant mt-1 text-[13px] leading-snug">
						{linked
							? 'Complete the authorisation in the GitHub window, then return here.'
							: currentStep.description}
					</p>
				</div>
			</div>

			{#if authState === 'checking'}
				<Card padding="lg">
					<p class="text-dark-onSurfaceVariant text-[13px]" role="status">
						Checking your session.
					</p>
				</Card>
			{:else if authState === 'unauthenticated'}
				<Card padding="lg">
					<p class="text-dark-onSurfaceVariant text-[13px]" role="status">
						Your session has ended. Redirecting to sign in.
					</p>
				</Card>
			{:else if linked}
				<Card padding="lg">
					<div class="flex flex-col gap-4">
						{#if popupBlocked}
							<p
								role="alert"
								class="border-dark-tertiary/40 bg-dark-tertiaryContainer/40 text-dark-onTertiaryContainer rounded-md border px-3 py-2 text-[12px] leading-snug"
							>
								Your browser blocked the GitHub window. Open it with the link below.
							</p>
						{/if}

						<a
							class="text-dark-primary text-[13px] underline underline-offset-4 hover:brightness-110"
							href={apiUrl('/link')}
							target="_blank"
							rel="noreferrer"
						>
							Open the GitHub authorisation page
						</a>

						<Button variant="primary" class="w-full" onclick={() => void goto('/home')}>
							Go to dashboard
						</Button>
					</div>
				</Card>
			{:else}
				<Card padding="lg">
					<div class="mb-4 flex items-center gap-3">
						<span
							class="text-dark-onSurfaceVariant text-[11px] font-medium tracking-wide uppercase"
						>
							Step {step} of {TOTAL_STEPS}
						</span>
						<span class="flex flex-1 gap-1" aria-hidden="true">
							{#each STEP_INDICES as index (index)}
								<span
									class="h-0.5 flex-1 rounded-full {index <= step
										? 'bg-dark-primary'
										: 'bg-dark-surfaceVariant'}"
								></span>
							{/each}
						</span>
					</div>

					<form class="flex flex-col gap-4" onsubmit={onSubmit} novalidate>
						{#if step === 1}
							<Input
								label="AWS access key ID"
								name="accessKeyId"
								type="text"
								autocomplete="off"
								autocapitalize="none"
								spellcheck={false}
								placeholder="AKIA..."
								value={awsAccessKeyId}
								oninput={(e) => (awsAccessKeyId = e.currentTarget.value)}
								error={fieldErrors.awsAccessKeyId ?? ''}
								disabled={submitting}
								required
							/>
							<Input
								label="AWS secret access key"
								name="secretAccessKey"
								type="password"
								autocomplete="off"
								placeholder="Your secret access key"
								hint="Stored encrypted once validated."
								value={awsSecretAccessKey}
								oninput={(e) => (awsSecretAccessKey = e.currentTarget.value)}
								error={fieldErrors.awsSecretAccessKey ?? ''}
								disabled={submitting}
								required
							/>
						{:else if step === 2}
							<Input
								label="OpenAI API key"
								name="apiKey"
								type="password"
								autocomplete="off"
								autocapitalize="none"
								spellcheck={false}
								placeholder="sk-..."
								hint="Validated against OpenAI before it is saved."
								value={openAiApiKey}
								oninput={(e) => (openAiApiKey = e.currentTarget.value)}
								error={fieldErrors.openAiApiKey ?? ''}
								disabled={submitting}
								required
							/>
						{:else if step === 3}
							<Input
								label="Portfolio root endpoint"
								name="rootEndpoint"
								type="url"
								autocomplete="off"
								autocapitalize="none"
								spellcheck={false}
								placeholder="https://example.com"
								value={portfolioRootEndpoint}
								oninput={(e) => (portfolioRootEndpoint = e.currentTarget.value)}
								error={fieldErrors.portfolioRootEndpoint ?? ''}
								disabled={submitting}
								required
							/>
							<Input
								label="Portfolio API key"
								name="portfolioApiKey"
								type="password"
								autocomplete="off"
								placeholder="Your portfolio API key"
								value={portfolioApiKey}
								oninput={(e) => (portfolioApiKey = e.currentTarget.value)}
								error={fieldErrors.portfolioApiKey ?? ''}
								disabled={submitting}
								required
							/>
						{:else}
							<p class="text-dark-onSurfaceVariant text-[13px] leading-snug">
								Your credentials are validated and encrypted before GitHub is
								contacted. Nothing is sent to GitHub until you approve it in the
								next window.
							</p>
						{/if}

						{#if formError}
							<p
								role="alert"
								class="border-dark-error/40 bg-dark-errorContainer/40 text-dark-onErrorContainer rounded-md border px-3 py-2 text-[12px] leading-snug"
							>
								{formError}
							</p>
						{/if}

						<div class="flex items-center gap-2">
							{#if step > 1}
								<Button variant="ghost" disabled={submitting} onclick={goBack}
									>Back</Button
								>
							{/if}
							<Button
								type="submit"
								variant="primary"
								loading={submitting}
								class="flex-1"
							>
								{step < TOTAL_STEPS ? 'Continue' : 'Validate and connect GitHub'}
							</Button>
						</div>
					</form>
				</Card>
			{/if}

			{#if authState === 'authenticated' && !linked}
				<p class="text-dark-onSurfaceVariant mt-4 text-center text-[13px]">
					<a
						href="/home"
						class="text-dark-primary underline-offset-4 hover:underline focus-visible:underline"
					>
						Skip for now
					</a>
				</p>
			{/if}
		</div>
	</div>
</main>
