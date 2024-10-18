function add_fade(targetID) {
	//TODO: (Brandon) - result is inconsistent
	const el = document.getElementById(targetID);
	const is_overflowing = el.clientWidth < el.scrollWidth;
	const images = el.querySelectorAll("div img");
	console.log("loaded category products", images.length, is_overflowing);

	for (const img of images) {
		const b = img.getBoundingClientRect();
		const x = b.left + window.scrollX;

		const is_overflown_r = x + b.width > window.screen.width;
		const is_overflown_l = x - b.width < 0;

		if (is_overflown_l) {
			img.classList.add("custom-mask-image-left");
		} else if (is_overflown_r) {
			img.classList.add("custom-mask-image-right");
		}
	}
}

document.body.addEventListener("htmx:afterProcessNode", function(evt) {
	const targetID = evt.detail.elt.id;
	if (!targetID.startsWith("products_categories_products_")) {
		return
	}

	// add_fade(targetID);

	let scroll_amount = 0;
	const scroll_inc = 64;
	const el = document.getElementById(targetID);
	const is_right = el.classList.contains("flex-row-reverse");
	const slide_timer = setInterval(function() {
		if (is_right) {
			scroll_amount -= scroll_inc;
		} else {
			scroll_amount += scroll_inc;
		}
		el.scrollTo({
			top: 0,
			left: scroll_amount,
			behavior: "smooth",
		});
	}, 1000);
});
