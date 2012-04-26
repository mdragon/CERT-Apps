ui = {};
ui.editable = {};
ui.editable.contentEditable = false;

ui.showHide = {};
ui.showHide.type = 'fade';

ui.editable.blur = function (e)
{
	console.group('ui.editable.blur');
	// lame that we're hooking the blur event

	if( ! ui.editable.contentEditable )
	{
		var target = $(e.currentTarget);
		var static = target.prev('span');
		
		target.hide(ui.showHide.type);
		static.show(ui.showHide.type);
	} else
	{
		document.designMode = 'off';	
	}
	
	console.groupEnd();
};

ui.editable.click = function(e)
{
	console.group('ui.editable.click');
	console.log('e', e);
	
	var target = $(e.currentTarget);
	var input = target.next('span.input');
	
	console.log('target', target, 'input', input);
	
	target.hide(ui.showHide.type);
	input.show(ui.showHide.type, function(){ input.find('input').focus();} );
	
	console.groupEnd();
};

ui.editable.focus = function(e)
{
	console.group('ui.editable.focus');
	console.log('e', e);
	
	document.designMode = 'on';
	
	//ui.editable.selection(e.currentTarget);

	console.groupEnd();
};

ui.editable.addInput = function(i)
{
	var input = i.next('span.input');
	if( input.length === 0 )
	{
		var classes = i.attr('class').split(' ');
		var name = null;
		for( var y = classes.length -1; y >= 0; y-- )
		{
			var c = classes[y];
			var idx = c.indexOf('-content');
			
			console.log('y', y, 'c', c, 'idx', idx);
			
			if( idx >= -1 )
			{
				name = c.substring(0, idx);
			}
		}
		
		if( name === null )
		{
			console.error('name should never be null');
		}
		
		size = i.data('input-size') || "0";
		i.after('<span class="input ' + name + '-input" style="display: none;"><input name="' + name + '" size="' + size + '"/></span>');
		var input = i.next('span.input');
	} 
	
	return input;
}

ui.editable.hook = function()
{
	console.group('ui.editable.focus');

	var inlines = $('.inline');
	
	if( ui.editable.contentEditable )
	{
		inlines.bind('focus', ui.editable.focus);
		inlines.bind('blur', ui.editable.blur);
	} else
	{
		inlines.bind('click', ui.editable.click);
		for( var x = inlines.length - 1; x >= 0; x-- )
		{
			var i = $(inlines[x]);
			var input = ui.editable.addInput(i);
			input.bind('blur', ui.editable.blur);
			input.bind('focusout', ui.editable.blur);
			
		}
	}
	//inlines.bind('keyup', ui.editable.keypress);
	
	console.log('inlines', inlines);
	
	console.groupEnd();
};

ui.editable.keypress = function (e)
{
	console.group('ui.editable.focus');
	
	window.getSelection().removeAllRanges();
	
	console.groupEnd();
};

ui.editable.selection = function(el)
{
    var range = document.createRange();
    range.selectNodeContents(el);
    var sel = window.getSelection();
    sel.removeAllRanges();
    sel.addRange(range);
};

ui.init = function()
{
	console.group('ui.init');

	ui.editable.hook();
	
	console.groupEnd();
};
