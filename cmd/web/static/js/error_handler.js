window.onload = function() {
	const mobile = (/iphone|ipad|ipod|android|blackberry|mini|windows\sce|palm/i.test(navigator.userAgent.toLowerCase()));
	if (mobile) {
		alert("This site is still in development. Visit this on a computer for better view.");
	}
};

document.body.addEventListener("htmx:afterRequest", function(evt) {
	const el_error_banner = document.getElementById("error_banner");
	if (el_error_banner === null) {
		console.warn("no element for handling error banner found");
		return
	}

	if (evt.detail.successful) {
		el_error_banner.setAttribute("hidden", "true");

	} else if (evt.detail.failed && evt.detail.xhr) {
		console.warn("Server error", evt.detail);

		const el_error_text = document.getElementById("error_banner_text");
		const el_error_closer = document.getElementById("error_banner_closer");
		const xhr = evt.detail.xhr;
		el_error_text.innerText = `Error: ${xhr.statusText} (${xhr.status}) - ${xhr.responseText}`;
		el_error_closer.removeAttribute("hidden");
		el_error_banner.removeAttribute("hidden");

	}
});
