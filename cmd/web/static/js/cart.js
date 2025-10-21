document.addEventListener("DOMContentLoaded", () => {
	const proceedBtn = document.getElementById("btn-proceed");

	async function checkShippingQuotationStatus() {
		try {
			const response = await fetch("/cchoice/shipping/quotation/status");
			return response.status === 200;
		} catch (error) {
			console.error("Error checking shipping quotation status:", error);
			return false;
		}
	}

	function areAddressFieldsComplete() {
		const shippingForm = document.getElementById("shipping-form");
		if (!shippingForm) {
			console.log("Shipping form not found");
			return false;
		}

		const requiredFields = shippingForm.querySelectorAll("[required]");
		let allValid = true;
		const fieldValues = {};

		requiredFields.forEach((field, index) => {
			const fieldName = field.name || field.id || `field-${index}`;
			const fieldValue = field.value;
			const isValid = fieldValue && fieldValue.trim() !== "";

			fieldValues[fieldName] = {
				value: fieldValue,
				valid: isValid
			};

			if (!isValid) {
				allValid = false;
			}
		});

		// console.log("Address validation:", fieldValues);
		// console.log("All address fields valid:", allValid);

		return allValid;
	}

	function hasCheckedItems() {
		const checkedItems = document.querySelectorAll("input[name='checked_item']:checked");
		return checkedItems.length > 0;
	}

	async function checkForm() {
		const requiredInputs = document.querySelectorAll("#cart-shipping [required]");
		const allValid = Array.from(requiredInputs).every(input => input.checkValidity());
		const paymentSelected = !!document.querySelector(
			"#cart-payments input[name='checked_payment_method']:checked"
		);

		const addressComplete = areAddressFieldsComplete();
		const shippingQuotationExists = addressComplete ? await checkShippingQuotationStatus() : false;
		const itemsChecked = hasCheckedItems();

		const shouldEnable = allValid && paymentSelected && shippingQuotationExists && itemsChecked;

		// console.log("=== Form Check Results ===");
		// console.log("HTML5 validation (allValid):", allValid);
		// console.log("Payment selected:", paymentSelected);
		// console.log("Address complete:", addressComplete);
		// console.log("Shipping quotation exists:", shippingQuotationExists);
		// console.log("Items checked:", itemsChecked);
		// console.log("Final decision (should enable):", shouldEnable);
		// console.log("=========================");

		if (shouldEnable) {
			proceedBtn.disabled = false;
		} else {
			proceedBtn.disabled = true;
		}
	}

	document
		.getElementById("cart-shipping")
		.addEventListener("input", checkForm);

	document
		.getElementById("cart-shipping")
		.addEventListener("change", checkForm);

	document
		.getElementById("cart-payments")
		.addEventListener("change", checkForm);

	document.body.addEventListener("htmx:afterRequest", (event) => {
		if (event.detail.elt && event.detail.elt.id === "shipping-form") {
			checkForm();
		}
	});

	// Listen for checkbox changes in cart lines
	document.addEventListener("change", (event) => {
		if (event.target.name === "checked_item") {
			checkForm();
		}
	});

	// Listen for HTMX requests that might affect cart lines
	document.body.addEventListener("htmx:afterRequest", (event) => {
		if (event.detail.elt && (
			event.detail.elt.id === "cart-lines" ||
			event.detail.elt.closest("#cart-lines")
		)) {
			checkForm();
		}
	});

	checkForm();
});
