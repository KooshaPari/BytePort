# BytePort current implementation mapping — slice 01: portfolio integration

Status: **SOURCE-MAPPED / JOURNEY OPEN**
Source snapshot: 0232cca16fedb7963a8c6f556dc5eee5c8c1674e
Date: 2026-09-30.

This slice maps the direct-user-intent-backed deployment→portfolio integration contract against the frozen implementation. It does not infer behavior from J3 route names or historical docs.

## Mature contract subject

Recovered user intent requires BytePort to turn deployed project state into a portfolio project/list representation, integrate with a configured portfolio endpoint/API, use generated project content, and retain a programmatic Git/server-side fallback when API publication is unavailable.

Slickport is a historical example backend, not core identity.

## Reachability map

| Contract step | Frozen implementation surface | Reachability / evidence | Disposition |
|---|---|---|---|
| Collect portfolio root endpoint + API key during setup | frontend/web/src/routes/fts/+page.svelte step 3, POST /link | Current UI source; explicit copy says “Where BytePort publishes your portfolio.” | **IMPLEMENTED SURFACE** |
| Edit portfolio integration later | frontend/web/src/routes/home/settings/integrations/+page.svelte | Current UI loads/saves portfolio endpoint/key through credential routes. | **IMPLEMENTED SURFACE** |
| Persist portfolio endpoint/key | models.User.Portfolio / models.Portfolio | Current model fields embedded in user record. | **IMPLEMENTED** |
| Validate portfolio API credentials | routes.ValidateLink → lib.ValidatePortfolioAPI | Current protected setup route validates before encryption/save. | **IMPLEMENTED** |
| Prevent obvious portfolio validation SSRF | portfolioValidationURL + allowlist + DNS/IP + redirect validation | Current source constrains scheme/host/resolved addresses. | **IMPLEMENTED SECURITY SUBSTRATE** |
| Encrypt/decrypt endpoint/key | routes/secrets.go | Current helper covers portfolio endpoint + API key. | **IMPLEMENTED** |
| Portfolio API health/contract probe | GET configured-root/byteport with Bearer key | ValidatePortfolioAPI proves configured endpoint responds 200 only. | **PARTIAL / CREDENTIAL VALIDATION ONLY** |
| Derive PortfolioProjection from Project + SourceSnapshot + RealizedDeployment + Observation | No mounted implementation found | Searches find requirements/docs/credentials but no projection generator tied to deployment evidence. | **MISSING** |
| Fetch list/project templates/components from configured portfolio API | No current implementation found | Direct-user intent; validation only GETs /byteport. | **MISSING** |
| Generate structured project description/media metadata with provenance | No current backend projectization path found | USER_JOURNEYS names backend/byteport/lib/llm.go, absent from frozen tree. | **MISSING / DOCUMENTATION FICTION** |
| Publish project-list representation through portfolio API | No current implementation found | No publisher POST found outside BytePort's own /link integration save. | **MISSING** |
| Publish full-project representation through portfolio API | No current implementation found | Same. | **MISSING** |
| Git/server-side publication fallback | No current portfolio-publisher Git fallback found | Git auth/listing is not publication. | **MISSING** |
| Public project route /p/:projectSlug | Named in USER_JOURNEYS; absent from current route-tree search | Journey prose is not reachability evidence. | **MISSING / DOCUMENTATION FICTION** |
| Interactive live project widget | No implementation found matching mature journey | Docs only. | **MISSING** |
| Independent deployment vs publication status | No first-class PortfolioPublication state found | No publication operation currently exists. | **MISSING DOMAIN MODEL** |

## J3 correction

USER_JOURNEYS J3 remains useful as intended-behavior evidence, but its Code Path column is not current-state evidence. The named public route, LLM module, generated-template pipeline and widget path are absent. The journey is therefore not merely untested; it is presently unclosed.

## Security/provenance consequences

The existing credential substrate can be reused by a future PortfolioPublisher, but publisher secrets must remain in secret execution context rather than PortfolioProjection payloads. Generated copy, verified endpoint/status facts, and publication receipts require separate authority/provenance classes.

## Current journey shape

setup credentials ✅ → validate endpoint ✅ → deployment evidence partial/blocked → projection ❌ → publication ❌ → public/render result ❌

Portfolio cannot contribute to mature journey closure yet.

## Next implementation-mapping work

1. inspect historical/deleted paths for a real template publisher worth recovering;
2. map Git integration for publication fallback without conflating source-repo access with publication-repo access;
3. define PortfolioProjection and PortfolioPublication operation schemas;
4. build a provider-independent static/API publisher oracle before selecting a renderer/backend;
5. add privacy/redaction and stale-endpoint adversarial fixtures.

No production implementation is authorized by this mapping alone.
