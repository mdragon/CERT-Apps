var CERT = Ember.Application.create(
	{
		rootElement: 'body',
		LOG_TRANSITIONS: true,
		enableLogging: true,
		LOG_BINDINGS: true
	}
);

CERT.loggedIn = {};

CERT.Router.map(function(match) {
  //this.route('team');
  //this.resource('test',{ path: '/abc'});
  //this.route("favorites");
  console.log('Router.map match', match);
	this.resource("team", { path: "teams" }, function() {
		this.resource("members", { path: "members" });
	});
	this.resource("profile", { path: "profile" }, function() {
		//this.resource("myProfile", { path: '/' }),
		//this.resource('profileEdit', { path: 'whee' });
		this.resource("profileEdit", { path: ":profile_id" }, function() {
//			this.resource("profileEditSpecific", { path: "/id/:profileEditSpecific_id"});
		});
	});
	//this.resource("profile", {path: 'member/:member_id' });
});

CERT.Store = DS.Store.extend({
	revision: 12
});

CERT.Member = DS.Model.extend({
	key: DS.attr('string'),
	firstName: DS.attr('string'),
	lastName: DS.attr('string'),

	team: DS.belongsTo('CERT.Team')
});

CERT.Team = DS.Model.extend({
	key: DS.attr('string'),
	name: DS.attr('string'),
	members: DS.hasMany(CERT.Member)
});

DS.RESTAdapter.map('CERT.Member', {
	id: { key: 'key' },
	firstName: { key: 'firstName' },
	lastName: { key: 'lastName' }
});

DS.RESTAdapter.map('CERT.Team', {
	id: { key: 'key' },
	name: { key: 'name' }
});

CERT.IndexRoute = Ember.Route.extend({
  setupController: function(controller, model) {
    console.log('IndexRoute', 'controller', controller, 'model', model);
  }
});

CERT.ApplicationRoute = Ember.Route.extend(
{
	member: null,
	loginURL: null,

	setupController: function(controller, model)
	{
		console.log("ApplicationRoute", 'controller', controller, 'model', model);
		controller.set('copyrightYear', new Date().getFullYear());
		controller.set('loginURL', '/log-me-in');
		if( CERT.loggedIn ) controller.set('member', CERT.loggedIn.member);
	},

	fetchGlobals: function()
	{
		$.ajax
		(
			{
				context: this,
				dataType: 'json',
				success: loadGlobals,
				url: '/global-state'
			}
		);
	},

	loadGlobals: function(data, status, xhr)
	{
		console.group("loadGlobals");
		
		console.log('loading global data from data to controller', data, this);
		for( var p in data )
		{
			console.log('copy data', p)
			if( data.hasOwnProperty(p) )
			{

			}
		}

		console.groupEnd();
	}
});

CERT.TeamRoute = Ember.Route.extend({
	setupController: function(controller, model)
	{
		console.log("RosterRoute", 'controller', controller, 'model', model);
	},
	model: function(params)
	{
		if( CERT.loggedIn && CERT.loggedIn.member )
		{
			return CERT.Team.find();
		}

		return null;
	}
});

CERT.TeamController = Em.ArrayController.extend({
  ready: function()
  {
		console.log("Created App namespace in RosterController ready");
  }
});

CERT.TeamView = Em.View.extend(
{
	willInsertElement: function()
	{
		console.log("teamview will insert");
	},

	didInsertElement: function()
	{
		console.group('TeamView didInsertElement');

		console.log('this, this.$()', this, this.$());
		console.log('controller', this.get('controller'));

		//$("body").timeago();

		console.groupEnd();
	}
});

CERT.MembersRoute = Ember.Route.extend({
	needs: 'team',
	setupController: function(controller, model)
	{
		console.log("MembersRoute", 'controller', controller, 'model', model);
	}
/*	model: function(params)
	{
		console.log('params', params, 'this', this);
		//{team: controller.team}
		return CERT.Member.find();
	}*/
});

CERT.MembersController = Em.ArrayController.extend({
  ready: function()
  {
		console.log("Created App namespace in RosterController ready");
  }
});

