function handleLogoPreview(input) {
	const file = input.files[0];
	const previewId = input.closest('.flex.flex-col.gap-4').querySelector('img[id$="logo-preview"]')?.id;
	if (!previewId) return;
	const preview = document.getElementById(previewId);
	if (!preview) return;
	if (file) {
		const reader = new FileReader();
		reader.onload = function(e) {
			preview.src = e.target.result;
			preview.classList.remove('hidden');
		};
		reader.readAsDataURL(file);
	}
}
