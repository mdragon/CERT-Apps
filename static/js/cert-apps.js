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
  this.resource('member', function()
  	{	
  		this.route('update');
  	});
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
		a.then(function(obj){ obj.data.year = 2014; });
		a.then(function(obj){ var move = this.moveUpData(obj); move.Member = CERTApps.Member.create(move.Member); console.log('App model returning', move); obj = move; return move; }.bind(this));

		console.groupEnd();

		return a;
	},

	setupController: function(controller, model)
	{
		console.group("CERTApps.ApplicationRouter setupController");

		console.log('model, controller, args', model, controller, arguments);

		controller.set('content', model);

		console.groupEnd();
	}	
});

CERTApps.ApplicationController = Ember.Controller.extend(
{

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

CERTApps.LandingController = Ember.Controller.extend(
{
	model: function(params)
	{
		console.group("CERTApps.LandingController model");

		console.log('params, args', params, arguments);

		console.groupEnd();
	}	
});

CERTApps.MemberRoute = Ember.Route.extend(
{
	actions:
	{
		saveContact: function(member)
		{

			console.group('CERTApps.MemberRoute actions.saveContact');
			console.log('member, args', member, arguments);

			member.save()

			console.groupEnd();
		}
	},

	model: function(params)
	{
		console.group("CERTApps.MemberRoute model");

		console.log('params, args', params, arguments);

		var appModel = this.modelFor('application');
		//var model = null;
		var model = appModel.data.Member;

		console.log('appModel, model', appModel, model);

		console.groupEnd();

		return model;
	},

	setupController: function(controller, model)
	{
		console.group("CERTApps.MemberRoute setupController");

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

CERTApps.Member = Ember.Object.extend(
{
	save: function()
	{
		console.group("CERTApps.Member save")
		
		console.log('saving', this);

		var settings = 
		{
			url: '/memberSave',
			type: 'json',
			dataType: 'json',
			method: 'post',
			data: JSON.stringify(this) + "\r\n"
		};

		console.log('save member request', settings)

		var a = $.ajax(settings);
		a.then(function(obj){ this.sync(obj) }.bind(this));

		console.groupEnd();
	},

	sync: function(obj)
	{
		console.group("CERTApps.Member sync");

		console.log('this', JSON.stringify(this));
		console.log('obj', JSON.stringify(obj));

		console.groupEnd();
	}

});
Ember.RSVP.configure('onerror', function(error) {
  Ember.Logger.assert(false, error);
});