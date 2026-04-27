import aspectRatio from '@tailwindcss/aspect-ratio';
import containerQueries from '@tailwindcss/container-queries';
import forms from '@tailwindcss/forms';
import typography from '@tailwindcss/typography';
import { fontFamily } from 'tailwindcss/defaultTheme';
import type { Config } from 'tailwindcss';

const darkTheme = {
	primary: '#80d5cf',
	surfaceTint: '#80d5cf',
	onPrimary: '#003734',
	primaryContainer: '#00504c',
	onPrimaryContainer: '#9df1eb',
	secondary: '#83d2e3',
	onSecondary: '#00363e',
	secondaryContainer: '#004e5a',
	onSecondaryContainer: '#a2eeff',
	tertiary: '#9bcbfb',
	onTertiary: '#003353',
	tertiaryContainer: '#0e4a73',
	onTertiaryContainer: '#cee5ff',
	error: '#ffb4ab',
	onError: '#690005',
	errorContainer: '#93000a',
	onErrorContainer: '#ffdada',
	background: '#0e1514',
	onBackground: '#dde4e2',
	surface: '#101418',
	onSurface: '#e1e2e8',
	surfaceVariant: '#3f4948',
	onSurfaceVariant: '#bec9c7',
	outline: '#889391',
	outlineVariant: '#3f4948',
	shadow: '#000000',
	scrim: '#000000',
	inverseSurface: '#e1e2e8',
	inverseOnSurface: '#2e3135',
	inversePrimary: '#006a65',
	primaryFixed: '#9df1eb',
	onPrimaryFixed: '#00201e',
	primaryFixedDim: '#80d5cf',
	onPrimaryFixedVariant: '#00504c',
	secondaryFixed: '#a2eeff',
	onSecondaryFixed: '#001f25',
	secondaryFixedDim: '#83d2e3',
	onSecondaryFixedVariant: '#004e5a',
	tertiaryFixed: '#cee5ff',
	onTertiaryFixed: '#001d33',
	tertiaryFixedDim: '#9bcbfb',
	onTertiaryFixedVariant: '#0e4a73',
	surfaceDim: '#101418',
	surfaceBright: '#36393e',
	surfaceContainerLowest: '#0b0e13',
	surfaceContainerLow: '#191c20',
	surfaceContainer: '#1d2024',
	surfaceContainerHigh: '#272a2f',
	surfaceContainerHighest: '#32353a'
};

const config: Config = {
	darkMode: ['class'], // Enable class-based dark mode
	content: ['./src/**/*.{html,js,svelte,ts}'],
	safelist: ['dark'],
	theme: {
		container: {
			center: true,
			padding: '2rem',
			screens: {
				'2xl': '1400px'
			}
		},
		extend: {
			colors: {
				// Ensure shadcn colors are preserved
				border: 'hsl(var(--border) / <alpha-value>)',
				input: 'hsl(var(--input) / <alpha-value>)',
				ring: 'hsl(var(--ring) / <alpha-value>)',
				background: 'hsl(var(--background) / <alpha-value>)',
				foreground: 'hsl(var(--foreground) / <alpha-value>)',
				primary: {
					DEFAULT: 'hsl(var(--primary) / <alpha-value>)',
					foreground: 'hsl(var(--primary-foreground) / <alpha-value>)'
				},
				secondary: {
					DEFAULT: 'hsl(var(--secondary) / <alpha-value>)',
					foreground: 'hsl(var(--secondary-foreground) / <alpha-value>)'
				},
				destructive: {
					DEFAULT: 'hsl(var(--destructive) / <alpha-value>)',
					foreground: 'hsl(var(--destructive-foreground) / <alpha-value>)'
				},
				muted: {
					DEFAULT: 'hsl(var(--muted) / <alpha-value>)',
					foreground: 'hsl(var(--muted-foreground) / <alpha-value>)'
				},
				accent: {
					DEFAULT: 'hsl(var(--accent) / <alpha-value>)',
					foreground: 'hsl(var(--accent-foreground) / <alpha-value>)'
				},
				popover: {
					DEFAULT: 'hsl(var(--popover) / <alpha-value>)',
					foreground: 'hsl(var(--popover-foreground) / <alpha-value>)'
				},
				card: {
					DEFAULT: 'hsl(var(--card) / <alpha-value>)',
					foreground: 'hsl(var(--card-foreground) / <alpha-value>)'
				},
				// Preserve your prior dark theme colors under "dark"
				dark: darkTheme
			},
			borderRadius: {
				lg: 'var(--radius)',
				md: 'calc(var(--radius) - 2px)',
				sm: 'calc(var(--radius) - 4px)'
			},
			fontFamily: {
				sans: [...fontFamily.sans]
			}
		}
	},
	plugins: [typography, forms, containerQueries, aspectRatio]
};

export default config;
