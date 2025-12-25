/**
 * Simple portal action to move an element to the body root.
 * This ensures modals escape parent stacking contexts.
 */
export function portal(node: HTMLElement) {
	document.body.appendChild(node);
	
	return {
		destroy() {
			if (node.parentNode) {
				node.parentNode.removeChild(node);
			}
		}
	};
}
