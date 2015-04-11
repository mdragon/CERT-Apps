var errorSpan = document.getElementById("error");
var latSpan = document.getElementById("lat");
var lonSpan = document.getElementById("long");
var atSpan = document.getElementById("at");

var form = document.forms["data"];
var name = form["name"];
var lat = form["lat"];
var lon = form["long"];
var submit = form["submit"];

var init = function() {
	submit.disabled = true;
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
			strTime = date.getHours() + ":" + date.getMinutes() + ":" + date.getSeconds();

			atSpan.innerHTML = "as of " + strTime;

			submit.disabled = false;

			console.groupEnd()
		});
	} else {
		console.error('Geolocation is not supported!');
		errorSpan.innerText = "Cannot use Geolocation.  Cannot continue in this browser";
		name.disabled = true;
		submit.disabled = true;
	}
}

window.onload = init();
