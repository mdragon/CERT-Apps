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
		a.then(function(obj)
			{ 
				var move = this.moveUpData(obj); 

				move.Member = CERTApps.Member.create(move.Member); 

				console.log('App model returning', move); 
				obj = move; 
				return move; 
			}.bind(this));

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

			var obj = Ember.Object.create({
				selected: true,
				cols: Ember.A([]),
				idx: x
			});

			for( var y = 0; y < colsLen; y++ )
			{
				obj.cols.pushObject({name: '_c' + y.toString(), value: cols[y]});
			}

			//console.debug('obj', JSON.stringify(obj));

			toImport.pushObject(obj);
		}

		console.log('toImport', toImport);

		console.groupEnd();

		return toImport;
	},

	saveImport: function(toImport)
	{
		console.group('CERTApps.Roster.saveImport');

		var mapping = toImport.mapping;
		var data = toImport.ParsedColumnsForImport;
		var teamID = toImport.Team.KeyID;

		console.log('mapping', mapping);
		console.log('data', data);

		var members = this.parseMembers(mapping, data);

		this.saveMembers(members, teamID);
	},

	parseMembers: function(mapping, data)
	{
		console.group("CERTApps.Roster.parseMembers");

		var members = Ember.A([]);
		var dataLen = data.length;
		for( var r = 0; r < dataLen; r++ )
		{
			var row = data[r];

			if( row.selected )
			{

				var m = CERTApps.Member.create();
				for( var c = mapping.length - 1; c >= 0; c-- )
				{
					var col = mapping[c];
					var val = row.cols[c].value;
					//console.log('r, c, col, val', r, c, col, val);

					if( col != undefined )
					{
						m.set(col, val);
					}
				}

				console.log('m', m);
				members.pushObject(m);
			}
		}

		console.log('members', members);
		console.groupEnd();

		return members;
	},

	saveMembers: function(members, teamID)
	{
		console.group('CERTApps.Roster.saveMembers');

		var obj = 
		{
			TeamID: teamID,
			Members: members
		}

		var settings = 
		{
			url: '/team/roster/import',
			type: 'json',
			dataType: 'json',
			method: 'post',
			data: JSON.stringify(obj) + "\r\n"
		};

		console.log('roster import request', settings)

		var a = $.ajax(settings);
		a.then(function(obj){ console.log('roster import', obj); }.bind(this));		

		console.groupEnd();
	},

	toggle: function(row, imports, up)
	{
		console.group('CERTApps.Roster.toggle');
		console.log('row, up', row, up);

		var start = 0;
		var limit = row.idx;

		if( up == false )
		{
			start = row.idx;
			limit = imports.length - 1;
		}

		console.log('start, limit', start, limit);

		for( var x = start; x <= limit; x++ )
		{
			var r = imports[x];

			//console.log('toggling', r);
			r.toggleProperty('selected');
		}

		console.groupEnd();
	}
});

CERTApps.RosterImportRoute = Ember.Route.extend(
{
	actions:
	{
		parseInput: function(content)
		{
			console.group('CERTApps RosterImportRoute actions parseInput');

			content = Ember.Object.create(content);
			console.log('content', content);

			var parsed = CERTApps.Roster.parseInput(content.data.RawPasteDataForImport);

			content.set('data.ParsedColumnsForImport', parsed);

			if( parsed.length > 0 )
			{
				content.set('data.importCols', parsed[0].cols);
			}
			
			console.groupEnd();
		},

		save: function(content)
		{
			console.group('CERTApps RosterImportRoute actions save');

			content = Ember.Object.create(content);
			console.log('content', content);

			CERTApps.Roster.saveImport(content.data);

			console.groupEnd();

		},

		toggleUp: function(row, imports)
		{
			console.group('CERTApps RosterImportRoute actions toggleUp');

			//console.log('row, imports', row, imports);

			CERTApps.Roster.toggle(row, imports, true);

			console.groupEnd();

		},

		toggleDown: function(row, imports)
		{
			console.group('CERTApps RosterImportRoute actions toggleUp');

			//console.log('row, imports', row, imports);

			CERTApps.Roster.toggle(row, imports, false);

			console.groupEnd();

		}

	},

	model: function(params)
	{
		console.group('CERTApps TeamRosterRoute model')
		var teamModel = this.modelFor('team');
		var appModel = this.modelFor('application');

		teamModel.data.loggedInMember = appModel.data.Member;

		teamModel.data.memberFields = teamModel.data.loggedInMember.getFields();

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
	}.property('Line2'),

	getFields: function()
	{
		console.group('getFields');

		var fields = Ember.A([]);
		var include = Ember.A(["ShowCell","ShowEmail","OKToText","RadioID","Town","OEM","Officer","Active","FirstName","LastName","Cell","HomePhone","Email","Email2","Line1","Line2","City","State","Zip"]);

		for( f in this )
		{
			//console.log('f', f);
			if( this.hasOwnProperty(f) )
			{
				//console.log('this, has', f);

				if( include.contains(f) )
				{
					fields.pushObject(f);	
				} else
				{
					console.log('skipping', f);
				}
			}
		}

		fields.sort();

		console.log('fields', JSON.stringify(fields));

		console.groupEnd();

		return fields;
	}
});

Ember.RSVP.configure('onerror', function(error) {
  Ember.Logger.assert(false, error);
});

CERTApps.ImportColumnSelect = Ember.Select.extend({
	change: function (event) 
	{
		console.group('CERTApps.ImportColumnSelect change')
		
		console.log('this, event, args', this, event, arguments);
		
		var element = $('#' + this.elementId);
		//console.log('loading this.elementId, element', this.elementId, element);

		var th = element.closest('th');
		var thead = th.closest('thead');
		var table = thead.closest('table');
		var tbody = table.find('tbody');

		//console.log('thead', thead);

		var selects = thead.find('select.import-column');

		for( var x =  selects.length - 1; x >= 0; x-- )
		{
			var s = selects[x];

			//console.log('s', s.id, s.value);

			if( s.selectedIndex > 0 && s.value == this.selection )
			{
				if( this.elementId != s.id )
				{
					console.log('reseting prior selected column', s.id, s.value);
					s.selectedIndex = 0;

					var $s = $(s);
					var innerTh = $s.closest('th');
					var matchClass = this.getMatchClass(innerTh);

					if( matchClass )
					{
						var tds = tbody.find('td.' + matchClass);

						console.log('tds that match, removing', tds.length, matchClass);

						tds.removeClass('mapped');
					}
				} else
				{
					console.log('no reseting because this is the column we just selected')
				}
			}
		}

		var matchClass = this.getMatchClass(th);

		var content = this.get('controller.model');

		if( ! content.data.mapping ) content.data.mapping = Ember.A([]);

		if( matchClass )
		{
			var strIdx = matchClass.substring(2);
			var idx = parseInt(strIdx);

			content.data.mapping[idx] = this.selection;

			console.log('matchClass, strIdx, idx', matchClass, strIdx, idx);

			var tds = tbody.find('td.' + matchClass);

			console.log('tds that match, adding', tds.length, matchClass);

			tds.addClass('mapped');
		}

		console.groupEnd();
	},

	getMatchClass: function(element)
	{
		console.group('getMatchClass');

		var match = null;
		var classList = element.attr('class').split(' ');

		console.log('checking classList from element', classList, element);
		for( var x = classList.length - 1; x >= 0; x-- )
		{
			var c = classList[x];

			console.log('checking', c, c[0]);

			if( c[0] == '_' )
			{
				match = c;
				break;
			}
		}

		console.log('match', match);

		console.groupEnd();

		return match;
	}
});