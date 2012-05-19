cert = {};
cert.editable = {};
cert.editable.save = {};
cert.editable.ui = {};
cert.editable.ui.contentEditable = false;
cert.editable.ui.clickToEdit = true;
cert.editable.changes = [];

cert.showHide = {};
cert.showHide.ui = {};
cert.showHide.ui.type = '';

var console = console || {};
console.log = console.log || function(){};
console.group = console.group || function(){};
console.groupEnd = console.groupEnd || function(){};
console.warn = console.warn || function(){};
console.error = console.error || function(){};

//window.onerror = function(){ alert('error');};

cert.showHide.ui.show = function(el, callback)
{
	el.show(0, null, callback);
};

cert.showHide.ui.hide = function(el, callback)
{
	el.hide(0, null, callback);
};

cert.editable.save.buildChanges = function(e)
{
	console.group('cert.editable.save.buildChanges');
	
	console.warn('loop the inputs to build a change list since we couldn\'t do that in real time');
	
	console.groupEnd();
};

cert.editable.save.buildJSON = function(changes, e)
{
	var obj = {};
	var target = $(e.currentTarget);
	var form = target.closest('form');
	var keyInput = form.find('input[name="key"]');
	
	var key = keyInput.val();
	
	console.log('target, form, keyInput, key', target, form, keyInput, key);
	
	obj.key = key;
	obj.changes = changes;
	
	console.log('obj', obj);

	return JSON.stringify(obj);
};

cert.editable.save.event = function(e)
{
	console.group('cert.editable.save.event');
	
	e.preventDefault();

	var changes = null;
	if( ! cert.editable.ui.clickToEdit )
	{
		changes = cert.editable.save.buildChanges(e);
 	}  else
	{
		changes = cert.editable.save.flattenChanges(cert.editable.changes);		
	}
	
	console.log('changes', changes);
	var json = cert.editable.save.buildJSON(changes, e);
	console.log('json', json);
	
	var settings =
	{
		url: '/profile/save',
		data: {json: json},
		cache: false,
		type: 'POST',
		success: cert.editable.save.success,
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

cert.editable.ui.checkSave = function(changes)
{
	console.group('cert.editable.ui.checkSave');
	
	var saveBtn = $('form .save');
	
	if( changes.length > 0 )
	{
		cert.showHide.ui.show(saveBtn);
		console.log('show save');
	} else
	{
		cert.showHide.ui.hid(saveBtn);
		console.log('hide save');
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
		var static = parent.prev('span');

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
			cert.editable.ui.trackChanges(target, static, e);
		}
		
		parent.hide(cert.showHide.ui.type);
		static.show(cert.showHide.ui.type);
		
		if( e.data.jumpToNext )
		{
			var next = $(static.nextAll('.inline')[0]);
			console.log('next', next);
			// wait a little bit to let the UI catch up
			window.setTimeout(function(){ console.log('trigger click'); next.trigger('click', e); }, 200);
		}
		
		target.unbind();
	} else
	{
		document.designMode = 'off';	
	}
	
	cert.editable.ui.checkSave(cert.editable.changes);
	
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

cert.editable.ui.trackChanges = function(target, static, e)
{
	console.group('cert.editable.ui.trackChanges');
	console.log('target, static', target, static, 'e', e);
	
	var prior = e.data.value.prior;// || static.data('value-prior');
	var org = e.data.value.original;// || static.data('value-original');
	var val = $.trim(target.val());
	
	console.log('check prior vs. new', prior, ' !== ',  val);
	if( prior !== val )
	{
		console.log('changed');
		static.text(val);
		var change = {prior: prior, value: val, name: e.data.name, orginal: org, stamp: new Date().getTime()};

		cert.editable.changes.push(change);
		
		console.error('pushed', JSON.stringify(change));
		
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
	var input = inputSpan.find('input');
	
	console.log('target', target, 'inputSpan', inputSpan, target.text(), input.val());
	
	var trimmed = $.trim(target.text());
	input.val(trimmed);
	e.data.value = {};
	e.data.value.prior = trimmed;
	e.data.value.original = $.trim(target.data('value-original'));
	target.data('value-prior', trimmed);
	
	input.unbind();
	input.bind('focusout', e.data, cert.editable.ui.blur);
	input.bind('keydown', e.data, cert.editable.ui.keypress);

	
	target.hide(cert.showHide.ui.type);
	cert.showHide.ui.show(inputSpan, function(){ test(input, target); } );
	
	console.groupEnd();
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

cert.editable.ui.hook = function()
{
	console.group('cert.editable.ui.focus');

	var inlines = $('.inline');
	
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
			var inputSpan = cert.editable.ui.addInput(i, name);
			var input = inputSpan.find('input');
			
			i.bind('click', {name: name}, cert.editable.ui.click);
			console.log('name', name, 'i', i, 'i.data', i.data());
			
			console.log('input', input);
		}
	}
	
	$('form .save').bind('click', cert.editable.save.event);
	
	console.log('inlines', inlines);
	
	console.groupEnd();
};

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
cert.routing.editAll = function(actions)
{
	console.group('cert.routing.editAll');
	
	console.log('actions', actions);
	
	console.groupEnd();
};

cert.routing.init = function()
{
	console.group('cert.routing.init');
	var AppRouter = Backbone.Router.extend({
		routes: {
			"profile/edit/:param": "editAll",
		},
		editAll: cert.routing.editAll,
	});
	// Initiate the router
	cert.routing.router = new AppRouter;
	// Start Backbone history a neccesary step for bookmarkable URL's
	Backbone.history.start({pushState: true});

	console.groupEnd();
};

cert.init = function()
{
	console.group('cert.init');

	cert.editable.ui.hook();
	cert.routing.init();
	
	console.groupEnd();
};
