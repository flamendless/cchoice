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

	const params = new URLSearchParams({ event: ev, value: value })
	const url = `${baseurl}/collect/event?${params.toString()}`
	fetch(url, {
		method: "POST",
		keepalive: true,
	})
	.catch(error => console.error("Error:", error));

	console.debug(`Event: '${ev}'. Value: '${value}'`);
}

console.log("metrics event loaded");
