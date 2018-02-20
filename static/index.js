/* Copyright 2018 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
var spinner;

function initSpinner() {
	var opts = {
		lines: 8, // The number of lines to draw
		length: 19, // The length of each line
		width: 4, // The line thickness
		radius: 11, // The radius of the inner circle
		scale: 1, // Scales overall size of the spinner
		corners: 1, // Corner roundness (0..1)
		color: '#ffb300', // CSS color or array of colors
		fadeColor: 'transparent', // CSS color or array of colors
		opacity: 0.25, // Opacity of the lines
		rotate: 0, // The rotation offset
		direction: 1, // 1: clockwise, -1: counterclockwise
		speed: 1.2, // Rounds per second
		trail: 57, // Afterglow percentage
		fps: 20, // Frames per second when using setTimeout() as a fallback in IE 9
		zIndex: 2e9, // The z-index (defaults to 2000000000)
		className: 'spinner', // The CSS class to assign to the spinner
		top: '50%', // Top position relative to parent
		left: '50%', // Left position relative to parent
		position: 'absolute' // Element positioning
	};

	spinner = new Spinner(opts)
}

function spin(target) {
	if (!spinner) {
		initSpinner();
	}
	spinner.spin(target);
}

function stopSpinner() {
	if (!spinner) {
		return;
	}
	spinner.stop();
}

function renderGraph(data) {
	d3.select("#graph").graphviz()
		.onerror(function(err) {
			console.log("d3-graphviz error: " + err);
		})
		.fade(true)
		.renderDot(data);
}

window.onload = function() {
	var spinnerTarget = document.getElementById("spinner");
	var err = document.getElementById("error");
	document.getElementById("typeform").addEventListener("submit", function(e){
		e.preventDefault();
		spin(spinnerTarget);
		var t = document.getElementById("type");
		d3.selectAll('svg').remove();
		d3.json('/rawdot?type=' + t.value, function(data) {
			err.innerHTML = data.Error;
			if (data.Error != "") {
				stopSpinner();
				return;
			}
			renderGraph(data.DOT);
			stopSpinner();
		});
	});
}