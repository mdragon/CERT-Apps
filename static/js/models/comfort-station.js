CERTModels.ComfortStation = CERTModels.BaseObject.extend(
{	
	init: function()
	{
		console.log('CERTApps.ComfortStation init');
	},


	reset: function()
	{
		//console.log('CERTApps.ComfortStation init');

		this.set("KeyID", 0);
	},

	save: function(teamID)
	{
		console.group("CERTApps.ComfortStation save");

		var key = this.get("KeyID") || ""

		var options =  {
			url: "/api/comfort-station/" + key,
			data: {Station: this, Team: {KeyID: teamID}}
		};

		var p = CERTApps.ajax(options);

		var o = p.then(function comfortStationSaveSuccess(obj) {
			console.group("CERTApps.ComfortStation comfortStationSaveSuccess");

			this.sync(obj.Station);

			console.groupEnd();

			return this;

		}.bind(this)
		);

		return o;
	},

	line2Display: function()
	{
		return CERTModels.line2Display(this);
	}.property("line2")
});	

CERTModels.ComfortStation.reopenClass(
{
	fetch: function(stationID)
	{
		console.group("CERTApps.ComfortStation.fetch")
		console.log("stationID", stationID);

		var options = 
		{
			url: "/api/comfort-station/" + stationID,
			type: "get"
		};

		var t = CERTApps.ajax(options);

		return t;
	},

	fetchForTeam: function(teamID)
	{
		console.group("CERTApps.ComfortStation.fetchForTeam")
		console.log("teamID", teamID);

		var options = 
		{
			url: "/api/comfort-stations/" + teamID,
			type: "get"
		};

		var t = CERTApps.ajax(options);

		return t;
	}
});
