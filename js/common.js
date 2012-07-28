cert = {};
cert.editable = {};
cert.editable.save = {};
cert.editable.ui = {};
cert.editable.ui.contentEditable = false;
cert.editable.ui.clickToEdit = true;
cert.editable.changes = {};

cert.member = {};
cert.member.ui = {};

cert.showHide = {};
cert.showHide.ui = {};
cert.showHide.ui.type = '';

cert.roster = {};
cert.roster.load = {};

cert.template = {};
cert.template.ui = {};
cert.template.cache = {};

cert.loggedIn = {};

var console = console || {};
console.log = console.log || function(){};
console.group = console.group || function(){};
console.groupEnd = console.groupEnd || function(){};
console.warn = console.warn || function(){};
console.error = console.error || function(){};

//window.onerror = function(){ alert('error');};

cert.template.ui.get = function(name)
{
	var cached = cert.template.cache[name];
	if( ! cached )
	{
		if( ! cert.template.div ) cert.template.div = $('#templates');
		
		cached = cert.template.div.children('.' + name);
		
		if( ! cached || cached.length === 0 ) 
		{
			cached = cert.template.div.children('table').find('tr.' + name);
		}
		cert.template.cache[name] = cached;
		console.log('cache miss on name', name, 'found element', cached);
	}
	
	return cached.clone();
};

cert.template.ui.populate = function(template, data, replaceNull)
{
	replaceNull = replaceNull || true;
	for( var p in data )
	{
		if( data.hasOwnProperty(p) )
		{
			if( ! data[p] ) 
			{
				console.warn('bad/missing data[p]', p, data[p], 'replace with empty string?', replaceNull);
			}
		
			if( data[p] && data[p].jquery )
			{
				console.log('it is a jquery object, not a js object');
			}
		
			if( replaceNull || typeof(data[p]) !== "object" || data[p] && data[p].jquery ) 
			{
				var pattern = '\{\{' + p + '\}\}';
				var regex = new RegExp(pattern, 'g');
				
				// debugging
				// if( unescape(template.html()).match(regex) )
				// {
					// console.log('match', p);
				// } else
				// {
					// console.log('miss', p, unescape(template.html()));
				// }	
				
				if( data[p] && data[p].jquery )
				{
					// it's actually an jquery object with html or whatever elements inside it
					// first give ourselves somewhere to latch on to
					template.html(unescape(template.html()).replace(regex, '<span class="injected-' + p +'"/>'));
					// then inject the jquery obj
					var injected = template.find('.injected-' + p);
					injected.append(data[p]);
					
					console.log('injected', injected, 'data[p]', data[p]);
				} else
				{
					// normal non-js object it's a string or an int or something
					var val = data[p];
					if( ! data[p] && replaceNull ) 
					{
						val = '';
					}
					
					template.html(unescape(template.html()).replace(regex, val));
				}
			} else
			{
				//console.log("skip p", p);
			}
		}
	}
};

cert.template.ui.getPopulated = function(name, data)
{
	var template = cert.template.ui.get(name);
	cert.template.ui.populate(template, data);
	
	return template;
};

cert.showHide.ui.show = function(el, callback)
{
	el.show(0, null, callback);
};

cert.showHide.ui.hide = function(el, callback)
{
	el.hide(0, null, callback);
};

cert.editable.save.buildChanges = function(e, key)
{
	console.group('cert.editable.save.buildChanges');
	console.log('e', e, 'key', key);
	var inputs = $('form input');
	key = key || e.data.key;
	if( ! key )
	{
		var keyField = inputs.filter('[name="key"]');
		key = keyField.val();
		console.log('key was empty, found', keyField);
	}
	console.log('inputs', inputs, 'key', key);
	
	for( var x = inputs.length - 1; x >= 0; x-- )
	{
		var i = $(inputs[x]);
		
		if( i.attr('type') !== 'hidden' )
		{
			var inputSpan = i.closest('.input');
			var inline = inputSpan.prev('.inline');
			var prior = inline.data('valuePrior');
			var val = i.val();
			
			console.log('x', x, 'i', i, i.val(), 'prior', prior, 'inline', inline.data(), inline);
			var e = { data: { name: i.attr('name'), value: { prior: prior, org: inline.data('valueOriginal'), } } };
			cert.editable.ui.trackChanges(i, inline, e, key)
		}
		else
		{
			console.log('skip hidden field', i);
		}
	}
	
	console.groupEnd();
	
	return cert.editable.changes[key];
};

