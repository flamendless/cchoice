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

function add_auto_scroll(targetID) {
	const el = document.getElementById(targetID);

	const main_timer = new easytimer.Timer();
	const sub_timer = new easytimer.Timer();

	main_timer.start({ countdown: true, startValues: { seconds: 3 } });
	main_timer.addEventListener("targetAchieved", function(e) {
		sub_timer.reset();
	});

	sub_timer.start({ precision: "secondTenths", startValues: { seconds: 0 }, target: { seconds: 2 } });
	sub_timer.stop();
	sub_timer.addEventListener("secondTenthsUpdated", function(e) {
		const is_right = el.classList.contains("flex-row");
		const x = 8;
		if (is_right) {
			el.scrollBy({
				top: 0,
				left: x,
				behavior: "smooth",
			});
		} else {
			el.scrollBy({
				top: 0,
				left: -x,
				behavior: "smooth",
			});
		}
	});
	sub_timer.addEventListener("targetAchieved", function(e) {
		main_timer.reset();
	});
}

document.body.addEventListener("htmx:afterProcessNode", function(evt) {
	const targetID = evt.detail.elt.id;
	if (!targetID.startsWith("products_categories_products_")) {
		return
	}
	// add_fade(targetID);
	add_auto_scroll(targetID);
});
