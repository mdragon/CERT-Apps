Ember.RSVP.on('error', function(error) {
	console.error("RSVP error");
	Ember.Logger.assert(false, error);
});

window.ComfortStation = Ember.Application.create(
	{
		// Basic logging, e.g. "Transitioned into 'post'"
		LOG_TRANSITIONS: true, 

		// Extremely detailed logging, highlighting every internal
		// step made while transitioning into a route, including
		// `beforeModel`, `model`, and `afterModel` hooks, and
		// information about redirects and aborted transitions
		LOG_TRANSITIONS_INTERNAL: true,
	
		LOG_VIEW_LOOKUPS: true,
		LOG_ACTIVE_GENERATION: true
	});

//Ember.LOG_BINDINGS = true;


ComfortStation.Router.map(function() 
{
	this.route('landing', { path: '/' });
	this.route('team', function() {
		this.route("id", { path: ':teamID' 	}, function() {
			this.route("comfortStation", function() {
				this.route("list");
				this.route("details", { path: ':stationID'});
			});
		});
	});
});

ComfortStation.moveUpData = function(data)
{
	console.log('moveUpData', data);

	if( ! data.error )
	{
		if( data.data !== undefined )
		{
//			console.log('data has data');
			data = data.data;
			//return data.data;
		}
	}

	return data;
}

ComfortStation.BaseRoute = Ember.Route.extend(
{
	moveUpData: function(data)
	{
		return CERTApps.moveUpData(data);
	}
});

ComfortStation.ApplicationRoute = CERTApps.BaseRoute.extend(
{
	model: function(params)
	{
		console.group("CERTApps.ApplicationRouter model");
		console.log('params, args', params, arguments);

		var settings = 
		{
			url: '/member',
			type: 'json',
			dataType: 'json'
		};

		console.log('requesting data', settings)

		var a = $.ajax(settings);
		var t = a.then(function(obj)
		{ 
			var obj = this.moveUpData(obj); 

			return obj;
		}.bind(this));

		t.then(function(obj){ obj.year = 2014; });

		var t2 = t.then(function(obj)
		{
			obj.Member = CERTApps.Member.create(obj.Member); 
			console.log('App model returning', obj); 

			var model = CERTApps.AppModel.create(obj);

			var p = CERTApps.Team.lookup().then(
				function(team)
				{ 
					console.log('model resolved for team', team); 
					model.set('Team', team); 
					console.log('after set model', model); 

					return model
				}
			);

			return p; 
		}.bind(this));

		console.groupEnd();

		return t2;
	},

	setupController: function(controller, model)
	{
		console.group("CERTApps.ApplicationRouter setupController");
		console.log('model, controller, args', model, controller, arguments);

		controller.set('model', model);

		console.groupEnd();
	}	
});
