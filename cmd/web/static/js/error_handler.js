window.onload = function() {
	// const mobile = (/iphone|ipad|ipod|android|blackberry|mini|windows\sce|palm/i.test(navigator.userAgent.toLowerCase()));
	// if (mobile) {
	// 	alert("This site is still in development. Visit this on a computer for better view.");
	// }

	checkForBannerMessages();
};

function checkForBannerMessages() {
	const url = new URL(window.location);
	const errorMsg = url.searchParams.get("error");
	const successMsg = url.searchParams.get("success");

	if (errorMsg) {
		showErrorBanner(decodeURIComponent(errorMsg));
		url.searchParams.delete("error");
		window.history.replaceState({}, "", url.toString());
	} else if (successMsg) {
		showSuccessBanner(decodeURIComponent(successMsg));
		url.searchParams.delete("success");
		window.history.replaceState({}, "", url.toString());
	}
}

if (typeof successTimeout === 'undefined') {
	var successTimeout = null;
}

document.body.addEventListener("htmx:afterRequest", function(evt) {
	const el_error_banner = document.getElementById("error_banner");
	const el_success_banner = document.getElementById("success_banner");

	if (evt.detail.successful) {
		el_error_banner.setAttribute("hidden", "true");

		const xhr = evt.detail.xhr;
		const successMessage = xhr.getResponseHeader("X-Success-Message");
		if (successMessage) {
			showSuccessBanner(successMessage);
		} else {
			el_success_banner.setAttribute("hidden", "true");
		}

	} else if (evt.detail.failed && evt.detail.xhr) {
		console.warn("Server error", evt.detail);

		const el_error_text = document.getElementById("error_banner_text");
		const el_error_closer = document.getElementById("error_banner_closer");
		const xhr = evt.detail.xhr;

		const errorMessage = xhr.getResponseHeader("X-Error-Message");
		if (errorMessage) {
			el_error_text.innerText = errorMessage;
		} else {
			el_error_text.innerText = `Error: ${xhr.statusText} (${xhr.status}) - ${xhr.responseText}`;
		}
		el_error_closer.removeAttribute("hidden");
		el_error_banner.removeAttribute("hidden");

		el_success_banner.setAttribute("hidden", "true");
	}
});

function showSuccessBanner(message) {
	const el_success_banner = document.getElementById("success_banner");
	const el_success_text = document.getElementById("success_banner_text");
	const el_success_closer = document.getElementById("success_banner_closer");

	el_success_text.innerText = message;
	el_success_closer.removeAttribute("hidden");
	el_success_banner.removeAttribute("hidden");

	if (successTimeout) {
		clearTimeout(successTimeout);
	}

	successTimeout = setTimeout(function() {
		el_success_banner.setAttribute("hidden", "true");
	}, 3000);
}

function showErrorBanner(message) {
	const el_error_banner = document.getElementById("error_banner");
	const el_error_text = document.getElementById("error_banner_text");
	const el_error_closer = document.getElementById("error_banner_closer");

	el_error_text.innerText = message;
	el_error_closer.removeAttribute("hidden");
	el_error_banner.removeAttribute("hidden");

	if (typeof errorTimeout === 'undefined') {
		var errorTimeout = null;
	}

	if (errorTimeout) {
		clearTimeout(errorTimeout);
	}

	errorTimeout = setTimeout(function() {
		el_error_banner.setAttribute("hidden", "true");
	}, 5000);
}
