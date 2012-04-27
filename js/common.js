app = {};
app.editable = {};
app.editable.ui = {};
app.editable.ui.contentEditable = false;
app.editable.changes = [];

app.showHide = {};
app.showHide.ui = {};
app.showHide.ui.type = '';

app.showHide.ui.show = function(el, callback)
{
	el.show(0, null, callback);
};

app.showHide.ui.hide = function(el, callback)
{
	el.hide(0, null, callback);
};

app.editable.ui.leave = function (e, keepChange)
{
	console.group('app.editable.ui.leave');
	console.log('e', e, 'keepChange', keepChange);
	e.stopPropagation();
	e.preventDefault();
	if( ! app.editable.ui.contentEditable )
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
			app.editable.ui.trackChanges(target, static, e);
		}
		
		parent.hide(app.showHide.ui.type);
		static.show(app.showHide.ui.type);
		
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
	
	console.groupEnd();
};

app.editable.ui.blur = function (e)
{
	console.group('app.editable.ui.blur');
	// lame that we're hooking the blur event
	
	e.data.jumpToNext = false;
	e.data.acceptChange = true;
	
	console.log('blur e', e);
	app.editable.ui.leave(e, true);
	
	console.groupEnd();
};

app.editable.ui.trackChanges = function(target, static, e)
{
	console.group('app.editable.ui.trackChanges');
	console.log('target, static', target, static, 'e', e);
	
	var prior = e.data.value.prior;// || static.data('value-prior');
	var org = e.data.value.original;// || static.data('value-original');
	var val = $.trim(target.val());
	
	console.log('check prior vs. new', prior, ' !== ',  val);
	if( prior !== val )
	{
		console.log('changed');
		static.text(val);
		var change = {prior: prior, value: val, name: e.data.name, orginal: org,};

		app.editable.changes.push(change);
		
		console.error('pushed', JSON.stringify(change));
		
	} else
	{
		console.log('unchanged');
	}
	
	console.groupEnd();
};

app.editable.ui.click = function(e)
{
	console.group('app.editable.ui.click');
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
	input.bind('focusout', e.data, app.editable.ui.blur);
	input.bind('keydown', e.data, app.editable.ui.keypress);

	
	target.hide(app.showHide.ui.type);
	app.showHide.ui.show(inputSpan, function(){ test(input, target); } );
	
	console.groupEnd();
};

test = function(input, target)
{
	input.focus(); 	
	input.select();
	console.log('after hide target.text()', target.text());
};

app.editable.ui.focus = function(e)
{
	console.group('app.editable.ui.focus');
	console.log('e', e);
	
	document.designMode = 'on';
	
	//app.editable.ui.selection(e.currentTarget);

	console.groupEnd();
};

app.editable.ui.getName = function(i)
{
	console.group('app.editable.ui.getName');

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

app.editable.ui.addInput = function(i, name)
{
	var input = i.next('span.input');
	if( input.length === 0 )
	{
		if( name === null )
		{
			name = app.editable.ui.getName(i);
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

app.editable.ui.keypress = function(e)
{
	//console.log('app.editable.ui.keypress', e);
	
	e.data.jumpToNext = true;
	e.data.acceptChange = true;

	//console.log('switch', e.keyCode);
	switch(e.keyCode)
	{
		case 9:
			// tab
			app.editable.ui.leave(e);
		break;
		case 13: 
			// enter
			app.editable.ui.leave(e);
			break;
		case 27:
			// esc
			e.data.jumpToNext = false;
			e.data.acceptChange = false;
			app.editable.ui.leave(e);
			break;
	}
}

app.editable.ui.hook = function()
{
	console.group('app.editable.ui.focus');

	var inlines = $('.inline');
	
	if( app.editable.ui.contentEditable )
	{
		inlines.bind('focus', app.editable.ui.focus);
		inlines.bind('blur', app.editable.ui.blur);
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
			
			var name = app.editable.ui.getName(i);
			var inputSpan = app.editable.ui.addInput(i, name);
			var input = inputSpan.find('input');
			
			i.bind('click', {name: name}, app.editable.ui.click);
			console.log('name', name, 'i', i, 'i.data', i.data());
			
			console.log('input', input);
		}
	}
	
	console.log('inlines', inlines);
	
	console.groupEnd();
};

// app.editable.ui.keypress = function (e)
// {
	// console.group('app.editable.ui.focus');
	
	// window.getSelection().removeAllRanges();
	
	// console.groupEnd();
// };

app.editable.ui.selection = function(el)
{
    var range = document.createRange();
    range.selectNodeContents(el);
    var sel = window.getSelection();
    sel.removeAllRanges();
    sel.addRange(range);
};

app.init = function()
{
	console.group('app.init');

	app.editable.ui.hook();
	
	console.groupEnd();
};
