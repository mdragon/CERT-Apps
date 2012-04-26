import datetime
import json
import time

from google.appengine.ext import db
from google.appengine.api import users

SIMPLE_TYPES = (int, long, float, bool, dict, basestring, list)

class Base(db.Model):
	def toJSON(self):
		return json.dumps(self.to_dict())

	def to_dict(model):
		output = {}

		for key, prop in model.properties().iteritems():
			value = getattr(model, key)

			if value is None or isinstance(value, SIMPLE_TYPES):
				output[key] = value
			elif isinstance(value, datetime.date):
				# Convert date/datetime to ms-since-epoch ("new Date()").
				ms = time.mktime(value.utctimetuple())
				ms += getattr(value, 'microseconds', 0) / 1000
				output[key] = int(ms)
			elif isinstance(value, db.GeoPt):
				output[key] = {'lat': value.lat, 'lon': value.lon}
			elif isinstance(value, db.Model):
				output[key] = to_dict(value)
			else:
				raise ValueError('cannot encode ' + repr(prop))

		return output

		
class Member(Base):
	firstName = db.StringProperty()
	lastName = db.StringProperty()
	email = db.EmailProperty()
	email2 = db.EmailProperty()
	address = db.PostalAddressProperty()
	cell = db.PhoneNumberProperty()
	home = db.PhoneNumberProperty()
	user = db.UserProperty()
	created = db.DateTimeProperty(auto_now_add=True)
	modified = db.DateTimeProperty(auto_now=True)
	createdBy = db.UserProperty(auto_current_user_add=True)
	modifiedBY = db.UserProperty(auto_current_user=True)
