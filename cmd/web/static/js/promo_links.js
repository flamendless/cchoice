(function() {
	"use strict";

	function isValidPromoCustomURL(url) {
		const value = (url || "").trim();
		if (!value) {
			return false;
		}
		return value.startsWith("http://") || value.startsWith("https://") || value.startsWith("/");
	}

	function validatePromoCustomLinkInput(input) {
		if (!input) {
			return false;
		}
		const value = input.value.trim();
		if (!value) {
			input.setCustomValidity("Please enter a custom URL.");
			input.reportValidity();
			return false;
		}
		if (!isValidPromoCustomURL(value)) {
			input.setCustomValidity("URL must be an absolute https:// URL or a path starting with /.");
			input.reportValidity();
			return false;
		}
		input.setCustomValidity("");
		return true;
	}

	function getPromoType(form) {
		const typeSelect = form.querySelector('select[name="type"]');
		if (typeSelect) {
			return typeSelect.value;
		}
		const typeInput = form.querySelector('input[name="type"]');
		return typeInput ? typeInput.value : "";
	}

	function setPromoLinkFieldState(linkType) {
		const trackedField = document.getElementById("promo_link_tracked_field");
		const customField = document.getElementById("promo_link_custom_field");
		if (!trackedField || !customField) {
			return;
		}
		const trackedSelect = trackedField.querySelector('select[name="tracked_link_id"]');
		const customInput = customField.querySelector('input[name="link_url"]');
		const isTracked = linkType === "TRACKED";
		const isCustom = linkType === "CUSTOM";

		trackedField.classList.toggle("hidden", !isTracked);
		customField.classList.toggle("hidden", !isCustom);

		if (trackedSelect) {
			trackedSelect.disabled = !isTracked;
			if (!isTracked) {
				trackedSelect.value = "";
			}
		}
		if (customInput) {
			customInput.disabled = !isCustom;
			if (!isCustom) {
				customInput.value = "";
				customInput.setCustomValidity("");
			}
		}
	}

	function validatePromoLinkFields(event) {
		const form = event.currentTarget;
		if (getPromoType(form) !== "BANNER_IMAGE") {
			return true;
		}

		const linkType = form.querySelector('input[name="link_type"]:checked')?.value;
		if (linkType === "TRACKED") {
			const trackedSelect = form.querySelector('select[name="tracked_link_id"]');
			if (!trackedSelect || !trackedSelect.value) {
				event.preventDefault();
				event.stopPropagation();
				alert("Please select a tracked link.");
				trackedSelect?.focus();
				return false;
			}
		}

		if (linkType === "CUSTOM") {
			const customInput = form.querySelector('input[name="link_url"]');
			if (!validatePromoCustomLinkInput(customInput)) {
				event.preventDefault();
				event.stopPropagation();
				return false;
			}
		}

		return true;
	}

	function togglePromoLinkType(linkType) {
		setPromoLinkFieldState(linkType);
	}

	window.isValidPromoCustomURL = isValidPromoCustomURL;
	window.validatePromoCustomLinkInput = validatePromoCustomLinkInput;
	window.setPromoLinkFieldState = setPromoLinkFieldState;
	window.validatePromoLinkFields = validatePromoLinkFields;
	window.togglePromoLinkType = togglePromoLinkType;
})();