cert.editable.save.buildJSON = function(changes, e, key)
{
	var obj = {};
	var target = $(e.currentTarget);
	var form = target.closest('form');
	
	if( ! key )
	{
		var keyField = form.find('input[name="key"]');
		key = keyField.val();
		console.log('key was empty, found', keyField);
	}
	
	console.log('target, form, key', target, form, key);
	
	obj.key = key;
	obj.changes = changes;
	
	console.log('obj', obj);

	return JSON.stringify(obj);
};

cert.editable.save.success =  function(data, status, xhr, spinner)
{
	console.group('cert.editable.save.success');
	
	window.setTimeout(function(){ cert.showHide.ui.hide(spinner); }, 200);	
	cert.notify.success.add(' ', 'Record saved');
		
	console.log('data', data, status, xhr);
	
	var keys = $('form input[name="key"]');
	var form = null;
	
	for( var x = keys.length - 1; x >= 0; x-- )
	{
		var key = $(keys[x]);
		console.log('x', x, 'key', key, key.val(), data.key);
		
		if( key.val() === data.key )
		{
			form = key.closest('form');
			var changes = cert.editable.changes[data.key];
			if( changes )
			{
				for( var y = changes.length - 1; y >= 0; y-- )
				{
					var c = changes[y];
					console.log('y', y, 'c', c);
					
					for( var z = data.changes.length - 1; z >= 0; z-- )
					{
						var dc = data.changes[z];
						console.log('checking if changes match', c.stamp, dc.stamp, c, dc);
						if ( c.stamp === dc.stamp )
						{
							console.log('remove matching change', c);
							cert.editable.changes[data.key].splice(y,1);
						}
					}
				}
			}
		}
	}

	if( form )
	{
		for( var z = data.changes.length - 1; z >= 0; z-- )
		{
			var dc = data.changes[z];
			console.log('updating field from change', dc.name, dc);
			
			var inline = form.find('.' + dc.name +'-content');
			if ( inline.length > 0 )
			{
				console.log('updating prior value', inline.data('valuePrior'), c.value);
				inline.data('valuePrior', c.value);
			} else
			{
				console.warn('inline not found', inline.selector);
			}
		}
	}

	if( cert.editable.ui.clickToEdit ) cert.editable.ui.checkSave(cert.editable.changes[key], form);
		
	console.groupEnd();
};

cert.editable.save.error = function(error, status, xhr)
{
	cert.notify.error.add('Unexpected error', 'There has been an unexpected Error when handling your request.  Please refresh the page and try again.');
}

cert.editable.save.event = function(e)
{
	console.group('cert.editable.save.event');
	
	e.preventDefault();

	var changes = null;
	var target = $(e.currentTarget);
	var form = target.closest('form');
	var keyField = form.find('input[name="key"]');
	var key = keyField.val();
	
	if( ! cert.editable.ui.clickToEdit )
	{
		changes = cert.editable.save.buildChanges(e, key);
 	}  else
	{
		changes = cert.editable.save.flattenChanges(cert.editable.changes[key]);		
	}
	
	console.log('changes', changes);

	var json = cert.editable.save.buildJSON(changes, e, key);
	console.log('json', json);

	if( changes.length == 0 ) 
	{
		console.warn('No changes found', changes, 'json', json);
		cert.notify.warn.add('No changes to the data were detected so nothing was saved', 'No changes found');
		return;
	}
	
	var target = $(e.currentTarget);
	var spinner = target.next('.spinner');
	cert.showHide.ui.show(spinner);
	
	console.log('target', target, 'spinner', spinner);
		
	var settings =
	{
		url: '/profile/save',
		data: {json: json},
		cache: false,
		type: 'POST',
		success: function(data, status, xhr){ cert.editable.save.success(data, status, xhr, spinner); },
		error: cert.editable.save.error,
	}
	
	console.log('settings', settings);
	
	$.ajax(settings);
	
	console.groupEnd();
};

cert.editable.save.flattenChanges = function(changes)
{
	console.group('cert.editable.save.flattenChanges');
	console.log('changes in', changes);

	var list = {};
	
	var cnt = changes.length;
	for( var x = cnt - 1 ; x >= 0 ; x-- )
	{
		var c = changes[x];
		var c2 = list[c.name];
		
		console.log('x', x, 'c', c, 'c2', c2);
		
		if( c2 )
		{
			console.log('later change already tracked remove', x, c);
			changes.splice(x,1);
		} else
		{
			list[c.name] = c;
		}
	}
	
	var flat = [];
	
	for( var c in list )
	{
		console.log('c', c, list[c]);
		flat.push(list[c]);
	}
	
	console.groupEnd();	
	
	return flat;
}

