var CERT = Em.Application.create(
	{
		rootElement: '#content'
	}
);

CERT.ApplicationController = Ember.Controller.extend();
CERT.ApplicationView = Ember.View.extend({
	templateName: 'layout',
	didInsertElement: function()
	{
		console.group('ApplicationView didInsertElement');

		console.log('this, this.$()', this, this.$());

		//$("body").timeago();

		console.groupEnd();
	}
});

CERT.RosterView = Em.View.extend({
	templateName: 'roster',
	didInsertElement: function()
	{
		console.group('RosterView didInsertElement');

		console.log('this, this.$()', this, this.$());
		console.log('controller', this.get('controller'));

		//$("body").timeago();

		console.groupEnd();
	}
});
CERT.RosterController = Em.ArrayController.extend({
  ready: function()
  {
	console.log("Created App namespace in RosterController ready");

  }
});

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
CERT.LoadingController = Em.ArrayController.extend({
  ready: function()
  {
	console.log("Created App namespace in Loading ready");

  }
});

CERT.Router = Ember.Router.extend({
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

CERT.Member = Ember.Object.extend();
CERT.Member.reopenClass({
  _listOfMembers:  Em.A(),
  all:  function(){
	console.log('CERT.Member.all');
	var allShoes = this._listOfMembers;
	// Mock an ajax call; like a jQuery.ajax might have done...
	setTimeout( function(){
		console.log('CERT.Member.success');
	allShoes.clear();
	allShoes.pushObjects(
		[
		{ id: 'rainbow',   name: "Rainbow Sandals",
			price: '$60.00', description: 'San Clemente style' },
		{ id: 'strappy',   name: "Strappy shoes",
			price: '$300.00', description: 'I heard Pénèlope Cruz say this word once.' },
		{ id: 'bluesuede', name: "Blue Suede",
			price: '$125.00', description: 'The King would never lie:  TKOB⚡!' }
		]
	);
	}, 2000);
	return this._listOfMembers;
  }
});

Ember.enableLogging = true;
Ember.debug = true;

CERT.initialize();