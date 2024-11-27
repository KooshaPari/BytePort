import { fontFamily } from 'tailwindcss/defaultTheme';
import type { Config } from 'tailwindcss';

const config: Config = {
	darkMode: ['class'],
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
				border: '#3f4948',
				input: '#3f4948',
				ring: '#889391',
				background: '#0e1514',
				foreground: '#dde4e2',
				primary: {
					DEFAULT: '#80d5cf',
					foreground: '#003734'
				},
				secondary: {
					DEFAULT: '#83d2e3',
					foreground: '#00363e'
				},
				destructive: {
					DEFAULT: '#ffb4ab',
					foreground: '#690005'
				},
				muted: {
					DEFAULT: '#101418',
					foreground: '#bec9c7'
				},
				accent: {
					DEFAULT: '#9bcbfb',
					foreground: '#003353'
				},
				popover: {
					DEFAULT: '#0e1514',
					foreground: '#dde4e2'
				},
				card: {
					DEFAULT: '#0e1514',
					foreground: '#dde4e2'
				}
			},
			borderRadius: {
				lg: '0.5rem', // Matches var(--radius)
				md: 'calc(0.5rem - 2px)',
				sm: 'calc(0.5rem - 4px)'
			},
			fontFamily: {
				sans: [...fontFamily.sans]
			}
		}
	},
	plugins: []
};

export default config;
