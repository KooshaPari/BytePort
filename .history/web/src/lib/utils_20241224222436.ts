import { type ClassValue, clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';
import { cubicOut } from 'svelte/easing';
import type { TransitionConfig } from 'svelte/transition';
import type { User } from '../stores/user';
import type { Repository } from './git';

import { platform } from '@tauri-apps/plugin-os';
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

export interface Project {
	UUID: string;
	User: User | null;
	Name: string;
	Repository: Repository | null;
	Description: string;
	LastUpdated: string;
	Type: string;
	Platform: string;
	NVMS: NVMS;
	ReadMe: string;
	AccessURL: string;
	// map string instance
	Deployments: Record<string, Instance>;
}

export interface Instance {
	UUID: string;
	Name: string;
	Status: string;
	User: User | null;
	Project: Project | null;
	Resources: Resource[];
	OS: string;
	RootProjectUUID: string;
	LastUpdated: string;
}
export interface NVMS {
	Name: string;
	Description: string;
	Services: Service[];
}
export interface Service {
	Name: string;
	Path: string;
	Port: number;
	Build: string[];
	Env: Record<string, string>;
}
export interface Resource {
	ID: string;
	Type: string;
	Name: string;
	ARN: string;
	Status: string;
	Region: string;
	Tags: Record<string, string>;
	Properties: Record<string, unknown>;
	Associates: ResourceAssociation[];
	Service: string;
}
export interface ResourceAssociation {
	ResourceID: string;
	Type: string;
	Role: string;
}
const getBaseUrl = async () => {
	if ((window as any).__TAURI_INTERNALS__) {
		const currentPlatform: string = platform();
		console.log(currentPlatform);
		switch (currentPlatform) {
			case 'android':
				return 'http://10.0.2.2:8081';
			case 'windows':
				return 'http://localhost:8081';
			default:
				return 'http://localhost:8081';
		}
	} else {
		return 'http://localhost:8081';
	}
};
export async function populateLists() {
	let projects = [];
	const baseUrl = await getBaseUrl();
	let response = await fetch(`${baseUrl}/projects`, {
		method: 'GET',
		headers: {
			'Content-Type': 'application/json'
		},
		credentials: 'include'
	});

	if (response.ok) {
		const data = await response.json();
		projects = data;
		console.log('Projects:', projects);
		console.log('All Response Headers:');
		for (const [key, value] of response.headers.entries()) {
			console.log(`${key}: ${value}`);
		}
		console.log('Response: ', response);
	} else {
		const errorData = await response.json();
		console.error('Error:', errorData);
	}
	response = await fetch(`${baseUrl}/instances`, {
		method: 'GET',
		headers: {
			'Content-Type': 'application/json'
		},

		credentials: 'include'
	});
	if (response.ok) {
		const data = await response.json();
		instances = data;
		console.log('Instances:', instances);
		console.log('All Response Headers:');
		for (const [key, value] of response.headers.entries()) {
			console.log(`${key}: ${value}`);
		}
		console.log('Response: ', response);
	} else {
		const errorData = await response.json();
		console.error('Error:', errorData);
	}
	return projects
}
