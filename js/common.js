app = {};
app.editable = {};
app.editable.ui = {};
app.editable.ui.contentEditable = false;
app.editable.changes = [];

app.showHide = {};
app.showHide.ui = {};
app.showHide.ui.type = '';

app.editable.ui.leave = function (e, keepChange)
{
	console.log('e', e, 'keepChange', keepChange);
	e.stopPropagation();
	if( ! app.editable.ui.contentEditable )
	{
		var target = $(e.currentTarget);
		var parent = target.closest('span');
		var static = parent.prev('span');

		keepChange = e.acceptChange || keepChange;
		
		if( keepChange ) 
		{
			app.editable.ui.trackChanges(target, static, e);
		}
	
		parent.hide(app.showHide.ui.type);
		static.show(app.showHide.ui.type);
		
		if( e.jumpToNext )
		{
			var next = $(static.nextAll('.inline')[0]);
			console.log('next', next);
			// wait a little bit to let the UI catch up
			window.setTimeout(function(){ next.click(); }, 200);
		}
	} else
	{
		document.designMode = 'off';	
	}
};

app.editable.ui.blur = function (e)
{
	console.group('app.editable.ui.blur');
	// lame that we're hooking the blur event
	
	e.jumpToNext = false;
	e.acceptChange = true;
	
	console.log('blur e', e);
	app.editable.ui.leave(e, true);
	
	console.groupEnd();
};

app.editable.ui.trackChanges = function(target, static, e)
{
	console.group('app.editable.ui.trackChanges');
	console.log('target, static', target, static);
	
	var old = $.trim(static.data('value-prior'));
	var val = $.trim(target.val());
	
	console.log('check old vs. new', old, ' !== ',  val);
	if( old !== val )
	{
		console.log('changed');
		static.text(val);
		var change = {static: static, prior: static.data('value-prior'), value: val, name: e.data.name, orginal: static.data('value-original'),};
		app.editable.changes.push(change);
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
	target.data('value-prior', trimmed);
	
	target.hide(app.showHide.ui.type);
	
	inputSpan.show(app.showHide.ui.type, function(){ input.focus(); 	console.log('after hide target.text()', target.text()); } );
	
	console.groupEnd();
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
	var classes = i.attr('class').split(' ');

	for( var y = classes.length -1; y >= 0; y-- )
	{
		var c = classes[y];
		var idx = c.indexOf('-content');
		
		console.log('y', y, 'c', c, 'idx', idx);
		
		if( idx >= -1 )
		{
			name = c.substring(0, idx);
			break;
		}
	}
	
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
	
	e.jumpToNext = true;
	e.acceptChange = true;

	console.log('switch', e.keyCode);
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
			e.jumpToNext = false;
			e.acceptChange = false;
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
		inlines.bind('click', app.editable.ui.click);
		for( var x = inlines.length - 1; x >= 0; x-- )
		{
			var i = $(inlines[x]);
			
			console.log('i', i, 'i.data', i.data());
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
			input.bind('blur', {name: name}, app.editable.ui.blur);
			input.bind('keydown', {name: name}, app.editable.ui.keypress);
			
			console.log('input', input);
		}
	}
	//inlines.bind('keyup', app.editable.ui.keypress);
	
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
