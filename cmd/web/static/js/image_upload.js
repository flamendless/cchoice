document.addEventListener("DOMContentLoaded", function() {
	const input = document.getElementById("product_image");
	const previewContainer = document.getElementById("image-preview-container");
	const previewImage = document.getElementById("product-image-preview");

	input.addEventListener("change", function() {
		const file = input.files[0];

		if (!file) {
			previewContainer.classList.add("hidden");
			previewImage.src = "";
			return;
		}

		// Optional: validate file size (10MB)
		const maxSize = 10 * 1024 * 1024;
		if (file.size > maxSize) {
			alert("File is too large. Max size is 10MB.");
			input.value = "";
			previewContainer.classList.add("hidden");
			previewImage.src = "";
			return;
		}

		const reader = new FileReader();

		reader.onload = function(e) {
			previewImage.src = e.target.result;
			previewContainer.classList.remove("hidden");
		};

		reader.readAsDataURL(file);
	});
});