cert.editable.ui.checkSave = function(changes, form)
{
	console.group('cert.editable.ui.checkSave');
	
	if( ! changes )
	{
		return;
	}
	
	if( ! $.isArray(changes) )
	{
		console.error('changes is not an array, it may need to be indexed by key', changes);
	}
	
	var saveBtn = form.find('.save');
	var explain = saveBtn.next('.explain');
	
	console.log('saveBtn', saveBtn, 'explain', explain);
	
	if( (changes && changes.length > 0) || cert.editable.ui.clickToEdit == false )
	{
		saveBtn.attr('disabled', false);
		cert.showHide.ui.hide(explain);
		console.log('enable save');
	} else
	{
		saveBtn.attr('disabled', true);
		cert.showHide.ui.show(explain);
		console.log('disable save');
	}
	
	console.groupEnd();
};

cert.editable.ui.leave = function (e, keepChange)
{
	console.group('cert.editable.ui.leave');
	console.log('e', e, 'keepChange', keepChange);
	e.stopPropagation();
	e.preventDefault();
	if( ! cert.editable.ui.contentEditable )
	{
		var target = $(e.currentTarget);
		var parent = target.closest('span');
		var form = parent.closest('form');
		var inline = parent.prev('span');
		
		console.log('target, parent, form, inline', target, parent, form, inline);

		keepChange = e.data.acceptChange || keepChange;
		
		if( ! keepChange ) 
		{
			console.log('throwing away change, calling stopImmediate');
			e.stopImmediatePropagation();
		}
		
		console.log('keepChange, e.data.jumpToNext', keepChange, e.data.jumpToNext);
		
		// not any more now we remove the blur bind so there is no blur that fires when you chose to leave && e.data.jumpToNext === false 
		// if we jumpToNext it will track the change when that click fires
		if( keepChange ) 
		{
			cert.editable.ui.trackChanges(target, inline, e);
		}
		
		parent.hide(cert.showHide.ui.type);
		inline.show(cert.showHide.ui.type);
		
		if( e.data.jumpToNext )
		{
			var next = $(inline.nextAll('.inline')[0]);
			console.log('next', next);
			// wait a little bit to let the UI catch up
			window.setTimeout(function(){ console.log('trigger click'); next.trigger('click', e); }, 200);
		}
		
		target.unbind();
	} else
	{
		document.designMode = 'off';	
	}
	
	cert.editable.ui.checkSave(cert.editable.changes[e.data.key], form);
	
	console.groupEnd();
};

cert.editable.ui.blur = function (e)
{
	console.group('cert.editable.ui.blur');
	// lame that we're hooking the blur event
	
	e.data.jumpToNext = false;
	e.data.acceptChange = true;
	
	console.log('blur e', e);
	cert.editable.ui.leave(e, true);
	
	console.groupEnd();
};

cert.editable.ui.trackChanges = function(target, inline, e, key)
{
	console.group('cert.editable.ui.trackChanges');
	console.log('target, inline', target, inline, 'e', e, 'key', key);
	
	var prior = e.data.value.prior;// || inline.data('value-prior');
	var org = e.data.value.original;// || inline.data('value-original');
	var val = $.trim(target.val());
	key = key || e.data.key
	
	console.log('check prior vs. new', prior, ' !== ',  val);
	if( prior !== val )
	{
		console.log('changed');
		if( inline ) inline.text(val);
		var change = {prior: prior, value: val, name: e.data.name, orginal: org, stamp: new Date().getTime()};

		if( ! cert.editable.changes[key] ) cert.editable.changes[key] = [];
		cert.editable.changes[key].push(change);
		
		console.log('change pushed', JSON.stringify(change));
		
	} else
	{
		console.log('unchanged');
	}
	
	console.groupEnd();
};

cert.editable.ui.click = function(e)
{
	console.group('cert.editable.ui.click');
	console.log('e', e);
	
	var target = $(e.currentTarget);
	var inputSpan = target.next('span.input');
	
	var input = cert.editable.ui.populateInput(target, inputSpan, e);

	input.unbind();
	input.bind('focusout', e.data, cert.editable.ui.blur);
	input.bind('keydown', e.data, cert.editable.ui.keypress);
	
	target.hide(cert.showHide.ui.type);
	cert.showHide.ui.show(inputSpan, function(){ test(input, target); } );
	
	console.groupEnd();
};

