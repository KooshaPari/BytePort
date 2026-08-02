# BytePort Internationalization (i18n)

BytePort ships with **6 locales** out of the box, selected via the LocaleSwitcher in the
header. The implementation is **zero-dependency** — no `svelte-i18n`, no `paraglide-js`,
no bundle transform. ~80 lines of TS in `src/lib/i18n/index.ts` and 6 JSON locale files.

## Supported locales

| Code | Language | Native   | Flag |
| ---- | -------- | -------- | ---- |
| `en` | English  | English  | 🇬🇧   |
| `es` | Spanish  | Español  | 🇪🇸   |
| `fr` | French   | Français | 🇫🇷   |
| `de` | German   | Deutsch  | 🇩🇪   |
| `ja` | Japanese | 日本語   | 🇯🇵   |
| `zh` | Chinese  | 中文     | 🇨🇳   |

## How it works

1. `src/lib/i18n/index.ts` exports a Svelte `writable<Locale>` store and a derived
   `t` function store. `$t('key.path')` reads the current locale and walks the JSON tree.
2. On the **browser**, the store hydrates from `localStorage['byteport-locale']`,
   then falls back to `navigator.language`, then `'en'`. On **SSR**, it always starts at
   `'en'`.
3. Missing keys: the active locale is tried first, then English, then the literal key
   (so missing translations are visible during dev — `console.warn` is emitted under
   `import.meta.env.DEV`).
4. Interpolation: `$t('projects.lastDeployed', { when: '3 minutes ago' })` substitutes
   `{when}` placeholders. Missing params stay literal so bugs surface.
5. Date / number formatting: `formatDate(d)` and `formatNumber(n)` route through
   `Intl.DateTimeFormat` / `Intl.NumberFormat` with the active locale.

## Adding a new locale

```bash
# 1. Copy an existing locale file as a starting point
cp frontend/web/src/lib/i18n/locales/en.json \
   frontend/web/src/lib/i18n/locales/pt.json

# 2. Translate every value

# 3. Register it in src/lib/i18n/index.ts
#    - add to LOCALES
#    - add to LOCALE_LABELS
#    - add to LOCALE_FLAGS (Unicode flag emoji)
#    - import the JSON + add to the `messages` map

# 4. Add the locale code to the LocaleSwitcher (already covered by LOCALES)
```

## Translation keys

The full key tree is documented inline in `src/lib/i18n/locales/en.json`. Every key
**must** exist in `en.json` (the fallback) — keys in other locales are optional but
strongly encouraged.

Top-level groups:

- `nav.*` — header navigation
- `common.*` — buttons, status, generic actions
- `auth.*` — login / signup flow
- `home.*` — dashboard widgets
- `projects.*` / `instances.*` / `monitor.*` — domain entities
- `settings.*` — preferences
- `errors.*` — error states (HTTP + generic)
- `emptyStates.*` — empty/zero-state illustrations
- `integrations.*` — credentials forms
- `splash.*` / `footer.*` — chrome

## Adding a new translation key

```ts
// In any .svelte component
import { t } from '$lib/i18n';

const label = $t('settings.saved'); // "Settings saved"
```

```jsonc
// src/lib/i18n/locales/en.json
{
	"settings": { "saved": "Settings saved" }
}
```

If the key is missing in the active locale, the English value is used. If missing in
English too, the key itself is rendered (and a console warning fires in dev).

## Why custom?

| Library            | Bundle cost             | Build transform | Notes                                         |
| ------------------ | ----------------------- | --------------- | --------------------------------------------- |
| `svelte-i18n` 0.21 | ~9 kB gz                | yes             | Mature but pulls `svelte/store` deeply        |
| `paraglide-js` 2.x | ~5 kB gz + compile step | yes (mandatory) | Great DX, requires compile                    |
| **BytePort i18n**  | **0 kB** (just JSON)    | no              | Full control over persistence + interpolation |

## CI lint

The i18n keys are validated by `npm run lint:i18n` (TODO: add). The rule is:
every key in `en.json` must exist in every other locale file (warning only,
not error — partial translations are allowed during rollout).

## Accessibility

The `LocaleSwitcher` is fully keyboard-navigable: `Tab` to focus, `Enter` / `Space`
to open, arrow keys to navigate (browser-native via `role="listbox"`), `Escape` to
close. `aria-haspopup`, `aria-expanded`, `aria-selected`, and `aria-label` are all
set per WAI-ARIA Authoring Practices.
