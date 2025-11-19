(function() {
	function setupBrandFilterLinks() {
		document.querySelectorAll(".brand-filter-link").forEach(function(link) {
			if (link.dataset.clickListenerAdded)
				return

			link.dataset.clickListenerAdded = "true";
			link.addEventListener("click", function(e) {
				e.preventDefault();

				document.querySelectorAll(".brand-filter-link").forEach(function(l) {
					l.classList.remove("underline", "bg-cchoice", "font-semibold", "text-white");
				});

				this.classList.add("underline", "bg-cchoice", "font-semibold", "text-white");

				// TODO (Brandon): Implement brand filtering logic. For now, just navigate to the URL
				const brandName = this.getAttribute("data-brand-name");
				console.log("Brand filter clicked:", brandName);

				return false;
			});
		});
	}

	function initBrandFeatures() {
		setupBrandFilterLinks();
	}

	function handleHTMXContentSwap() {
		setTimeout(function() {
			setupBrandFilterLinks();
		}, 100);
	}

	if (document.readyState === "loading") {
		document.addEventListener("DOMContentLoaded", initBrandFeatures);
	} else {
		initBrandFeatures();
	}

	document.body.addEventListener("htmx:afterSwap", handleHTMXContentSwap);
	document.body.addEventListener("htmx:afterSettle", handleHTMXContentSwap);
})();

