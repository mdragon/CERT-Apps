window.CERTApps = Ember.Application.create(
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

CERTApps.Router.map(function() {
  this.resource('landing', { path: '/landing' });
});

CERTApps.ApplicationRoute = Ember.Route.extend(
{
	moveUpData: function(data)
	{
		console.log('moveUpData', data);

		if( ! data.error )
		{
			if( data.data )
			{
				console.log('data has data');
				data = data.data;
				//return data.data;
			}
		}

		return data;
	},

	model: function(params)
	{
		console.group("CERTApps.ApplicationRouter model");

		console.log('params, args', params, arguments);

		var settings = 
		{
			url: '/memberData',
			type: 'json',
			dataType: 'json'
		};

		console.log('requesting data', settings)

		var a = $.ajax(settings);
		a.then(function(data){ data.year = 2014; });
		a.then(function(data){ return this.moveUpData(data); }.bind(this));

		return a;

		console.groupEnd();
	},

	setupController: function(controller, model)
	{
		console.group("CERTApps.ApplicationRouter setupController");

		console.log('model, controller, args', model, controller, arguments);

		controller.set('content', model);

		console.groupEnd();
	}	
});

CERTApps.LandingRoute = Ember.Route.extend(
{
	model: function(params)
	{
		console.group("CERTApps.LandingRouter model");

		console.log('params, args', params, arguments);

		console.groupEnd();
	},

	setupController: function(controller, model)
	{
		console.group("CERTApps.LandingRouter setupController");

		console.log('model, controller, args', model, controller, arguments);

		console.groupEnd();
	}	

});

// CERTApps.ApplicationController = Ember.Controller.create(
// {
// 	model: function(params)
// 	{
// 		console.group("CERTApps.ApplicationController model");

// 		console.log('params, args', params, arguments);

// 		console.groupEnd();
// 	}	
// });

CERTApps.LandingController = Ember.Controller.extend(
{
	model: function(params)
	{
		console.group("CERTApps.LandingController model");

		console.log('params, args', params, arguments);

		console.groupEnd();
	}	
});

Ember.RSVP.configure('onerror', function(error) {
  Ember.Logger.assert(false, error);
});