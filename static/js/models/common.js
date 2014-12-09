
window.CERTModels = window.CERTModels || {};

CERTModels.line2Display = function(obj) {
	//Member.line2Display
	var val = ""
	
	var line2 = obj.get('line2')
	if( $.trim(line2).length > 0 )
	{
		val = ", " + line2;
	}

	return val;
}

CERTModels.moveUpData = function(data)
{
	console.log('moveUpData', data);

	if( ! data.error )
	{
		if( data.data !== undefined )
		{
//			console.log('data has data');
			data = data.data;
			//return data.data;
		}
	}

	return data;
}

CERTModels.BaseObject = Ember.Object.extend(
{
	moveUpData: function(data)
	{
		return CERTModels.moveUpData(data);
	},

	sync: function(obj)
	{
		console.group("CERTModels.BaseObject sync");

		if( obj )
		{
			obj = this.moveUpData(obj);

			if( obj )
			{
				console.log('this', JSON.stringify(this));
				console.log('obj', JSON.stringify(obj));

				if( obj.KeyID ) {
					console.log('setting KeyID', this.get('KeyID'), obj.KeyID);
					this.set('KeyID', obj.KeyID);
				} else {
					console.warn("obj passed to sync had no KeyID value")
				}
			} else
			{
				console.warn("obj passed to sync had a data value that was null or undefined");
			}
		}
		else
		{
			console.warn("obj passed to sync was null or undefined");
		}
		
		console.groupEnd();
	},

	cleanData: function()
	{
		this.trimAll();
	},

	trimAll: function()
	{
		Ember.keys(this).forEach( function(prop)
		{
			//console.log(prop, 'checking hasOwn prop')
			if( this.hasOwnProperty(prop) )
			{
				//console.log(prop, 'checking type prop', $.type(this[prop]))

				if( $.type(this[prop]) === $.type("") )
				{
					var val = this.get(prop);

					var trimVal = $.trim(val);

					//console.log(prop, "checking trimmed vs not", "*" + trimVal + "*", "*" + val + "*");

					if( val !== trimVal )
					{
					//	console.log(prop, 'set trimmed', trimVal);

						this.set(prop, trimVal);
					} else
					{
					//	console.log(prop, 'already trimmed');
					}
				}
			}
		}.bind(this)
		);
	}
});
