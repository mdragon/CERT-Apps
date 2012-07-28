import datetime
import logging
import json
import time

from google.appengine.ext import ndb
from google.appengine.api import users

SIMPLE_TYPES = (int, long, float, bool, dict, basestring, list)

class Base(ndb.Model):
	created = ndb.DateTimeProperty(auto_now_add=True)
	modified = ndb.DateTimeProperty(auto_now=True)
	createdBy = ndb.UserProperty(auto_current_user_add=True)
	modifiedBy = ndb.UserProperty(auto_current_user=True)
	
	def to_dict(self, obj):
		if type(obj) == dict:
			return obj

		if type(obj) == users.User:
			output = {}

			for key, prop in obj._properties.iteritems():
				value = getattr(obj, key)

				if value is None or isinstance(value, SIMPLE_TYPES):
					output[key] = value
				elif isinstance(value, datetime.date):
					# Convert date/datetime to ms-since-epoch ("new Date()").
					ms = time.mktime(value.utctimetuple())
					ms += getattr(value, 'microseconds', 0) / 1000
					output[key] = int(ms)
				elif isinstance(value, ndb.GeoPt):
					output[key] = {'lat': value.lat, 'lon': value.lon}
				elif isinstance(value, ndb.Model):
					output[key] = to_dict(value)
				else:
					raise ValueError('cannot encode ' + repr(prop))

			return output
		
		if type(obj) == list:
			outList = []
			for o in obj:
				outList.append(self.to_dict(o))
			
			return outList
		
		#if 'to_dict' in dir(obj):
		#	logging.debug('has to_dict')
		#	return obj.to_dict(obj)
			
		if obj is None:
			return obj
		
		newObj = dict()
		for p in type(obj)._properties:
			val = unicode(getattr(obj, p))
			if getattr(obj, p) is None:
				val = None
			newObj[p] = val

		if 'key' in dir(obj):
			newObj['key'] = str(obj.key.urlsafe())
			logging.debug('adding model object key ' + str(newObj['key']))
		else:
			logging.debug('not adding model object key as it does not exist')
		
		return newObj
		
	def toJSON(self):
		obj = self
		d = self.to_dict(obj)

		return json.dumps(self.to_dict(d))
		
class Member(Base):
	firstName = ndb.StringProperty()
	lastName = ndb.StringProperty()
	email = ndb.StringProperty()
	email2 = ndb.StringProperty()
	address = ndb.StringProperty()
	cell = ndb.StringProperty()
	home = ndb.StringProperty()
	user = ndb.UserProperty()
	
	rosterEmail = ndb.BooleanProperty()
	rosterEmail2 = ndb.BooleanProperty()
	rosterCell = ndb.BooleanProperty()
	rosterHome = ndb.BooleanProperty()
	
	def filterForRoster(self, membership):
		if not(membership.officer or membership.oem):
			self.rosterEmail = True if self.rosterEmail is None else self.rosterEmail
			self.rosterEmail2 = True if self.rosterEmail2 is None else self.rosterEmail2
			self.rosterCell = True if self.rosterCell is None else self.rosterCell
			self.rosterHome = True if self.rosterHome is None else self.rosterHome
		
			self.email = self.email if self.rosterEmail else 'none@example.com'
			self.email2 =  self.email2 if self.rosterEmail2 else 'none@example.com'
			self.cell = self.cell if self.rosterCell else '555-555-5555'
			self.home = self.home if self.rosterHome else '555-555-5555'
		
		return self

class Team(Base):
	name = ndb.StringProperty()
	location = ndb.StringProperty()
	
class Membership(Base):
	team = ndb.KeyProperty(kind=Team)
	member = ndb.KeyProperty(kind=Member)
	
	joined = ndb.DateTimeProperty()
	
	officer = ndb.BooleanProperty()
	oem = ndb.BooleanProperty()
