function show(id) {
	var a = document.getElementById("example_" + id);
	var d = document.getElementById("d_" + id);
	d.style.display = "block";
	a.innerHTML = "☟ <i>Example</i>";
}
function ex(id) {
	var a = document.getElementById("example_" + id);
	var d = document.getElementById("d_" + id);
	if (d.style.display === "block") {
		d.style.display = "none";
		a.innerHTML = "☞ <i>Example</i>";
	} else {
		d.style.display = "block";
		a.innerHTML = "☟ <i>Example</i>";
	}
}