CERT.LoginView = Em.View.extend({
	willInsertElement: function()
	{
		console.log('loginView willInsert', this.get('loginURL'));
	},

	didInsertElement: function()
	{
		console.log('loginView didInsert', this.get('loginURL'));
	}

});

CERT.ProfileRoute = Ember.Route.extend({
	setupController: function(controller, model)
	{
		console.log("ProfileRoute", 'controller', controller, 'model', model);
	},
	model: function(params)
	{
		console.log("ProfileRoute params", params)

		if( CERT.loggedIn && CERT.loggedIn.member )
		{
			var id = CERT.loggedIn.member.id;

			if( CERT.loggedIn.member.isAdmin() )
			{
				console.log('loading other member to edit because admin has requested', params.member_id, cert.loggedIn.member);
				if( params.member_id ) id = params.member_id;
			}
			console.log('calling find for id', id);
			return CERT.Member.find(id);
		}

		return null;
	}
});

CERT.ProfileEditRoute = Ember.Route.extend({
	setupController: function(controller, model)
	{
		console.log("ProfileEditRoute", 'controller', controller, 'model', model);
	},
	model: function(params)
	{
		console.log('ProfileEditRoute params', params)
		return null;
	}
});

CERT.ProfileController = Em.ArrayController.extend({
  ready: function()
  {
		console.log("Created App namespace in ProfileController ready");
  }
});

CERT.ProfileView = Em.View.extend(
{
	willInsertElement: function()
	{
		console.log("ProfileView will insert");
	},

	didInsertElement: function()
	{
		console.group('ProfileView didInsertElement');

		console.log('this, this.$()', this, this.$());
		console.log('controller', this.get('controller'));

		//$("body").timeago();

		console.groupEnd();
	}
});

/*CERT.ApplicationController = Ember.Controller.extend();
CERT.ApplicationView = Ember.View.extend({
	templateName: 'application',
	didInsertElement: function()
	{
		console.group('ApplicationView didInsertElement');

		console.log('this, this.$()', this, this.$());

		//$("body").timeago();

		console.groupEnd();
	}
});
*/
/*CERT.RosterView = Em.View.extend({
	templateName: 'roster',
	didInsertElement: function()
	{
		console.group('RosterView didInsertElement');

		console.log('this, this.$()', this, this.$());
		console.log('controller', this.get('controller'));

		//$("body").timeago();

		console.groupEnd();
	}
});*/

CERT.LoadingView = Em.View.extend({
	templateName: 'loading',
	didInsertElement: function()
	{
		console.group('LoadingView didInsertElement');

		console.log('this, this.$()', this, this.$());
		console.log('controller', this.get('controller'));

		//$("body").timeago();

		console.groupEnd();
	}
});
CERT.LoadingController = Em.Controller.extend({
  ready: function()
  {
	console.log("Created App namespace in Loading ready");

  }
});

/*CERT.Router = Ember.Router.extend({
	enableLogging: true,
	root: Ember.Route.extend({
		index: Ember.Route.extend({
			route: '/',
			connectOutlets: function(router)
			{
				console.group('/ connectOutlets');

				console.log('router', router);

				console.groupEnd();
			},
			enter: function(router)
			{
				console.group('/ enter');

				//CERT.hist = CERT.HistoryController.fetch();
				//console.log('hist', CERT.hist);
				//CERT.hist.addObserver('CERT.hist.@each', this, function() { console.log('hist.@each', arguments); })


				console.groupEnd();
			},
			exit: function(router)
			{
				console.log('/ exit');
			}
		}),
		roster: Ember.Route.extend({
			route: '/roster',
			connectOutlets:  function(router, context)
			{
				console.group('/roster connectOutlets');
				console.log('context', context);

				//router.get('applicationController').connectOutlet('rosterMembers', 'loading');
				router.get('applicationController').connectOutlet('rosterMembers', 'roster', CERT.Member.all());

				console.groupEnd();
			}
		})
	})
});

*/

CERT.initialize();