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

CERTApps.Router.map(function() 
{
	this.resource('landing', { path: '/landing' });
	this.resource('member', function()
		{	
			this.route('update');
		});
	this.resource('team', function()
		{
			this.resource("roster", function()
				{
					this.route("index", { path: '/' });
					this.route("teamID", { path: '/:teamID' });
					this.route('import');
				});
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
			url: '/member',
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


CERTApps.TeamRoute = Ember.Route.extend(
{
	actions:
	{
	},

	model: function(params)
	{
		console.group("CERTApps.TeamRoute model");

		console.log('params, args', params, arguments);

		var appModel = this.modelFor('application');
		//var model = null;
		var team = appModel.data.Team;

		var settings = 
		{
			url: '/team/roster',
			//type: 'json',
			dataType: 'json',
			method: 'get',
			data: 
			{
				team: params.teamID || 0
			}
		};

		console.log('getRoster', settings)

		var model = $.ajax(settings);

		var appModel = this.modelFor('application');

		model.then(function(obj) { 
			//obj.data.Team = team;  

			var parsed = Ember.A([]);
			if( obj.data.Members )
			{
				for( var x = obj.data.Members.length - 1; x >= 0; x-- )
				{
					var o = obj.data.Members[x];
					var m = CERTApps.Member.create(o);

					m.loggedInMember = appModel.data.Member;

					parsed.unshiftObject(m);
				}
			}
			
			obj.data.Members = parsed;

			return obj; 
		}.bind(this));		

		console.groupEnd();

		return model;
	},

});

CERTApps.Roster = Ember.Object.extend();

CERTApps.Roster.reopenClass(
{
	parseInput: function(rawData)
	{
		console.group('CERTApps.Roster.parseInput');

		var lines = rawData.split('\n');
		var lineLen = lines.length;

		var delimiter = '\t';

		if( lines[0].indexOf(delimiter) == -1 )
		{
			delimiter = ','
		}

		console.log('parsing', lineLen, 'lines using delimiter *', delimiter,'*');

		var toImport = Ember.A([]);

		for( var x = 0; x < lineLen; x++ )
		{
			var line = lines[x];
			console.log('line', x, '\t\t', line);

			var cols = line.split(delimiter);
			var colsLen = cols.length;

			var obj = {
				selected: true,
				cols: Ember.A([])
			};

			for( var y = 0; y < colsLen; y++ )
			{
				obj.cols.pushObject(cols[y]);
			}

			//console.debug('obj', JSON.stringify(obj));

			toImport.pushObject(obj);
		}

		console.log('toImport', toImport);

		console.groupEnd();

		return toImport;
	}
});

CERTApps.RosterImportRoute = Ember.Route.extend(
{
	actions:
	{
		parseInput: function(content)
		{
			console.group('CERTApps RosterImportRoute parseInput');

			content = Ember.Object.create(content);
			console.log('content', content);

			var parsed = CERTApps.Roster.parseInput(content.data.RawPasteDataForImport);

			content.set('data.ParsedColumnsForImport', parsed);
			
			console.groupEnd();
		}
	},

	model: function(params)
	{
		console.group('CERTApps TeamRosterRoute model')
		var teamModel = this.modelFor('team');
		var appModel = this.modelFor('application');

		teamModel.data.loggedInMember = appModel.data.Member;

		console.log('teamModel', teamModel);
		console.groupEnd();

		return teamModel;
	},

	setupController: function(controller, model)
	{
		console.group('CERTApps TeamRosterRoute setupController')

		controller.set('content', model);

		console.log('controller, model', controller, model);
		console.groupEnd();

		return;		
	}

});

CERTApps.RosterIndexRoute = Ember.Route.extend(
{
	actions:
	{
	},

	model: function(params)
	{
		console.group('CERTApps TeamRosterRoute model')
		var teamModel = this.modelFor('team');
		var appModel = this.modelFor('application');

		teamModel.data.loggedInMember = appModel.data.Member;

		console.log('teamModel', teamModel);
		console.groupEnd();

		return teamModel;
	},

	setupController: function(controller, model)
	{
		console.group('CERTApps TeamRosterRoute setupController')

		controller.set('content', model);

		console.log('controller, model', controller, model);
		console.groupEnd();

		return;		
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
			url: '/member/save',
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
	},

	EmailDisplay: function()
	{
		return this.get('ShowEmail') || this.get('loggedInMember.CanLookup');
	}.property('ShowEmail', 'loggedInMember'),

	CellDisplay: function()
	{
		return this.get('ShowCell') || this.get('loggedInMember.CanLookup');
	}.property('ShowCell', 'loggedInMember'),

	Line2Display: function()
	{
		var val = ""
		
		var line2 = this.get('Line2')
		if( $.trim(line2).length > 0 )
		{
			val = ", " + line2;
		}

		return val;
	}.property('Line2')


});
Ember.RSVP.configure('onerror', function(error) {
  Ember.Logger.assert(false, error);
});