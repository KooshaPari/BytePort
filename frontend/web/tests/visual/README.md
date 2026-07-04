# Visual regression policy

This directory contains Playwright snapshot tests for the BytePort UI surface.
Pillar: **L60 Visual regression** (PILLAR-TAXONOMY-v2.md v2.2).

## Scope

Snapshots cover brand-visible surfaces that end users see at boot or in empty states:

- `splash` — desktop launch surface (Tauri splash, branded)
- `empty-state` — no-data / no-results / error / mascot variants
- `mascot` — mascot idle-frame (mid-breath)

When adding a new brand surface (a new splash variant, a new illustration,
a hero image), add a snapshot in `snapshots.spec.ts`. Treat it like a test:
the new surface is unowned until it has a snapshot.

## Running locally

```bash
cd frontend/web
npx playwright install --with-deps chromium   # one-time
npm run build                                  # build SvelteKit
PLAYWRIGHT=1 npx playwright test --config=playwright.visual.config.ts
```

To update snapshots after an intentional visual change:

```bash
PLAYWRIGHT=1 npx playwright test --config=playwright.visual.config.ts --update-snapshots
```

To run a single project (e.g. just mobile-light):

```bash
PLAYWRIGHT=1 npx playwright test --config=playwright.visual.config.ts --project=mobile-light
```

## Anti-flake rules

- All snapshots use `animations: 'disabled'` — CSS transitions are paused
  and SMIL animations are not in motion during snapshot capture.
- For animated SVG surfaces (the mascot), we wait one breath cycle
  (`page.waitForTimeout(2100)`) before snapshot so the SMIL state is
  mid-animation, not frame-0.
- `maxDiffPixelRatio: 0.01` (1%) — anything beyond 1% pixel diff fails the
  test. Lower it to 0.005 once the suite has stabilized.

## CI

The `visual-regression.yml` workflow runs on every PR touching the
relevant paths. On failure, the workflow:

1. Uploads the actual/diff PNGs as a `visual-regression-diffs` artifact
   (retained 7 days).
2. Comments on the PR with a remediation playbook.

The CI matrix is desktop-light / desktop-dark / mobile-light / mobile-dark.
A mobile-dark snapshot is intentional and will be added if a reviewer
requests it.

## Snapshot retention

Snapshots are committed to the repo (`tests/visual/__snapshots__/`) so
diffs are reviewable in PR diffs. Do not gitignore them.

When deleting a snapshot, delete the entire `.png` file in the
`__snapshots__/` directory AND remove the matching `test()` block in
`snapshots.spec.ts`.

## Adding a new snapshot

1. Add a `<svelte:head>` route under `frontend/web/src/routes/__visual/<name>/+page.svelte`
   that renders only the surface you want to snapshot.
2. Add a `test()` block in `snapshots.spec.ts` that navigates to it.
3. Run `npx playwright test --update-snapshots` to capture the initial baseline.
4. Commit both the spec and the snapshot.