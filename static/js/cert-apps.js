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
			this.resource("teamID", { path: '/:teamID' 	}, function()
			{
				this.resource('event', function()
				{
					this.route("create", { path: '/' });
					this.route("update", { path: '/:eventID' });
				});

				this.resource("roster", function()
				{
					this.route('import');
				});

				this.route('events');
			});
		});
	});

CERTApps.moveUpData = function(data)
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
}

CERTApps.BaseRoute = Ember.Route.extend(
{
	moveUpData: function(data)
	{
		return CERTApps.moveUpData(data);
	}
});

CERTApps.ApplicationRoute = CERTApps.BaseRoute.extend(
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

			CERTApps.Team.lookup().then(function(team){ console.log('model resolved for team', team); model.set('Team', team); console.log('after set model', model); });

			return model; 
		}.bind(this));

		console.groupEnd();

		return t2;
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
		var model = appModel.Member;

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
		var model = null;

		if(params)
		{
			if( params.teamID )
			{
				model = CERTApps.Team.lookup(params.teamID);
			}
		}

		if( !model )
		{
			model = CERTApps.Team.lookup();
		}
	
		model.then(function(model){console.log("resolved TeamRoute model", model);});
		console.groupEnd();

		return model;
	},

});


CERTApps.RosterRoute = CERTApps.BaseRoute.extend(
{
	actions:
	{
	},

	model: function(params)
	{
		console.group("CERTApps.RosterRoute model");

		console.log('params, args', params, arguments);

		var team = this.modelFor('team');

		var settings = 
		{
			url: '/team/roster',
			//type: 'json',
			dataType: 'json',
			method: 'get',
			data: 
			{
				team: team.get('KeyID') || 0
			}
		};

		console.log('getRoster', settings)

		var a = $.ajax(settings);

		var appModel = this.modelFor('application');

		var t = a.then(function(obj){ return this.moveUpData(obj); }.bind(this));
		var t2 = t.then(function(obj) 
		{ 
			console.log('t2 obj', obj, 'appModel', appModel);

			var parsed = Ember.A([]);
			if( obj.Members )
			{
				for( var x = obj.Members.length - 1; x >= 0; x-- )
				{
					var o = obj.Members[x];
					var m = CERTApps.Member.create(o);

					m.loggedInMember = appModel.Member;

					parsed.unshiftObject(m);
				}
			}
			obj.loggedInMember = appModel.Member;
			obj.Members = parsed;

			return obj; 
		}.bind(this));		

		console.groupEnd();

		return t2;
	},

});

CERTApps.BaseObject = Ember.Object.extend(
{
	moveUpData: function(data)
	{
		return CERTApps.moveUpData(data);
	}	
});

CERTApps.AppModel = CERTApps.BaseObject.extend(
{
	Member: null,
	Team: null,
});

