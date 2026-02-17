const CCHOICE = "cchoice"

function metrics_event(ev, value) {
	let baseurl = ""

	const pathArray = window.location.pathname.split("/");
	if (pathArray.length >= 2) {
		if (pathArray[0] == CCHOICE || pathArray[1] == CCHOICE) {
			baseurl = "/cchoice"
		}
	}

	if (value === undefined) {
		value = "";
	}

	const url = `${baseurl}/metrics/event?event=${ev}&value=${value}`
	fetch(url, {
		method: "POST",
	})
	.catch(error => console.error("Error:", error));

	console.debug(`Event: '${ev}'. Value: '${value}'`);
}

console.log("metrics event loaded");
