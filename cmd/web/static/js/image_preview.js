function handleImagePreview(input) {
	const file = input.files[0];
	if (!file) return;
	const previewId = input.closest('.flex.flex-col.gap-4')?.querySelector('img[id$="-preview"]')?.id;
	if (!previewId) return;
	const preview = document.getElementById(previewId);
	if (!preview) return;
	if (file.type.startsWith('image/')) {
		const reader = new FileReader();
		reader.onload = function(e) {
			preview.src = e.target.result;
			preview.classList.remove('hidden');
		};
		reader.readAsDataURL(file);
	}
}

function handlePromoImagePreview(input) {
	const file = input.files[0];
	if (!file) return;
	const previewContainer = document.getElementById('promo-image-preview');
	if (!previewContainer) return;
	const preview = previewContainer.querySelector('img');
	if (!preview) return;
	if (file.type.startsWith('image/')) {
		const reader = new FileReader();
		reader.onload = function(e) {
			preview.src = e.target.result;
			preview.classList.remove('hidden');
			previewContainer.classList.remove('hidden');
		};
		reader.readAsDataURL(file);
	}
}
