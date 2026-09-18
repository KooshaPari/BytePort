// Tailwind is compiled by `@tailwindcss/vite` (see vite.config.ts), not PostCSS.
// Leaving `@tailwindcss/postcss` here would run Tailwind a second time over
// already-expanded CSS. autoprefixer stays: it post-processes the utilities the
// Vite plugin emits, which is the order we want.
export default {
	plugins: {
		autoprefixer: {}
	}
};
