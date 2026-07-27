(function() {
	function filterCards(query) {
		const cards = document.querySelectorAll(".admin-dashboard-card");
		const noResults = document.getElementById("admin-dashboard-no-results");
		const normalizedQuery = query.trim().toLowerCase();
		let visibleCount = 0;

		cards.forEach(function(card) {
			const searchText = card.dataset.searchText || "";
			const matches = normalizedQuery === "" || searchText.includes(normalizedQuery);
			card.classList.toggle("hidden", !matches);
			if (matches) {
				visibleCount++;
			}
		});

		if (noResults) {
			noResults.classList.toggle("hidden", visibleCount > 0 || normalizedQuery === "");
		}
	}

	function initAdminDashboardSearch() {
		const searchInput = document.getElementById("admin-dashboard-search");
		if (!searchInput) {
			return;
		}

		searchInput.addEventListener("input", function() {
			filterCards(searchInput.value);
		});
	}

	if (document.readyState === "loading") {
		document.addEventListener("DOMContentLoaded", initAdminDashboardSearch);
	} else {
		initAdminDashboardSearch();
	}
})();
