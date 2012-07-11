import datetime
import json
import time

from google.appengine.ext import db
from google.appengine.api import users

SIMPLE_TYPES = (int, long, float, bool, dict, basestring, list)

class Base(db.Model):
	created = db.DateTimeProperty(auto_now_add=True)
	modified = db.DateTimeProperty(auto_now=True)
	createdBy = db.UserProperty(auto_current_user_add=True)
	modifiedBy = db.UserProperty(auto_current_user=True)
	
	def to_dict(self, obj):
		if type(obj) == dict:
			return obj

		if type(obj) == users.User:
			output = {}

			for key, prop in obj.properties().iteritems():
				value = getattr(obj, key)

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
		
		if type(obj) == list:
			outList = []
			for o in obj:
				outList.append(self.to_dict(o))
			
			return outList
			
		if obj is None:
			return obj
		
		newObj = dict()
		for p in obj.properties():
			val = unicode(getattr(obj, p))
			if getattr(obj, p) is None:
				val = None
			newObj[p] = val
		
		return newObj
		
	def toJSON(self):
		obj = self
		return json.dumps(self.to_dict(obj))
		
class Member(Base):
	firstName = db.StringProperty()
	lastName = db.StringProperty()
	email = db.EmailProperty()
	email2 = db.EmailProperty()
	address = db.PostalAddressProperty()
	cell = db.PhoneNumberProperty()
	home = db.PhoneNumberProperty()
	user = db.UserProperty()
	
	officer = db.BooleanProperty()
	oem = db.BooleanProperty()
	
	rosterEmail = db.BooleanProperty()
	rosterEmail2 = db.BooleanProperty()
	rosterCell = db.BooleanProperty()
	rosterHome = db.BooleanProperty()
	
	def filterForRoster(self, member):
		if not(member.officer or member.oem):
			self.email = self.email if self.rosterEmail else 'none@example.com'
			self.email2 =  self.email2 if self.rosterEmail2 else 'none@example.com'
			self.cell = self.cell if self.rosterCell else '555-555-5555'
			self.home = self.home if self.rosterHome else '555-555-5555'
		
		return self

class Team(Base):
	name = db.StringProperty()
	location = db.PostalAddressProperty()
	
class Membership(Base):
	team = db.ReferenceProperty(Team)
	member = db.ReferenceProperty(Member)
	
	joined = db.DateTimeProperty()