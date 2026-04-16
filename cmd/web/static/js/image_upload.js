(function() {
	"use strict";

	const DEFAULT_MAX_SIZE = 10 * 1024 * 1024; // 10MB

	function initImageUpload(input) {
		if (!input) return;

		const maxSize = parseInt(input.dataset.maxSize) || DEFAULT_MAX_SIZE;
		const previewContainer = input.closest("form")?.querySelector('[data-upload-preview-container]')
			|| document.getElementById(input.id + "-preview-container")
			|| document.querySelector('[data-upload-preview-container]');
		const previewImage = input.closest("form")?.querySelector('[data-upload-preview-image]')
			|| document.getElementById(input.id + "-preview")
			|| document.querySelector('[data-upload-preview-image]');

		input.addEventListener("change", function() {
			const file = this.files[0];

			if (!file) {
				if (previewContainer && previewImage) {
					previewContainer.classList.add("hidden");
					previewImage.src = "";
				}
				return;
			}

			if (file.size > maxSize) {
				alert("File is too large. Max size is " + (maxSize / 1024 / 1024) + "MB.");
				this.value = "";
				if (previewContainer) previewContainer.classList.add("hidden");
				if (previewImage) previewImage.src = "";
				return;
			}

			if (!file.type.startsWith("image/")) {
				alert("Please select an image file.");
				this.value = "";
				return;
			}

			const reader = new FileReader();
			reader.onload = function(e) {
				if (previewImage) {
					previewImage.src = e.target.result;
					previewImage.classList.remove("hidden");
				}
				if (previewContainer) {
					previewContainer.classList.remove("hidden");
				}
			};
			reader.readAsDataURL(file);
		});
	}

	document.addEventListener("DOMContentLoaded", function() {
		const inputs = document.querySelectorAll("[data-upload-preview]");
		inputs.forEach(function(input) {
			initImageUpload(input);
		});

		document.addEventListener("htmx:load", function(event) {
			const inputs = event.target.querySelectorAll("[data-upload-preview]");
			inputs.forEach(function(input) {
				initImageUpload(input);
			});
		});
	});
})();
