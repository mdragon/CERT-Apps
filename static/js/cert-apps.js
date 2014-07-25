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
	this.resource('eventDetails', { path: '/event' }, function()
		{
			this.resource('eventID', { path: '/:eventID' }, function()
				{
					this.resource('response',  function()
						{
								this.route('/:responseID');
						});
				});
		});
	// this.resource('response', function()
	// 	{
	// 			this.route('responseID', { path: '/:responseID' }	);
	// 			// , function()
	// 			// 	{
	// 			// 		this.route('/:eventID');
	// 			// 	});
	// 	});
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

CERTApps.TimesObject = CERTApps.BaseObject.extend(
{
	init: function()
	{
		console.group('CERTApps.TimesObject init');
		this._super();
		
		console.groupEnd();
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

	leftZeroPad: function(value)
	{
		if( value.length == 1 ) value = '0' + value;

		return value;
	},

	dateBreakout: function(dateString)
	{
		var date = new Date(dateString);

		var retval = 
		{
			dateObj: date,
			dateString: dateString
		};

		retval.timestamp = date.getTime();

		retval.month = (date.getMonth() + 1).toString();
		retval.date = date.getDate().toString();
		retval.fullYear = date.getFullYear().toString();
		retval.shortYear = retval.fullYear.substring(2);

		retval.month = this.leftZeroPad(retval.month);
		retval.date = this.leftZeroPad(retval.date);

		retval.hours = date.getHours().toString();
		retval.minutes = date.getMinutes().toString();
		retval.seconds = date.getSeconds().toString();

		retval.hours = this.leftZeroPad(retval.hours);
		retval.minutes = this.leftZeroPad(retval.minutes);
		retval.seconds = this.leftZeroPad(retval.seconds);

		retval.prettyDate =  retval.month + '/' + retval.date + '/' + retval.shortYear;
		retval.prettyDateFullYear =  retval.month + '/' + retval.date + '/' + retval.fullYear;
		retval.prettyTime =  retval.hours + '' + retval.minutes

		retval.prettyDateTime = retval.prettyDate + ' ' + retval.prettyTime

		return retval;
	},

	parseDatesToJS: function()
	{
		console.group('parseDatesToJS');

		var strDate = this.get('EventStart');
		var date = new Date(strDate);

		console.log('strDate, date', strDate, date);

		this.set('startTimestamp', date.getTime());

		console.groupEnd();
	},

});

CERTApps.Event = CERTApps.TimesObject.extend(
{
	responses: null,

	init: function()
	{
		console.group('CERTApps.Event init');
		this._super();
		
		this.parse();
		
		console.groupEnd();
	},

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

				console.log('save returning', move); 
				obj = move; 
				return move; 
			}.bind(this));

		a.then(function(obj){ console.log('a.then', JSON.stringify(obj)); }.bind(this));
		t.then(function(obj){ console.log('t.then', JSON.stringify(obj)); }.bind(this));

		console.groupEnd();

		return a;
	},

	sendReminder: function()
	{
		console.group('CERTApps.Event sendReminder')
		
		this.parseDates();

		console.log('will send reminders for  event', this);

		var ev = {KeyID: this.get('KeyID')};

		var settings = 
		{
			url: '/event/reminders/send',
			type: 'json',
			dataType: 'json',
			data: JSON.stringify({ Event: ev, RemindersToSend: { NoResponse: true}})
		};

		console.log('saving event data', settings)

		var a = $.ajax(settings);
		var t = a.then(function(obj)
			{ 
				var move = this.moveUpData(obj); 

				console.log('sendReminder returning', move); 
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

	parse: function()
	{
		console.group("CERTApps.Event parse")

		var es = this.get('EventStart');
		var startDB = this.dateBreakout(es);
		console.log('EventStart', es);

		this.set('eventStartDate', startDB.prettyDate);
		this.set('eventStartTime', startDB.prettyTime);

		var ef = this.get('EventFinish');
		var finishDB = this.dateBreakout(ef);

		if( startDB.prettyDate !== finishDB.prettyDate )
		{
			this.set('eventFinishDate', finishDB.prettyDate);
		}
		this.set('eventFinishTime', finishDB.prettyTIme);

		this.set('startDateBreakout', startDB);
		this.set('finishDateBreakout', finishDB);

		console.groupEnd();
	},

	prettyEventStart: function()
	{
		var es = this.get('EventStart');

		var db = this.dateBreakout(es);

		return db.prettyDateTime;

	}.property('EventStart'),

	responsesYes: function()
	{
		var responses = this.get('responses');

		var retval = responses.filterBy('Attending', true).filterBy('Sure', true);

		return retval;
	}.property('responses.@each'),

	responsesMaybe: function()
	{
		var responses = this.get('responses');

		var retval = responses.filterBy('Attending', true).filterBy('Sure', false);

		return retval;
	}.property('responses.@each'),

	responsesNo: function()
	{
		var responses = this.get('responses');

		var retval = responses.filterBy('Attending', false);

		retval = retval.filterBy('Sure', true);

		return retval;
	}.property('responses.@each'),


	responsesPending: function()
	{
		var responses = this.get('responses');

		var retval = responses.filterBy('Attending', false);

		retval = retval.filterBy('Sure', false);

		return retval;
	}.property('responses.@each'),

	link1Label: function()
	{
		var deploy = this.get('Deployment');
		var retval = "???";

		if( deploy )	
		{
			retval = "IAP";
		} else
		{
			retval = "Agenda";
		}

		return retval;
	}.property('Deployment'),

	link2Label: function()
	{
		var meeting = this.get('Meeting');
		var retval = "???";

		if( meeting )	
		{
		} else
		{
			retval = "Roster";
			
		}

		return retval;
	}.property('Meeting'),

	eventType: function(key, value, previous)
	{
		// setter
		if (arguments.length > 1) 
		{
			this.set("Deployment", false);
			this.set("Exercise", false);
			this.set("Meeting", false);

			this.set(value, true);
		}

		// getter
		var retval = null;
		var match = false;

		if( match == false )
		{
			retval = 'Deployment';
			if( this.get(retval) )
			{
				match = true;
			}
		}

		if( match == false )
		{
			retval = 'Exercise';
			if( this.get(retval) )
			{
				match = true;
			}
		}

		if( match == false )
		{
			retval = 'Meeting';
			if( this.get(retval) )
			{
				match = true;
			}
		}

		return retval;
	}.property('Deployment', 'Exercise', 'Meeting', 'Training'),

	showSecondLink: function()
	{
		var retval = "display: none;";
		if( this.get('Deployment') || this.get('Excercise') )
		{
			retval =  "display: show;";
		}

		return retval;
	}.property('Deployment', 'Exercise', 'Meeting', 'Training'),

	summary: Ember.computed.alias('Summary')
});

CERTApps.Event.reopenClass(
{
	lookup: function(eventID)
	{
		var settings = 
		{
			url: '/eventA',
			//type: 'json',
			dataType: 'json',
			method: 'post',
			data: JSON.stringify( 
				{
					KeyID: parseInt(eventID || 0)
				}
			)
		};

		console.log('get Event', settings)

		var a = $.ajax(settings);

		var t = a.then(function(data)
		{
			var eventData = CERTApps.moveUpData(data).Event;
			var ev = CERTApps.Event.create(eventData);
			
			console.log('returning Event', ev);

			return ev;
		});

		return t;
	}
});


Ember.RSVP.configure('onerror', function(error) {
	Ember.Logger.assert(false, error);
});

CERTApps.ImportColumnSelect = Ember.Select.extend({
	change: function (ev) 
	{
		console.group('CERTApps.ImportColumnSelect change')
		
		console.log('this, ev, args', this, ev, arguments);
		
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
			var ev = content.Event;

			var team = this.modelFor('team');
			ev.save(team);

			console.groupEnd();
		},

		locationLookup: function(content)
		{
			console.group('CERTApps.EventRoute saveEvent');
			var ev = content.Event;

			ev.save();

			console.groupEnd();
		},

		sendReminder: function(content)
		{
			console.group('CERTApps.EventRoute saveEvent');
			
			console.log('content', content);

			var ev = content.Event;

			ev.sendReminder();

			console.groupEnd();
		}
	},

	model: function(params)
	{
		console.group('CERTApps.EventRoute model')

		console.log('params', params);

		var model = { Event: CERTApps.Event.create({ EventLocation: {}, ParkingLocation: {} } )};

		console.groupEnd();

		return model;
	},

	setupController: function(controller, model)
	{
		console.group('CERTApps.EventRoute setupController')


		controller.set('content', model);

		console.log('controller, model', controller, model);
		console.groupEnd();

		return;		
	},

	serialize: function(model)
	{
		console.group("CERTApps.EventRoute serialize");

	//	console.log('model', model);
		debugger;
		var obj = { teamID: model.get('KeyID') };

		console.groupEnd();

		return obj;
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

		model.possibleEventTypes = Ember.A(["Deployment", "Exercise", "Meeting", "Training"]);

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
		});

		console.groupEnd();

		return t;
	},

	setupController: function(controller, model)
	{
		console.group('CERTApps.EventUpdateRoute setupController')

		controller.set('content', model);

		model.possibleEventTypes = Ember.A(["Deployment", "Exercise", "Meeting", "Training"]);

		console.log('controller, model', controller, model);
		console.groupEnd();

		return;		
	},

	serialize: function(model)
	{
		console.group("CERTApps.EventUpdateRoute serialize");

		var ev = model;
		if( ev.Event ) ev = ev.Event;

		var obj = { eventID: ev.get('KeyID') };

		console.groupEnd();

		return obj;
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

CERTApps.MemberEvent = CERTApps.TimesObject.extend(
{
	value:  null,
	arriveTime: null,
	departTime: null,

	init: function()
	{
		console.group('CERTApps.MemberEvent init');
		this._super();
		
		this.parse();
		
		console.groupEnd();
	},

	defaults: function(ev)
	{
		console.group('CERTApps.MemberEvent defaults')

		console.log('MemberEvent', this);
		console.log('ev', ev);

		var startDB = this.dateBreakout(ev.EventStart);
		var finishDB = this.dateBreakout(ev.EventFinish);


		if( ! this.arriveTime )
		{
			if( startDB )
			{
				this.set('arriveTime', startDB.prettyTime);
			}
		}

		if( ! this.departTime )
		{
			if( finishDB )
			{
				this.set('departTime', finishDB.prettyTime);
			}
		}

		console.groupEnd();
	},

	parse: function()
	{
		console.group('CERTApps.MemberEvent parse');

		var arrive = this.get('Arrive');
		var depart = this.get('Depart');
		if( arrive )
		{
			var arriveDB = this.dateBreakout(arrive);
			console.log('arriveDB', arriveDB)
			this.set('arriveTime', arriveDB.prettyTime);
		}

		if( depart )
		{
			var departDB = this.dateBreakout(depart);
		}



		console.groupEnd();
	},

	save: function(ev)
	{
		console.group('CERTApps.MemberEvent save')

		this.parseInputs(ev);

		var obj = 
			{
				Event: {KeyID: ev.KeyID},
				Response: this
			}

		var settings = 
			{
				url: '/response/save',
				type: 'json',
				dataType: 'json',
				data: JSON.stringify(obj),
				method: 'POST'
			};

		console.log('requesting data', settings)

		var a = $.ajax(settings);
		var t = a.then(function(obj)
		{ 
			var obj = this.moveUpData(obj); 

			return obj;
		}.bind(this));

		t.then(function(obj)
		{ 
			return obj;
		});

		console.groupEnd();

		return t;
	},

	parseInputs: function(ev)
	{
		console.group('CERTApps.MemberEvent parseInputs')

		console.log('this, ev', this, ev);

		switch(this.get('value'))
		{
			case 'Yes':
				this.set('Attending', true);
				this.set('Sure', true)
				break;
			case 'Maybe':
				this.set('Attending', true);
				this.set('Sure', false);
				break;
			case 'No':
				this.set('Attending', false);
				this.set('Sure', true);
				break;
		}

		var eventStart = ev.get('EventStart');
		var startDB = this.dateBreakout(eventStart);
		var arriveT = this.get('arriveTime');
		arriveT = arriveT.substring(0,2) + ':' + arriveT.substring(2) + ':00';
		var arriveStamp = eventStart.replace(startDB.hours + ':' + startDB.minutes + ':00', arriveT);

		var eventFinish = ev.get('EventFinish');
		var departT = this.get('departTime');
		departT = departT.substring(0,2) + ':' + departT.substring(2) + ':00';
		var departStamp = eventStart.replace(startDB.hours + ':' + startDB.minutes + ':00', departT);

		this.set('Arrive', arriveStamp);
		this.set('Depart', departStamp);

		console.groupEnd();
	}
});

CERTApps.MemberEvent.reopenClass(
{
	lookupByEvent: function(eventID)
	{
		console.group("lookupByEvent");

		var settings = 
		{
			url: '/response',
			//type: 'json',
			dataType: 'json',
			method: 'get',
			data: 
			{
				event: eventID || 0
			}
		};

		console.log('get Response', settings)

		var a = $.ajax(settings);

		var t = a.then(function(data)
		{
			var dataObj = CERTApps.moveUpData(data).Response;
			var retval = CERTApps.MemberEvent.create(dataObj);
			
			return retval;
		});

		console.groupEnd();

		return t;
	}
});

CERTApps.TeamIDEventsRoute = CERTApps.BaseRoute.extend(
{	
	actions:
	{
		userResponseForm: function(ev)
		{
			console.group('CERTApps.TeamIDEventsRoute.actions userResposne');

			console.log('ev', ev);

			this.transitionTo('response', {Event: ev});

			console.groupEnd();
		},

		eventEdit: function(ev)
		{
			console.group('CERTApps.TeamIDEventsRoute.actions eventEdit');

			console.log('ev', ev);
			var team = this.modelFor('team');

			this.transitionTo('event.update', team, {Event: ev});

			console.groupEnd();
		}

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

				var events = this.parseEventsData(obj);

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
	},

	serialize: function(model)
	{
		console.group("CERTApps.TeamIDEventsRoute serialize");

		console.log('model', model);

		var e = model;
		if( model.Event ) e = model.Event;

		var obj = { eventID: e.get('KeyID') };

		console.groupEnd();

		return obj;
	},

	parseEventsData: function(obj)
	{
		console.group('CERTApps.TeamIDEventsRoute parseEventsData');
		console.log('obj', obj);

		var list = Ember.A([]);
		for( var x = obj.Events.length - 1; x >= 0; x-- )
		{
			var e = obj.Events[x];
			var ev = CERTApps.Event.create(e); 

			ev.parseDatesToJS();

			console.log('adding event to list', ev);

			list.pushObject(ev)
		}

		var events = CERTApps.Events.create({events: list});

		console.log('TeamIDEventsRoute model returning', events); 

		this.parseResponses(events, obj.Responses);

		console.log('events after parseResponses', events);

		console.groupEnd();

		return events; 
	},

	parseResponses: function(events, responses)
	{
		console.group('CERTApps.TeamIDEventsRoute parseResponses');
		console.log('events, responses', events, responses);

		if( responses )
		{

			var lookup = {};

			events.events.forEach( 
				function(e)
				{
					var eventID = e.Key;
					lookup[eventID] = e;
					e.responses = Ember.A([]);

					console.log('added to lookup key', eventID, e);
				}.bind(this)
			);
			
			responses.forEach(
				function(r)
				{
					var eventID = r.EventKey;
					var ev = lookup[eventID];

					console.log('found event using key', eventID, ev);
					if( ev ) 
					{
						ev.responses.pushObject(r);
					} else
					{
						console.warn('No event found for Key from Response', eventID, r);
					}
				}.bind(this)
			);
		} else 
		{
			console.log('no responses found to parse');
		}

		console.groupEnd();
	}
});

CERTApps.ResponseRoute = CERTApps.BaseRoute.extend(
{
	actions:
	{
		submitResponse: function(response)
		{
			console.group('CERTApps.ResponseRoute.actions submitResponse')

			var ev = this.modelFor('eventID');

			if( ev.Event ) ev = ev.Event;

			console.log('response, ev', response, ev);

			var respObj = CERTApps.MemberEvent.create(response);

			var t = respObj.save(ev);

			console.groupEnd();

			return t;
		}
	},

	model: function(params, transition)
	{
		console.group("CERTApps.EventIDResponseRoute model");

		console.log('params, args', params, arguments);

		var e = this.modelFor('eventID');

		if( ! e )
		{
			log.warn('Doing bad things using transition to get eventID model because modelFor returned null');

			if( transition )
			{
				e = transition.get('resolvedModels.eventID');
			} else
			{
				log.warn('transition was null');
			}
		}

		if( e.Event ) e = e.Event;
		console.log('e', e);


		var a = CERTApps.MemberEvent.lookupByEvent(e.KeyID).then(function(data){ data = this.moveUpData(data); return data; }.bind(this));

		console.groupEnd();

		return a;
	},

	setupController: function(controller, model)
	{
		console.group('CERTApps.EventIDResponseRoute setupController')
		console.log('controller, model', controller, model);

		//if( model.Event ) model = model.Event;

		if( ! model.Response ) model = {Response: CERTApps.MemberEvent.create()};
		var ev = this.modelFor('eventID');
		if( ev.Event ) ev = ev.Event;

		model.Response.defaults(ev);

		controller.set('content', model);

		console.groupEnd();

		return;		
	},

	serialize: function(model)
	{
		console.group("CERTApps.EventIDResponseRoute serialize");

		console.log('model', model);

		var e = model;
		if( model.Event ) e = model.Event;

		var obj = { eventID: e.get('KeyID') };

		console.groupEnd();

		return obj;
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
				//console.log('upcoming because',  e.get('startTimestamp'),' > ', now);
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

CERTApps.EventIDRoute = CERTApps.BaseRoute.extend(
{
	afterModel: function(model)
	{
		console.log('afterModel', model, arguments)
		if( model.Event )
		{
		}
	},

	model: function(params)
	{
		var m = CERTApps.Event.lookup(params.eventID);

		var model = m.then(
		function(data)
			{
				console.log('m.then', data);

				data = this.moveUpData(data);

				var model = {Event: CERTApps.Event.create(data)};

				return model;
			}.bind(this)
		);

		return model;
	},

	serialize: function(model)
	{
		console.group("CERTApps.EventIDRoute serialize");

		console.log('model', model);
		var m = model;
		if( model.Event ) m = model.Event;

		var obj = { eventID: m.get('KeyID') };

		console.groupEnd();

		return obj;
	}
});

CERTApps.RadioButton = Ember.View.extend({  
	tagName : "input",
	type : "radio",
	attributeBindings : [ "name", "type", "value", "checked:checked:" ],
	click : function() {
		console.group('CERTApps.RadioButton click');

		var val = this.$().val();
		this.set("selection", val)

		console.groupEnd();
	},
	checked : function() {
    	return this.get("value") == this.get("selection");   
	}.property()
});

CERTApps.ResponseRadioButton = CERTApps.RadioButton.extend({  
	tagName : "input",
	type : "radio",
	attributeBindings : [ "name", "type", "value", "checked:checked:" ],
	click : function() {
		console.group('CERTApps.ResponseRadioButton click');

		this._super();

		var $ = this.$();

		var val = $.val();
		var form$ = $.closest('form');
		var details$ = form$.find('div.details');
		var button$ = form$.find('div.submit-response');
		var arrive$ = Ember.$(details$.find('input[name=arriveTime]')[0]);

		console.log('val, this, $', val, this, $);
		if( val === "Yes" || val === "Maybe") 
		{
			details$.slideDown();
			arrive$.focus();

		} else
		{
			details$.slideUp();
		}

		if( button$.find(":visible").length === 0 ) button$.slideDown()

		console.groupEnd();

		return true;
    },
});