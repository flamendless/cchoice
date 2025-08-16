document.addEventListener("DOMContentLoaded", () => {
	const proceedBtn = document.getElementById("btn-proceed");

	function checkForm() {
		const requiredInputs = document.querySelectorAll("#cart-shipping [required]");
		const allValid = Array.from(requiredInputs).every(input => input.checkValidity());
		const paymentSelected = !!document.querySelector(
			"#cart-payments input[name='checked_payment_method']:checked"
		);

		if (allValid && paymentSelected) {
			proceedBtn.disabled = false;
		} else {
			proceedBtn.disabled = true;
		}
	}

	document
		.getElementById("cart-shipping")
		.addEventListener("input", checkForm);

	document
		.getElementById("cart-payments")
		.addEventListener("change", checkForm);

	checkForm();
});