CERTApps.Roster = CERTApps.BaseObject.extend();
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

			var parsed = CERTApps.Roster.parseInput(content.RawPasteDataForImport);

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
		console.group('CERTApps RosterImportRoute model')
		var teamModel = this.modelFor('team');
		var appModel = this.modelFor('application');

		teamModel.loggedInMember = appModel.Member;

		teamModel.memberFields = teamModel.loggedInMember.getFields();

		console.log('teamModel', teamModel);
		console.groupEnd();

		return teamModel;
	},

	setupController: function(controller, model)
	{
		console.group('CERTApps RosterImportRoute setupController')

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

/*	model: function(params)
	{
		console.group('CERTApps RosterIndexRoute model')
		var teamModel = this.modelFor('team');
		var appModel = this.modelFor('application');

		teamModel.loggedInMember = appModel.Member;

		console.log('teamModel', teamModel);
		console.groupEnd();

		return teamModel;
	},*/

	setupController: function(controller, model)
	{
		console.group('CERTApps RosterIndexRoute setupController')

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

CERTApps.Member = CERTApps.BaseObject.extend(
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

	emailDisplay: function()
	{
		var retval = this.get('ShowEmail') || this.get('loggedInMember.CanLookup');
	
		//console.log('EmailDisplay', this.get('ShowEmail'), this.get('loggedInMember.CanLookup'), retval, this.get('loggedInMember'));

		return retval;
	}.property('ShowEmail', 'loggedInMember.CanLookup'),

	cellDisplay: function()
	{
		var retval = this.get('ShowCell') || this.get('loggedInMember.CanLookup');
	
		//console.log('CellDisplay', this.get('ShowCell'), this.get('loggedInMember.CanLookup'), retval, this.get('loggedInMember'));
		
		return this.get('ShowCell') || this.get('loggedInMember');
	}.property('ShowCell', 'loggedInMember'),

	line2Display: function()
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

CERTApps.Event = CERTApps.BaseObject.extend(
{
	// eventStartDate: function(key, value, previous)
	// {
	// 	console.group('eventStartDate');
	// 	console.log('arguments', arguments)

	// 	if( arguments.length > 1 )
	// 	{
	// 		var time = this.get('eventStartTime');
	// 		var parsed = this.parseDate(value);
	// 		parsed = parsed + 'T' + time;

	// 		console.log('setting parsed', parsed);
	// 		this.set('EventStart', parsed);
	// 		this.set('privEventStartDate', value);
	// 	}

	// 	console.groupEnd();

	// 	return this.get('privEventStartDate');

	// }.property('EventStart', 'eventStartTime'),

	// eventStartTime: function(key, value, previous)
	// {
	// 	console.group('eventStartTime');
	// 	console.log('arguments', arguments)

	// 	if( arguments.length > 1 )
	// 	{
	// 		if( value && value.length == 4 )
	// 		{
	// 			value = value.substring(0,2) + ':' + value.substring(2);
	// 			console.log('formatted value', value);
	// 		}

	// 		this.set('EventStart', this.get('eventStartDate') + ' ' + value);
	// 		this.set('privEventStartTime', value);
	// 	}

	// 	console.groupEnd();

	// 	return this.get('privEventStartTime');

	// }.property('EventStart', 'eventStartDate'),

	// eventFinishDate: function(key, value, previous)
	// {
	// 	console.group('eventFinishDate');
	// 	console.log('arguments', arguments)

	// 	if( arguments.length > 1 )
	// 	{
	// 		var time = this.get('eventFinishTime');

	// 		var parsed = this.parseDate(value);
	// 		parsed = parsed + 'T' + time;
			
	// 		console.log('setting parsed', parsed);
	// 		this.set('EventFinish', parsed);
	// 		this.set('privEventFinishDate', value);
	// 	}

	// 	console.groupEnd();

	// 	return this.get('privEventFinishDate');

	// }.property('EventFinish', 'eventFinishTime', 'eventStartDate', 'privEventFinsishTime'),

	// eventFinishTime: function(key, value, previous)
	// {
	// 	console.group('eventFinishTime');
	// 	console.log('arguments', arguments)

	// 	if( arguments.length > 1 )
	// 	{
	// 		var finish = this.get('eventFinishDate') || this.get('eventStartDate');

	// 		if( value && value.length == 4 )
	// 		{
	// 			value = value.substring(0,2) + ':' + value.substring(2);
	// 			console.log('formatted value', value);
	// 		}

	// 		this.set('EventFinish', finish + ' ' + value);
	// 		this.set('privEventFinishTime', value);
	// 	}

	// 	console.groupEnd();

	// 	return this.get('privEventFinishTime');

	// }.property('EventEnd', 'eventStartDate'),

	save: function(team)
	{
		console.group('CERTApps.Event save')
		
		this.parseDates();

		console.log('will save event', this);

		var settings = 
		{
			url: '/event/save',
			type: 'json',
			dataType: 'json',
			data: JSON.stringify({ Event: this, Team: team})
		};

		console.log('saving event data', settings)

		var a = $.ajax(settings);
		var t = a.then(function(obj)
			{ 
				var move = this.moveUpData(obj); 

				move.Member = CERTApps.Member.create(move.Member); 

				console.log('App model returning', move); 
				obj = move; 
				return move; 
			}.bind(this));

		a.then(function(obj){ console.log('a.then', JSON.stringify(obj)); }.bind(this));
		t.then(function(obj){ console.log('t.then', JSON.stringify(obj)); }.bind(this));

		console.groupEnd();

		return a;
	},

	parseDates: function()
	{
		this.set('EventStart', this.parseDate(this.get('eventStartDate'), this.get('eventStartTime')));
		
		var finishDate = this.get('eventFinishDate') || this.get('eventStartDate');
		this.set('EventFinish', this.parseDate(finishDate, this.get('eventFinishTime')));

		var meetTime = this.get('arriveTime');
		if( meetTime )
		{
			this.set('MeetTime', this.parseDate(this.get('eventStartDate'), meetTime));
		} else
		{
			this.set('MeetTime', undefined);
		}

	},

	parseDate: function(date, time)
	{
		var retval = null;

		if( date )
		{
			var pieces = date.split('/');
			var dateParsed = '20' + pieces[2] + '-' + pieces[0] + '-' + pieces[1];
			console.log('formatted date', dateParsed);


			var timeParsed = time;
			if( timeParsed && timeParsed.length == 4 )
			{
				timeParsed = timeParsed.substring(0,2) + ':' + timeParsed.substring(2) + ':00';
				console.log('formatted time', timeParsed);
			}		

			var dateObj = new Date();
			offsetHours = dateObj.getTimezoneOffset()/60;
			retval = dateParsed + 'T' + timeParsed + '-0' + offsetHours + ':00';
		}
		return retval;
	},

	parse: function()
	{
		console.group("CERTApps.Event parse")

		var date = new Date(this.EventStart);

		console.log('EventStart', this.EventStart, 'date', date);

		var month = (date.getMonth() + 1).toString();
		var d = date.getDate().toString();
		if( month.length == 1 ) month = '0' + month;
		if( d.length == 1 ) d = '0' + d;

		var startDate =  month + '/' + d + '/' + date.getFullYear().toString().substring(2);
		var startTime = date.getHours() + '' + date.getMinutes();		

		this.set('eventStartDate', startDate);
		this.set('eventStartTime', startTime);

		var date = new Date(this.EventFinish);

		month = (date.getMonth() + 1).toString();
		d = date.getDate().toString();
		if( month.length == 1 ) month = '0' + month;
		if( d.length == 1 ) d = '0' + d;

		var finishDate =  month + '/' + d + '/' + date.getFullYear().toString().substring(2);
		var finishTime = date.getHours() + '' + date.getMinutes();		

		if( startDate !== finishDate )
		{
			this.set('eventFinishDate', finishDate);
		}
		this.set('eventFinishTime', finishTime);

		console.groupEnd();
	},

	parseDatesToJS: function()
	{
		console.group('parseDatesToJS');

		var strDate = this.get('EventStart');
		var date = new Date(strDate);

		console.log('strDate, date', strDate, date);

		this.set('startTimestamp', date.getTime());

		console.groupEnd();
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

		if( ! content.mapping ) content.mapping = Ember.A([]);

		if( matchClass )
		{
			var strIdx = matchClass.substring(2);
			var idx = parseInt(strIdx);

			content.mapping[idx] = this.selection;

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

CERTApps.EventRoute = Ember.Route.extend(
{
	actions:
	{
		saveEvent: function(content)
		{
			console.group('CERTApps.EventRoute saveEvent');
			var event = content.Event;

			var team = this.modelFor('team');
			event.save(team);

			console.groupEnd();
		},

		locationLookup: function(content)
		{
			console.group('CERTApps.EventRoute saveEvent');
			var event = content.Event;

			event.save();

			console.groupEnd();
		}
	},

	model: function(params)
	{
		console.group('CERTApps.EventRoute model')

		var model = { Event: CERTApps.Event.create({ EventLocation: {}, ParkingLocation: {} } )};

		console.groupEnd();

		model.possibleEventTypes = Ember.A(["Deployment", "Exercise", "Meeting", "Training"]);

		return model;
	},

	setupController: function(controller, model)
	{
		console.group('CERTApps.EventRoute setupController')

		controller.set('content', model);

		console.log('controller, model', controller, model);
		console.groupEnd();

		return;		
	}

});

CERTApps.EventCreateRoute = Ember.Route.extend(
{	
	actions:
	{
	},

	model: function(params)
	{
		console.group('CERTApps.EventCreateRoute model')

		var eventModel = this.modelFor('event');

		console.groupEnd();

		return eventModel;
	},

	setupController: function(controller, model)
	{
		console.group('CERTApps.EventCreateRoute setupController')

		controller.set('content', model);

		console.log('controller, model', controller, model);
		console.groupEnd();

		return;		
	}

});

CERTApps.EventCreateView = Ember.View.extend(
{
	templateName: 'event/update'
});

CERTApps.EventUpdateRoute = CERTApps.BaseRoute.extend(
{	
	actions:
	{
	},

	model: function(params)
	{
		console.group('CERTApps.EventUpdateRoute model')

		console.log('params', params);

		var obj = {
			KeyID: parseInt(params.eventID)
		};

		var settings = 
		{
			url: '/event',
			type: 'json',
			dataType: 'json',
			data: JSON.stringify(obj)
		};

		console.log('requesting data', settings)

		var a = $.ajax(settings);
		var t = a.then(function(obj)
			{ 
				console.log('obj', obj);

				var move = this.moveUpData(obj);

				console.log('move', move);

				move.Event = CERTApps.Event.create(move.Event); 

				console.log('EventUpdate model returning', move); 
				obj = move; 
				return move; 
			}.bind(this)
		);

		t.then(function(obj)
		{
			obj.Event.parse();
		});

		console.groupEnd();

		return t;
	},

	setupController: function(controller, model)
	{
		console.group('CERTApps.EventUpdateRoute setupController')

		controller.set('content', model);

		console.log('controller, model', controller, model);
		console.groupEnd();

		return;		
	}

});


CERTApps.TeamIndexRoute = CERTApps.BaseRoute.extend(
{
	// model: function(params)
	// {
	// 	console.group("CERTApps.TeamIndexRoute model");

	// 	console.log('params', params);

	// 	var model = this.modelFor('team');
		
	// 	console.log('model', model);

	// 	console.groupEnd();

	// 	return model;
	// },

	afterModel: function(team, transition) 
	{
		console.group("CERTApps.TeamIndexRoute afterModel");

		console.log('args', arguments);

		this.transitionTo('teamID', team);

		console.groupEnd();
	},

	serialize: function(model)
	{
		console.group("CERTApps.TeamIndexRoute serialize");

	//	console.log('model', model);

		var obj = { teamID: model.get('KeyID') };

		console.groupEnd();

		return obj;
	}
});

CERTApps.TeamIDRoute = CERTApps.BaseRoute.extend(
{
/*	model: function(params)
	{
		console.group("CERTApps.TeamIDRoute model");

		console.log('params', params);

		var model = this.modelFor('team');
		
		console.log('model', model);

		console.groupEnd();

		return model;
	},

	afterModel: function(team, transition) 
	{
		console.group("CERTApps.TeamIDRoute afterModel");

		console.log('args', arguments);

   		//this.transitionTo('teamID', team);

		console.groupEnd();

	}
*/
	serialize: function(model)
	{
		console.group("CERTApps.TeamIDRoute serialize");

		//console.log('model', model);

		var obj = { teamID: model.get('KeyID') };

		console.groupEnd();

		return obj;
	}
});

CERTApps.Team = CERTApps.BaseObject.extend({});

CERTApps.Team.reopenClass(
{
	lookup: function(teamID)
	{
		var settings = 
		{
			url: '/team',
			//type: 'json',
			dataType: 'json',
			method: 'get',
			data: 
			{
				team: teamID || 0
			}
		};

		console.log('get Team', settings)

		var a = $.ajax(settings);

		var t = a.then(function(data)
		{
			var teamData = CERTApps.moveUpData(data).Team;
			var team = CERTApps.Team.create(teamData);
			
			return team;
		});

		return t;
	}
});

CERTApps.TeamIDEventsRoute = CERTApps.BaseRoute.extend(
{	
	actions:
	{
	},

	model: function(params)
	{
		console.group('CERTApps.TeamIDEventsRoute model')
		console.log('params', params);

		var team = this.modelFor('team');

		var obj = {
			KeyID: team.KeyID
		};

		var settings = 
		{
			url: '/events',
			type: 'json',
			dataType: 'json',
			data: JSON.stringify(obj)
		};

		console.log('requesting data', settings)

		var a = $.ajax(settings);
		var t = a.then(
			function(obj)
			{ 
				obj = this.moveUpData(obj);
				console.log('obj', obj);

				var list = Ember.A([]);
				for( var x = obj.Events.length - 1; x >= 0; x-- )
				{
					var e = obj.Events[x];
					var event = CERTApps.Event.create(e); 

					event.parseDatesToJS();

					console.log('adding event to list', event);

					list.pushObject(event)
				}

				var events = CERTApps.Events.create({events: list});

				console.log('TeamIDEventsRoute model returning', events); 
				return events; 
			}.bind(this),
			function(xhr)
			{
				console.error(xhr);
			}.bind(this)
		);

		console.groupEnd();

		return t;
	},

	setupController: function(controller, model)
	{
		console.group('CERTApps.TeamIDEventsRoute setupController')

		controller.set('content', model);

		console.log('controller, model', controller, model);
		console.groupEnd();

		return;		
	}

});


CERTApps.Events = CERTApps.BaseObject.extend({
	events: null,

	init: function()
	{
		events = Ember.A([]);
	},

	upcoming: function()
	{
		var now = Date.now();

		var list = Ember.A([]);

		var events = this.get('events');

		for( var x = events.length - 1; x >= 0; x-- )
		{
			var e = events[x];

			if( e.get('startTimestamp') > now )
			{
				console.log('upcoming because',  e.get('startTimestamp'),' > ', now);
				list.pushObject(e);
			}
		}

		list.sort(this.sortByDate.bind(this));

		return list;

	}.property('events.@each.startTimestamp'),

	recent: function()
	{
		var now = Date.now();

		var list = Ember.A([]);

		var events = this.get('events');

		for( var x = events.length - 1; x >= 0; x-- )
		{
			var e = events[x];

			if( e.get('startTimestamp') < now )
			{
				console.log('recent because',  e.get('startTimestamp'),' < ', now);
				list.pushObject(e);
			}
		}

		list.sort(this.sortByDate.bind(this));

		return list;
	}.property('events.@each.startTimestamp'),

	sortByDate: function(a, b)
	{
		var aStamp = a.get('startTimestamp');
		var bStamp = b.get('startTimestamp');

		if( aStamp > bStamp )
		{
			return 1;
		} else if( bStamp > aStamp) 
		{
			return -1;
		}

		return 0;
	}
});