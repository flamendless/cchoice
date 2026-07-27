(function() {
	const PRELOAD_SECTIONS = 7;
	const DEFAULT_SECTION_HEIGHT = 280;
	const NEAR_VIEWPORT_SELECTOR = '[hx-trigger*="nearViewport"], [data-hx-trigger*="nearViewport"]';

	let observer = null;
	const observed = new Set();
	let lastRootMargin = null;

	function estimateSectionHeight() {
		const sections = document.querySelectorAll("[data-category-product-section]");
		if (sections.length === 0) {
			return DEFAULT_SECTION_HEIGHT;
		}

		let total = 0;
		let count = 0;
		sections.forEach(function(el) {
			if (el.offsetHeight > 0) {
				total += el.offsetHeight;
				count++;
			}
		});

		return count > 0 ? Math.ceil(total / count) : DEFAULT_SECTION_HEIGHT;
	}

	function getRootMargin() {
		const px = PRELOAD_SECTIONS * estimateSectionHeight();
		return "0px 0px " + px + "px 0px";
	}

	function eagerLoadImages(container) {
		if (!container) {
			return;
		}
		container.querySelectorAll('img[loading="lazy"]').forEach(function(img) {
			img.loading = "eager";
		});
	}

	function createObserver() {
		const rootMargin = getRootMargin();
		if (observer && lastRootMargin === rootMargin) {
			return observer;
		}

		if (observer) {
			observer.disconnect();
			observed.clear();
		}

		lastRootMargin = rootMargin;
		observer = new IntersectionObserver(function(entries) {
			entries.forEach(function(entry) {
				if (entry.isIntersecting) {
					htmx.trigger(entry.target, "nearViewport");
				}
			});
		}, { rootMargin: rootMargin, threshold: 0 });

		return observer;
	}

	function observeElement(el) {
		if (!(el instanceof Element) || !document.contains(el)) {
			return;
		}

		const trigger = el.getAttribute("hx-trigger") || el.getAttribute("data-hx-trigger") || "";
		if (!trigger.includes("nearViewport")) {
			return;
		}
		if (observed.has(el)) {
			return;
		}

		observed.add(el);
		createObserver().observe(el);
	}

	function observeNearViewportElements(root) {
		if (root instanceof Element) {
			observeElement(root);
		}

		const scope = root && root.querySelectorAll ? root : document;
		scope.querySelectorAll(NEAR_VIEWPORT_SELECTOR).forEach(observeElement);
	}

	function refreshObserver() {
		if (observer) {
			observer.disconnect();
			observed.clear();
			observer = null;
			lastRootMargin = null;
		}
		observeNearViewportElements(document);
	}

	function handleAfterSwap(evt) {
		if (!evt.detail || !evt.detail.target) {
			return;
		}
		eagerLoadImages(evt.detail.target);
		observeNearViewportElements(evt.detail.target);
		observeNearViewportElements(document);
	}

	function init() {
		refreshObserver();
		document.body.addEventListener("htmx:afterSwap", handleAfterSwap);
		document.body.addEventListener("htmx:afterSettle", function(evt) {
			if (!evt.detail || !evt.detail.target) {
				return;
			}
			observeNearViewportElements(evt.detail.target);
			observeNearViewportElements(document);
		});
		window.addEventListener("pageshow", refreshObserver);
	}

	if (document.readyState === "loading") {
		document.addEventListener("DOMContentLoaded", init);
	} else {
		init();
	}
})();