cert.editable.ui.populateInput = function(target, inputSpan, e)
{
	var input = inputSpan.find('input');
	
	console.log('target', target, 'inputSpan', inputSpan, target.text(), input.val());
	
	if( ! e ) console.warn('faking e... will that matter?');
	e = e || { data: {}};
	
	var trimmed = $.trim(target.text());
	input.val(trimmed);
	e.data.value = {};
	e.data.value.prior = trimmed;
	e.data.value.original = $.trim(target.data('value-original'));
	target.data('value-prior', trimmed);
	
	return input;
};

test = function(input, target)
{
	input.focus(); 	
	input.select();
	console.log('after hide target.text()', target.text());
};

cert.editable.ui.focus = function(e)
{
	console.group('cert.editable.ui.focus');
	console.log('e', e);
	
	document.designMode = 'on';
	
	//cert.editable.ui.selection(e.currentTarget);

	console.groupEnd();
};

cert.editable.ui.getName = function(i)
{
	console.group('cert.editable.ui.getName');

	var classes = i.attr('class').split(' ');

	console.log('classes', classes);
	var name = null;
	
	for( var y = classes.length - 1; y >= 0; y-- )
	{
		var c = classes[y];
		var idx = c.indexOf('-content');
		
		console.log('y', y, 'c', c, 'idx', idx);
		
		if( idx > -1 )
		{
			name = c.substring(0, idx);
			break;
		}
	}
	
	console.log('name', name);
	console.groupEnd();
	
	return name;
};

cert.editable.ui.addInput = function(i, name)
{
	var input = i.next('span.input');
	if( input.length === 0 )
	{
		if( name === null )
		{
			name = cert.editable.ui.getName(i);
		}
		
		if( name === null )
		{
			console.error('name should never be null', i);
		}
		
		size = i.data('input-size') || "0";
		i.after('<span class="input ' + name + '-input" style="display: none;"><input name="' + name + '" size="' + size + '"/></span>');
		var input = i.next('span.input');
	} 
	
	return input;
}

cert.editable.ui.keypress = function(e)
{
	//console.log('cert.editable.ui.keypress', e);
	
	e.data.jumpToNext = true;
	e.data.acceptChange = true;

	//console.log('switch', e.keyCode);
	switch(e.keyCode)
	{
		case 9:
			// tab
			cert.editable.ui.leave(e);
		break;
		case 13: 
			// enter
			cert.editable.ui.leave(e);
			break;
		case 27:
			// esc
			e.data.jumpToNext = false;
			e.data.acceptChange = false;
			cert.editable.ui.leave(e);
			break;
	}
}

cert.editable.ui.hook = function(editAll)
{
	console.group('cert.editable.ui.focus');
	var forms = $('form');
	
	for( var y = forms.length - 1; y >= 0; y-- )
	{
		var form = $(forms[y]);
		var inlines = form.find('.inline');
		editAll = editAll || false;
		
		cert.editable.ui.clickToEdit = !editAll;
		
		var keyField = form.find('input[name="key"]');
		var key = keyField.val();
		console.log('keyField', keyField, 'key', key);
		
		if( cert.editable.ui.contentEditable )
		{
			inlines.bind('focus', cert.editable.ui.focus);
			inlines.bind('blur', cert.editable.ui.blur);
		} else
		{
			for( var x = inlines.length - 1; x >= 0; x-- )
			{
				var i = $(inlines[x]);
				
				if( $.trim(i.text()) === "" )
				{
					var val = i.data('value-default');
					console.log('using default value', val, i);
					i.text(val);
				}
				
				i.data('value-original', i.text());
				
				var name = cert.editable.ui.getName(i);
				var inputSpan = cert.editable.ui.addInput(i, name, editAll);
				var input = inputSpan.find('input');
				
				if( editAll == false ) 
				{
					i.bind('click', {name: name, key: key}, cert.editable.ui.click);
					cert.showHide.ui.show(i);
					cert.showHide.ui.hide(inputSpan);
				} else
				{
					//  edit all in effect
					var inputSpan = i.next('span.input');
		
					var input = cert.editable.ui.populateInput(i, inputSpan);
		
					cert.showHide.ui.show(inputSpan);
					cert.showHide.ui.hide(i);
				}
				console.log('name', name, 'i', i, 'i.data', i.data(), 'input', input);
			}
		}
		
		var saveBtn = form.find('.save');
		saveBtn.unbind().bind('click', cert.editable.save.event);
		
		//if( editAll === false ) 
		//{
			cert.editable.ui.checkSave(cert.editable.changes[key], form);
		//} else
		//{
		//	saveBtn.attr('disabled', false);
		//}
		
		console.log('inlines', inlines);
	} 
	
	var editAll = $('a.editAll');
	
	editAll.unbind().bind('click', {key: key}, cert.editable.ui.editAll.event);
	
	console.log('editAll', editAll);
	
	console.groupEnd();
};

