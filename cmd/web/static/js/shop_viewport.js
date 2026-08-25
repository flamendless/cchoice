(function() {
	const BATCH_SIZE = 4;
	const PREFETCH_CONCURRENCY = 4;
	const SHOP_PRELOAD_SELECTOR = "[data-shop-preload]:not([data-section-loaded])";

	const inFlightBatches = new Set();
	const prefetchedURLs = new Set();
	let draining = false;

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
			return null;
		}

		const key = batchKey(batch);
		if (inFlightBatches.has(key)) {
			return null;
		}

		const batchURL = triggerEl.getAttribute("data-shop-batch-url");
		if (!batchURL) {
			return null;
		}

		const ids = batch.map(function(el) {
			return el.getAttribute("data-section-id");
		}).filter(Boolean);

		if (ids.length === 0) {
			return null;
		}

		inFlightBatches.add(key);

		const url = batchURL + (batchURL.indexOf("?") >= 0 ? "&" : "?") + "ids=" + encodeURIComponent(ids.join(","));

		return htmx.ajax("GET", url, {
			source: triggerEl,
			swap: "none"
		}).finally(function() {
			inFlightBatches.delete(key);
		});
	}

	function drainPendingBatches() {
		if (draining) {
			return;
		}

		const first = document.querySelector(SHOP_PRELOAD_SELECTOR);
		if (!first) {
			return;
		}

		const promise = requestBatch(first);
		if (!promise) {
			return;
		}

		draining = true;
		promise.finally(function() {
			draining = false;
			drainPendingBatches();
		});
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
		drainPendingBatches();
	}

	function init() {
		prefetchGridImages([document.body], { concurrency: PREFETCH_CONCURRENCY });
		drainPendingBatches();
		document.body.addEventListener("htmx:afterSwap", handleAfterSwap);
		document.body.addEventListener("htmx:afterSettle", function(evt) {
			if (!evt.detail || !evt.detail.target) {
				return;
			}
			markLoadedFromResponse(evt.detail.target);
			prefetchGridImages([evt.detail.target], { concurrency: PREFETCH_CONCURRENCY });
			drainPendingBatches();
		});
		window.addEventListener("pageshow", drainPendingBatches);
	}

	if (document.readyState === "loading") {
		document.addEventListener("DOMContentLoaded", init);
	} else {
		init();
	}
})();
