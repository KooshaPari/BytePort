import { type ClassValue, clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';
import { cubicOut } from 'svelte/easing';
import type { TransitionConfig } from 'svelte/transition';

export function cn(...inputs: ClassValue[]) {
	return twMerge(clsx(inputs));
}

type FlyAndScaleParams = {
	y?: number;
	x?: number;
	start?: number;
	duration?: number;
};

export const flyAndScale = (
	node: Element,
	params: FlyAndScaleParams = { y: -8, x: 0, start: 0.95, duration: 150 }
): TransitionConfig => {
	const style = getComputedStyle(node);
	const transform = style.transform === 'none' ? '' : style.transform;

	const scaleConversion = (valueA: number, scaleA: [number, number], scaleB: [number, number]) => {
		const [minA, maxA] = scaleA;
		const [minB, maxB] = scaleB;

		const percentage = (valueA - minA) / (maxA - minA);
		const valueB = percentage * (maxB - minB) + minB;

		return valueB;
	};

	const styleToString = (style: Record<string, number | string | undefined>): string => {
		return Object.keys(style).reduce((str, key) => {
			if (style[key] === undefined) return str;
			return str + `${key}:${style[key]};`;
		}, '');
	};

	return {
		duration: params.duration ?? 200,
		delay: 0,
		css: (t) => {
			const y = scaleConversion(t, [0, 1], [params.y ?? 5, 0]);
			const x = scaleConversion(t, [0, 1], [params.x ?? 0, 0]);
			const scale = scaleConversion(t, [0, 1], [params.start ?? 0.95, 1]);

			return styleToString({
				transform: `${transform} translate3d(${x}px, ${y}px, 0) scale(${scale})`,
				opacity: t
			});
		},
		easing: cubicOut
	};
};
export interface Repository {
	id: number; // Repository ID
	node_id: string; // Node ID
	name: string; // Repository name
	full_name: string; // Full repository name (e.g., "owner/repo")
	private: boolean; // Whether the repository is private
	owner: {
		login: string; // Owner's username
		id: number; // Owner's ID
		avatar_url: string; // Owner's avatar URL
		html_url: string; // Owner's HTML URL
	};
	html_url: string; // URL to the repository on GitHub
	description: string | null; // Repository description
	fork: boolean; // Whether the repository is a fork
	created_at: string; // ISO string of creation date
	updated_at: string; // ISO string of last update
	pushed_at: string; // ISO string of last push
	language: string | null; // Main programming language
	size: number; // Repository size in KB
	visibility: string; // Visibility (e.g., "public", "private")
	permissions: {
		admin: boolean; // Admin access
		maintain: boolean; // Maintain access
		push: boolean; // Push access
		triage: boolean; // Triage access
		pull: boolean; // Pull access
	};
}
