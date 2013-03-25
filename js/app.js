var CERT = Ember.Application.create(
	{
		rootElement: 'body',
		LOG_TRANSITIONS: true,
		enableLogging: true,
		LOG_BINDINGS: true
	}
);

CERT.loggedIn = {};

CERT.Router.map(function() {
  //this.route('team');
  //this.resource('test',{ path: '/abc'});
  //this.route("favorites");
	this.resource("team", { path: "/teams" }, function() {
		this.resource("members", { path: "/members" });
	});
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

CERT.TeamRoute = Ember.Route.extend({
	setupController: function(controller, model)
	{
		console.log("RosterRoute", 'controller', controller, 'model', model);
	},
	model: function(params)
	{
		return CERT.Team.find();
	}
});

CERT.TeamController = Em.ArrayController.extend({
  ready: function()
  {
		console.log("Created App namespace in RosterController ready");
  }
});

CERT.TeamView = Em.View.extend({
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
	},
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