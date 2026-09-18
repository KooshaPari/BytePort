import type { PageServerLoad } from './$types.js';

/**
 * The integrations form is driven entirely on the client: the app runs with
 * `ssr = false` (src/routes/+layout.ts), so a server load never has a session
 * cookie and cannot read the user's credentials. The previous version tried
 * anyway, from a module-scope subscription to the client user store that is
 * always empty on the server, and returned a sveltekit-superforms payload the
 * page never read.
 *
 * The page now loads credentials itself from GET /user/:id/creds and saves
 * through POST /link, so this load only has to exist to keep the route's
 * server surface explicit.
 */
export const load: PageServerLoad = async () => ({});
