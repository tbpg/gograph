window.onload = function() {
	document.getElementById("typeform").addEventListener("submit", function(e){
		e.preventDefault();
		var t = document.getElementById("type");
		d3.selectAll('svg').remove();
		d3.text('/rawdot?type=' + t.value, function(data) {
			d3.select("#graph").graphviz()
			.fade(true)
			.renderDot(data);
		});
	});
}