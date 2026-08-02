/**
 * Roving tabindex helper for arrow-key navigable widget groups
 * (toolbars, menu bars, radio groups, lists, tablists).
 *
 * PILLAR-TAXONOMY-v2.md v2.2 §L76 (accessibility).
 *
 * WCAG-AA 2.1.1 (keyboard) + WAI-ARIA Authoring Practices.
 *
 * Usage:
 *   <script>
 *     import { rovingTabindex } from '$lib/a11y/rovingTabindex';
 *   </script>
 *   <ul use:rovingTabindex={{ orientation: 'vertical' }}>
 *     <li tabindex="-1">Item 1</li>
 *     <li tabindex="-1">Item 2</li>
 *   </ul>
 *
 * The first item gets tabindex=0 by default (the tab-stop), the rest
 * get tabindex=-1. Arrow keys move focus between items. Home/End jump
 * to first/last. Wrapping is supported.
 */
import type { Action } from 'svelte/action';

export type Orientation = 'horizontal' | 'vertical' | 'both';

export interface RovingTabindexOptions {
	orientation?: Orientation;
	wrap?: boolean;
	initialIndex?: number;
}

export const rovingTabindex: Action<HTMLElement, RovingTabindexOptions> = (node, options = {}) => {
	let opts: RovingTabindexOptions = {
		orientation: 'vertical',
		wrap: true,
		initialIndex: 0,
		...options
	};

	function getItems(): HTMLElement[] {
		return Array.from(node.querySelectorAll<HTMLElement>(':scope > *'));
	}

	function setActive(index: number): void {
		const items = getItems();
		if (items.length === 0) return;
		items.forEach((el, i) => {
			el.setAttribute('tabindex', i === index ? '0' : '-1');
		});
		items[index]?.focus();
	}

	$effect(() => {
		setActive(opts.initialIndex ?? 0);

		function onKeyDown(e: KeyboardEvent) {
			const items = getItems();
			if (items.length === 0) return;

			const currentIndex = items.indexOf(document.activeElement as HTMLElement);
			if (currentIndex === -1) return;

			const orientation = opts.orientation ?? 'vertical';
			const wrap = opts.wrap ?? true;

			let nextIndex = currentIndex;

			const isPrev =
				(orientation === 'vertical' && e.key === 'ArrowUp') ||
				(orientation === 'horizontal' && e.key === 'ArrowLeft') ||
				(orientation === 'both' && (e.key === 'ArrowUp' || e.key === 'ArrowLeft'));
			const isNext =
				(orientation === 'vertical' && e.key === 'ArrowDown') ||
				(orientation === 'horizontal' && e.key === 'ArrowRight') ||
				(orientation === 'both' && (e.key === 'ArrowDown' || e.key === 'ArrowRight'));

			if (isPrev) {
				e.preventDefault();
				nextIndex = currentIndex - 1;
				if (nextIndex < 0) nextIndex = wrap ? items.length - 1 : 0;
			} else if (isNext) {
				e.preventDefault();
				nextIndex = currentIndex + 1;
				if (nextIndex >= items.length) nextIndex = wrap ? 0 : items.length - 1;
			} else if (e.key === 'Home') {
				e.preventDefault();
				nextIndex = 0;
			} else if (e.key === 'End') {
				e.preventDefault();
				nextIndex = items.length - 1;
			}

			if (nextIndex !== currentIndex) {
				setActive(nextIndex);
			}
		}

		node.addEventListener('keydown', onKeyDown);

		return () => {
			node.removeEventListener('keydown', onKeyDown);
		};
	});

	return {
		update(next: RovingTabindexOptions) {
			opts = { orientation: 'vertical', wrap: true, initialIndex: 0, ...next };
		}
	};
};
