(function() {
	let categoryObserver = null;

	function getHeaderHeight() {
		const header = document.querySelector("header");
		return header ? header.offsetHeight : 150;
	}

	function setupCategoryScrollLinks() {
		document.querySelectorAll(".category-scroll-link").forEach(function(link) {
			if (link.dataset.scrollListenerAdded)
				return
			link.dataset.scrollListenerAdded = "true";
			link.addEventListener("click", function(e) {
				e.preventDefault();
				const targetId = this.getAttribute("data-scroll-target");
				const target = document.getElementById(targetId);
				if (target) {
					const headerHeight = getHeaderHeight();
					const scrollPos = target.offsetTop - headerHeight;
					window.scrollTo({top: scrollPos, behavior: "smooth"});
				}
				return false;
			});
		});
	}

	function setupCategoryActiveHighlight() {
		if (categoryObserver) {
			categoryObserver.disconnect();
		}

		const categoryLinks = document.querySelectorAll(".category-scroll-link");
		if (categoryLinks.length === 0)
			return

		const categorySections = Array.from(categoryLinks).map(function(link) {
			const targetId = link.getAttribute("data-scroll-target");
			return {
				link: link,
				section: document.getElementById(targetId)
			};
		}).filter(function(item) {
			return item.section !== null;
		});

		if (categorySections.length === 0)
			return

		const headerHeight = getHeaderHeight();
		const rootMargin = "-" + (headerHeight + 20) + "px 0px -60% 0px";

		categoryObserver = new IntersectionObserver(function(entries) {
			entries.forEach(function(entry) {
				if (entry.isIntersecting) {
					const targetId = entry.target.id;
					const link = document.getElementById("category-link-" + targetId);
					if (link) {
						document.querySelectorAll(".category-scroll-link").forEach(function(l) {
							l.classList.remove("underline", "bg-cchoice", "font-semibold", "text-white");
						});
						link.classList.add("underline", "bg-cchoice", "font-semibold", "text-white");
					}
				}
			});
		}, {
			rootMargin: rootMargin,
			threshold: 0.1
		});

		categorySections.forEach(function(item) {
			categoryObserver.observe(item.section);
		});
	}

	function initCategoryFeatures() {
		setupCategoryScrollLinks();
		setupCategoryActiveHighlight();
	}

	function handleHTMXContentSwap() {
		setTimeout(function() {
			setupCategoryScrollLinks();
			setupCategoryActiveHighlight();
		}, 100);
	}

	if (document.readyState === "loading") {
		document.addEventListener("DOMContentLoaded", initCategoryFeatures);
	} else {
		initCategoryFeatures();
	}

	document.body.addEventListener("htmx:afterSwap", handleHTMXContentSwap);
	document.body.addEventListener("htmx:afterSettle", handleHTMXContentSwap);
})();

