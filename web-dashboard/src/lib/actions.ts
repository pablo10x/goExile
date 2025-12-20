export function autofocus(node: HTMLElement) {
	node.focus();
}

export function portal(node: HTMLElement, target: HTMLElement | string = 'body') {
	let targetEl: HTMLElement | null;
	if (typeof target === 'string') {
		targetEl = document.querySelector(target);
	} else {
		targetEl = target;
	}

	if (targetEl) {
		targetEl.appendChild(node);
		node.hidden = false;
	}

	return {
		update(newTarget: HTMLElement | string) {
			if (typeof newTarget === 'string') {
				targetEl = document.querySelector(newTarget);
			} else {
				targetEl = newTarget;
			}
			if (targetEl) {
				targetEl.appendChild(node);
			}
		},
		destroy() {
			if (node.parentNode) {
				node.parentNode.removeChild(node);
			}
		}
	};
}