cert.editable.ui.editAll = {};
cert.editable.ui.editAll.event = function(e)
{
	console.group('cert.editable.ui.editAll.event');
	e.preventDefault();
	
	console.log('e', e);
	cert.routing.router.navigate('profile/edit/' + e.data.key, true);
	console.groupEnd();
}

// cert.editable.ui.keypress = function (e)
// {
	// console.group('cert.editable.ui.focus');
	
	// window.getSelection().removeAllRanges();
	
	// console.groupEnd();
// };

cert.editable.ui.selection = function(el)
{
    var range = document.createRange();
    range.selectNodeContents(el);
    var sel = window.getSelection();
    sel.removeAllRanges();
    sel.addRange(range);
};

cert.routing = {};
cert.routing.common = {};
cert.routing.profile = {};
cert.routing.roster = {};
cert.routing.profile.editAll = function(key)
{
	console.group('cert.routing.profile.editAll');
	
	console.log('key', key);
	cert.editable.ui.hook(true);
	
	console.groupEnd();
};

cert.routing.profile.edit = function(key)
{
	console.group('cert.routing.profile.edit');
	
	console.log('key', key);
	cert.editable.ui.hook(false);
	
	console.groupEnd();
};

cert.routing.roster.list = function()
{
	var m = 'cert.routing.roster.list';
	console.group(m);
	if( cert.loggedIn.member.notLoggedIn )
	{
		cert.member.ui.notLoggedIn($('.roster-content'));
	} else
	{
		cert.roster.load.event();
	}
	console.groupEnd();
};

cert.routing.init = function()
{
	console.group('cert.routing.init');
	var AppRouter = Backbone.Router.extend({
		routes: {
			'profile/edit/:key': 'editAll',
			'profile/': 'edit',
			'profile/:key': 'edit',
			'roster/': 'roster',
		},
		editAll: cert.routing.profile.editAll,
		edit: cert.routing.profile.edit,
		roster: cert.routing.roster.list,
	});
	// Initiate the router
	cert.routing.router = new AppRouter;
	// Start Backbone history a neccesary step for bookmarkable URL's
	Backbone.history.start({pushState: true});

	console.groupEnd();
};

cert.init = function(user, member, memberships)
{
	console.group('cert.init');
	
	console.log('user', user, 'member', member);
	cert.loggedIn.user = user;
	cert.loggedIn.member = member;

	$('button').button();
	cert.routing.init();
	
	cert.permissions.check(cert.loggedIn.memberships);
	
	console.groupEnd();
};

cert.permissions = {};
cert.permissions.check = function(memberships, team)
{
	console.group('cert.permissions.check');
	
	var ms = null
	for( m in memberships )
	{
		ms = memberships[m];
	}
	
	var elements = $('.oem-officer');
	console.log('ms', ms, 'elements', elements);
	if( ms.oem || ms.officer )
	{
		cert.showHide.ui.show(elements);
	} else
	{
		cert.showHide.ui.hide(elements);
	}
	console.groupEnd();
};

cert.notify = {};
cert.notify.add = function(msg, title, sticky, image, className, time)
{
	console.log('notify', title, msg);

	var unique_id = $.gritter.add({
		// (string | mandatory) the heading of the notification
		title: title,
		// (string | mandatory) the text inside the notification
		text: msg,
		// (string | optional) the image to display on the left
		image: image,
		// (bool | optional) if you want it to fade out on its own or just sit there
		sticky: sticky,
		// (int | optional) the time you want it to be alive for before fading out
		time: time,
		// (string | optional) the class name you want to apply to that specific message
		class_name: className
	});
	
	return unique_id;
};

