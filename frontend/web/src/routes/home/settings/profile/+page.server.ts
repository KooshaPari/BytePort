import type { PageServerLoad } from './$types.js';

/**
 * The profile form is driven entirely on the client: the app runs with
 * `ssr = false` (src/routes/+layout.ts), so a server load never has a session
 * to read. The previous version subscribed to the client-side user store at
 * module scope on the server, where it is always `{ status: 'pending' }`, and
 * returned a sveltekit-superforms payload that the page never used.
 *
 * Returning an empty object keeps the server surface explicit instead of
 * pretending to have data it cannot have. The page seeds itself from the
 * session store and saves through PUT /user/:id/creds.
 */
export const load: PageServerLoad = async () => ({});
