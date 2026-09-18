import aspectRatio from '@tailwindcss/aspect-ratio';
import containerQueries from '@tailwindcss/container-queries';
import forms from '@tailwindcss/forms';
import tailwindScrollbar from 'tailwind-scrollbar';
import { fontFamily } from 'tailwindcss/defaultTheme';
import type { Config } from 'tailwindcss';

// BytePort dark palette.
//
// Replaces the 2023-era Material 3 (Material You) scheme: that palette used
// desaturated pastel "containers" with a muddy teal-brown cast and heavy
// rounding, which read as a mobile app rather than a desktop tool.
//
// This scale is a cool graphite neutral ramp with a single teal accent. The
// surface steps are deliberately close together (a few percent lightness each)
// so elevation reads as subtle layering instead of striped boxes. Accent roles
// are kept only where they carry meaning: primary = action/brand,
// secondary = informational, tertiary = deployment/VM state, error = failure.
//
// Values stay named `dark*` because ~200 existing `bg-dark-*` / `text-dark-*`
// class references across 15 components depend on this object; re-skinning the
// values upgrades every one of them without touching component markup.
const materialTheme = {
	// --- Accent roles ---
	primary: '#2dd4bf', // teal 400
	onPrimary: '#04211d',
	primaryContainer: '#0f3d38',
	onPrimaryContainer: '#8df0e3',

	secondary: '#60a5fa', // blue 400 — informational
	onSecondary: '#04202e',
	secondaryContainer: '#12354d',
	onSecondaryContainer: '#c7e4ff',

	tertiary: '#a78bfa', // violet 400 — deployment / VM state
	onTertiary: '#1b1436',
	tertiaryContainer: '#2f2757',
	onTertiaryContainer: '#ddd4ff',

	error: '#f87171',
	onError: '#2a0a0a',
	errorContainer: '#4a1d1d',
	onErrorContainer: '#ffd9d9',

	// --- Neutrals (cool graphite ramp) ---
	background: '#0d0f13',
	onBackground: '#e6e9ee',
	surface: '#121519',
	onSurface: '#e6e9ee',
	surfaceDim: '#0a0c0f',
	surfaceBright: '#31363e',
	surfaceVariant: '#262a31',
	onSurfaceVariant: '#9ba3af',
	outline: '#3a4048',
	outlineVariant: '#262a31',
	surfaceContainerLowest: '#0a0c0f',
	surfaceContainerLow: '#16191e',
	surfaceContainer: '#1a1e23',
	surfaceContainerHigh: '#1f2329',
	surfaceContainerHighest: '#262a31'
};

const config: Config = {
	darkMode: ['class'], // Enable dark mode support
	content: ['./src/**/*.{html,js,svelte,ts}'], // Ensure all files are scanned for Tailwind classes
	safelist: ['dark'], // Ensure dark mode classes are preserved
	theme: {
		extend: {
			colors: {
				// Map shadcn variables to your material scheme
				border: 'hsl(var(--border) / <alpha-value>)',
				input: 'hsl(var(--input) / <alpha-value>)',
				ring: 'hsl(var(--ring) / <alpha-value>)',
				background: materialTheme.background,
				foreground: materialTheme.onBackground,
				primary: {
					DEFAULT: materialTheme.primary,
					foreground: materialTheme.onPrimary
				},
				secondary: {
					DEFAULT: materialTheme.secondary,
					foreground: materialTheme.onSecondary
				},
				destructive: {
					DEFAULT: materialTheme.error,
					foreground: materialTheme.onError
				},
				muted: {
					DEFAULT: materialTheme.surfaceVariant,
					foreground: materialTheme.onSurfaceVariant
				},
				accent: {
					DEFAULT: materialTheme.secondary,
					foreground: materialTheme.onSecondary
				},
				popover: {
					DEFAULT: materialTheme.surface,
					foreground: materialTheme.onSurface
				},
				card: {
					DEFAULT: materialTheme.surface,
					foreground: materialTheme.onSurface
				},
				// Retain darkTheme for existing references (e.g., bg-dark-X)
				dark: materialTheme
			},
			borderRadius: {
				lg: 'var(--radius)',
				md: 'calc(var(--radius) - 2px)',
				sm: 'calc(var(--radius) - 4px)'
			},
			fontFamily: {
				sans: [...fontFamily.sans]
			}
		},
		container: {
			center: true,
			padding: '2rem',
			screens: {
				'2xl': '1400px'
			}
		}
	},
	plugins: [forms, containerQueries, aspectRatio, tailwindScrollbar]
};

export default config;