cert.notify.defaults = {};
cert.notify.defaults.all = {};
cert.notify.defaults.all.time = 3000;
cert.notify.defaults.info = {};
cert.notify.defaults.info.image = '/images/icons/32x32/information.png';
cert.notify.defaults.info.className='info';
cert.notify.defaults.error = {};
cert.notify.defaults.error.image = '/images/icons/32x32/exclamation.png';
cert.notify.defaults.error.className='error';
cert.notify.defaults.success = {};
cert.notify.defaults.success.time = 3000;
cert.notify.defaults.success.image = '/images/icons/32x32/accept.png';
cert.notify.defaults.success.className='success';
cert.notify.defaults.warn = {};
cert.notify.defaults.warn.time = 8000;
cert.notify.defaults.warn.image = '/images/icons/32x32/error.png';
cert.notify.defaults.warn.className='warn';

cert.notify.info = {};
cert.notify.info.add = function(msg, title, sticky, time)
{
	title = title || "Information";
	sticky = sticky || false;
	return cert.notify.add(msg, title, sticky, cert.notify.defaults.info.image, cert.notify.defaults.info.className);
}

cert.notify.warn = {};
cert.notify.warn.add = function(msg, title, sticky, time)
{
	sticky = sticky || false;
	title = title || "Warning";
	time = time || cert.notify.defaults.warn.time || cert.notify.defaults.all.time ;
	return cert.notify.add(msg, title, sticky, cert.notify.defaults.warn.image, cert.notify.defaults.warn.className);
}

cert.notify.error = {};
cert.notify.error.add = function(msg, title, sticky, time)
{
	sticky = sticky || true;
	title = title || "Error";
	time = time || cert.notify.defaults.error.time || cert.notify.defaults.all.time ;
	return cert.notify.add(msg, title, sticky, cert.notify.defaults.error.image, cert.notify.defaults.error.className);
}

cert.notify.success = {};
cert.notify.success.add = function(msg, title, sticky, time)
{
	sticky = sticky || false;
	title = title || "Success";
	time = time || cert.notify.defaults.success.time || cert.notify.defaults.all.time ;
	return cert.notify.add(msg, title, sticky, cert.notify.defaults.success.image, cert.notify.defaults.success.className);
}

cert.roster.load.failure = function(error, status, xhr)
{
	var m = 'cert.roster.load.event';
	console.group(m);
	
	try
	{
		console.log('error, status, xhr',error, status, xhr);
	} catch(e)
	{
		console.error(m, e);
	}
	console.groupEnd();	
};

cert.roster.load.success = function(data, status, xhr)
{
	var m = 'cert.roster.load.event';
	console.group(m);
	
	console.log('data, status, xhr',data, status, xhr);
	
	var cnt = data.length;
	
	var membership = null;
	for( var m in cert.loggedIn.memberships )
	{
		membership = cert.loggedIn.memberships[m];
	}
	
	console.log('membership', membership);
	
	var tName = (membership.oem || membership.officer) && ! cert.forcePublic ? 'private' : 'public';
	tName += '-roster';
	var headerT = cert.template.ui.get(tName + '-header');
	
	var table = $('table.roster');
	table.append(headerT);
	
	console.log('cnt, headerT', cnt, headerT);
	
	for( var x = 0; x < cnt; x++ )
	{
		var m = data[x];
		var rowT = cert.template.ui.getPopulated(tName,m);
		
		if( x % 2 === 0 ) rowT.addClass('even');
		
		console.log('x, m', x, m, 'rowT', rowT);
		
		table.append(rowT);		
	}
	
	var parent = table.closest('div.roster-content');
	var loading = parent.find('div.loading');
	cert.showHide.ui.hide(loading);
	console.log('parent, loading', parent, loading);
	
	console.groupEnd();	
};

cert.roster.load.event = function()
{
	var m = 'cert.roster.load.event';
	console.group(m);
	
	try
	{
		var settings =
		{
			url: '/roster/list',
			data: {},
			cache: false,
			type: 'GET',
			success: cert.roster.load.success,
			error: cert.roster.load.success,
		}
		
		console.log('settings', settings);
		
		$.ajax(settings);
	} catch(e)
	{
		console.error(m, e);
	}
	console.groupEnd();	
};

cert.member.ui.notLoggedIn = function(holder)
{
	var t = cert.template.ui.get('not-logged-in');

	holder.html(t.html());
}