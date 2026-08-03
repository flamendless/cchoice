(function() {
	const PRELOAD_SECTIONS = 10;
	const DEFAULT_SECTION_HEIGHT = 360;
	const TOP_ROOT_MARGIN = 400;
	const BATCH_SIZE = 4;
	const PREFETCH_CONCURRENCY = 4;
	const SHOP_PRELOAD_SELECTOR = "[data-shop-preload]:not([data-section-loaded])";

	let observer = null;
	const observed = new Set();
	const inFlightBatches = new Set();
	const prefetchedURLs = new Set();
	let lastRootMargin = null;

	function estimateSectionHeight() {
		const sections = document.querySelectorAll(
			"[data-category-product-section][data-section-loaded]"
		);
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
		const bottomPx = PRELOAD_SECTIONS * estimateSectionHeight();
		return TOP_ROOT_MARGIN + "px 0px " + bottomPx + "px 0px";
	}

	function sortByVisibility(images) {
		return images.slice().sort(function(a, b) {
			const aRect = a.getBoundingClientRect();
			const bRect = b.getBoundingClientRect();
			const aVisible = aRect.bottom > 0 && aRect.top < window.innerHeight;
			const bVisible = bRect.bottom > 0 && bRect.top < window.innerHeight;
			if (aVisible !== bVisible) {
				return aVisible ? -1 : 1;
			}
			return aRect.top - bRect.top;
		});
	}

	function prefetchGridImages(containers, options) {
		const concurrency = (options && options.concurrency) || PREFETCH_CONCURRENCY;
		const urls = [];

		(containers || []).forEach(function(container) {
			if (!container || !container.querySelectorAll) {
				return;
			}
			container.querySelectorAll("img[data-product-grid-image]").forEach(function(img) {
				const src = img.currentSrc || img.src;
				if (!src || prefetchedURLs.has(src)) {
					return;
				}
				prefetchedURLs.add(src);
				urls.push(src);
			});
		});

		if (urls.length === 0) {
			return;
		}

		let index = 0;
		function worker() {
			while (index < urls.length) {
				const url = urls[index++];
				const image = new Image();
				image.decoding = "async";
				image.src = url;
			}
		}

		const workers = Math.min(concurrency, urls.length);
		for (let i = 0; i < workers; i++) {
			worker();
		}
	}

	function markSectionLoaded(el) {
		if (!(el instanceof Element)) {
			return;
		}
		el.setAttribute("data-section-loaded", "true");
		el.removeAttribute("data-shop-preload");
		observed.delete(el);
	}

	function markLoadedFromResponse(root) {
		if (!root) {
			return;
		}
		if (root.matches && root.matches("[data-category-product-section]")) {
			markSectionLoaded(root);
		}
		root.querySelectorAll("[data-category-product-section]").forEach(markSectionLoaded);
		root.querySelectorAll("[hx-swap-oob]").forEach(function(el) {
			if (el.id) {
				const target = document.getElementById(el.id);
				if (target) {
					markSectionLoaded(target);
				}
			}
		});
	}

	function collectBatch(triggerEl) {
		const pending = Array.from(document.querySelectorAll(SHOP_PRELOAD_SELECTOR));
		const startIdx = pending.indexOf(triggerEl);
		if (startIdx < 0) {
			return [];
		}
		return pending.slice(startIdx, startIdx + BATCH_SIZE);
	}

	function batchKey(batch) {
		return batch.map(function(el) {
			return el.getAttribute("data-section-id") || el.id;
		}).join(",");
	}

	function requestBatch(triggerEl) {
		const batch = collectBatch(triggerEl);
		if (batch.length === 0) {
			return;
		}

		const key = batchKey(batch);
		if (inFlightBatches.has(key)) {
			return;
		}

		const batchURL = triggerEl.getAttribute("data-shop-batch-url");
		if (!batchURL) {
			return;
		}

		const ids = batch.map(function(el) {
			return el.getAttribute("data-section-id");
		}).filter(Boolean);

		if (ids.length === 0) {
			return;
		}

		inFlightBatches.add(key);

		const url = batchURL + (batchURL.indexOf("?") >= 0 ? "&" : "?") + "ids=" + encodeURIComponent(ids.join(","));

		htmx.ajax("GET", url, {
			source: triggerEl,
			swap: "none"
		}).then(function() {
			inFlightBatches.delete(key);
		}).catch(function() {
			inFlightBatches.delete(key);
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
					requestBatch(entry.target);
				}
			});
		}, { rootMargin: rootMargin, threshold: 0 });

		return observer;
	}

	function observeElement(el) {
		if (!(el instanceof Element) || !document.contains(el)) {
			return;
		}
		if (!el.hasAttribute("data-shop-preload")) {
			return;
		}
		if (el.hasAttribute("data-section-loaded")) {
			return;
		}
		if (observed.has(el)) {
			return;
		}

		observed.add(el);
		createObserver().observe(el);
	}

	function observePreloadElements(root) {
		if (root instanceof Element) {
			observeElement(root);
		}

		const scope = root && root.querySelectorAll ? root : document;
		scope.querySelectorAll(SHOP_PRELOAD_SELECTOR).forEach(observeElement);
	}

	function refreshObserver() {
		if (observer) {
			observer.disconnect();
			observed.clear();
			observer = null;
			lastRootMargin = null;
		}
		observePreloadElements(document);
	}

	function handleAfterSwap(evt) {
		if (!evt.detail || !evt.detail.target) {
			return;
		}
		markLoadedFromResponse(evt.detail.target);
		if (evt.detail.xhr && evt.detail.xhr.responseText) {
			const parser = new DOMParser();
			const doc = parser.parseFromString(evt.detail.xhr.responseText, "text/html");
			doc.querySelectorAll("[hx-swap-oob]").forEach(function(el) {
				if (el.id) {
					const target = document.getElementById(el.id);
					if (target) {
						markSectionLoaded(target);
						prefetchGridImages([target], { concurrency: PREFETCH_CONCURRENCY });
					}
				}
			});
		}
		prefetchGridImages([evt.detail.target], { concurrency: PREFETCH_CONCURRENCY });
		observePreloadElements(evt.detail.target);
		observePreloadElements(document);
	}

	function init() {
		prefetchGridImages([document.body], { concurrency: PREFETCH_CONCURRENCY });
		refreshObserver();
		document.body.addEventListener("htmx:afterSwap", handleAfterSwap);
		document.body.addEventListener("htmx:afterSettle", function(evt) {
			if (!evt.detail || !evt.detail.target) {
				return;
			}
			markLoadedFromResponse(evt.detail.target);
			prefetchGridImages([evt.detail.target], { concurrency: PREFETCH_CONCURRENCY });
			observePreloadElements(evt.detail.target);
			observePreloadElements(document);
		});
		window.addEventListener("pageshow", refreshObserver);
	}

	if (document.readyState === "loading") {
		document.addEventListener("DOMContentLoaded", init);
	} else {
		init();
	}
})();
