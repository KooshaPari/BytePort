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
interface Repository {
	id: number; // Unique identifier for the repository
	name: string; // Repository name
	fullName: string; // Full name, including the owner (e.g., "owner/repo")
	private: boolean; // Indicates if the repository is private
	owner: {
		login: string; // Owner's username
		avatarUrl: string; // Owner's avatar URL
	};
	htmlUrl: string; // URL to the repository on GitHub
	description: string | null; // Optional repository description
	fork: boolean; // Indicates if the repository is a fork
	createdAt: string; // ISO string of when the repository was created
	updatedAt: string; // ISO string of when the repository was last updated
	pushedAt: string; // ISO string of when the repository was last pushed
	language: string | null; // Main language used in the repository
	size: number; // Size of the repository (in KB)
	visibility: 'public' | 'private' | 'internal'; // Repository visibility
	permissions: {
		admin: boolean; // Admin access
		push: boolean; // Push access
		pull: boolean; // Pull access
	};
	linkedTo: string; // UUID of the user who linked this repository
}
