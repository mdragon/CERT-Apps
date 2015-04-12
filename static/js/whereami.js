var errorSpan = document.getElementById("error");
var latSpan = document.getElementById("lat");
var lonSpan = document.getElementById("long");
var atSpan = document.getElementById("at");

var form = document.forms["data"];
var name = form["name"];
var lat = form["lat"];
var lon = form["long"];
var updated = form["updated"];
var submit = form["submit"];

var init = function() {
	submit.disabled = true;
};

var getPrettyDate = function(date) {
	date = date || new Date();
	var hour = date.getHours().toString();
	if( hour.length == 1 ) hour = "0" + hour;
	
	var minute = date.getMinutes().toString();
	if( minute.length == 1 ) minute = "0" + minute;

	var second = date.getSeconds().toString();
	if( second.length == 1 ) second = "0" + second;

	var strTime = hour + ":" + minute + ":" + second;

	return strTime;
};

var updateLocation = function() {
	//disable until we get a position

	if (navigator.geolocation) {
		console.log('Geolocation is supported!');

	 	var position;
		navigator.geolocation.getCurrentPosition( function gotCurrentPosition(position) {			
			console.group("gotCurrentPosition");
			console.log("position", position);
			
			var latitude = position.coords.latitude;
			var longitude = position.coords.longitude;

			lat.value = latitude;
			lon.value = longitude;

			var stringLat = latitude.toString();
			var stringLong = longitude.toString();
			var len = 8;

			if( stringLat.substring(0,1) === "-") len=9;
			latSpan.innerHTML = stringLat.substring(0,len);

			if( stringLong.substring(0,1) === "-") {
				len=9 
			} else {
				len=8;	
			} 
			lonSpan.innerHTML = stringLong.substring(0,len);

			var date = new Date();
			strTime = getPrettyDate(date);

			atSpan.innerHTML = "as of " + strTime;

			updated.value = date.toISOString();// + "00:00";

			submit.disabled = false;

			console.groupEnd()
		}, function(error) {
			var errMsg = null;

			switch(error.code)
			{
				case 0: 
					errMsg = "Unknown Error";
					break;
				case 1: 
					errMsg = "Permission denied, please refresh the page and/or \"Allow\" the Location permission";
					break;
				case 2: 
					errMsg = "Location unavailble, try again";
					breal;
				case 3: 
					errMsg = "Location request timed out, try again.";
			}

			errorSpan.innerText = "Geolocation error: " + errMsg;
		});
	} else {
		console.error('Geolocation is not supported!');
		errorSpan.innerText = "Cannot use Geolocation.  Cannot continue in this browser";
		name.disabled = true;
		submit.disabled = true;
	}
}

window.onload = init();
