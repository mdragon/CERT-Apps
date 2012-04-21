import json
import webapp2
from google.appengine.ext import db
from google.appengine.api import users

import models

class Base(db.Model):
	def to_dict(self):
		return dict([(p, unicode(getattr(self, p))) for p in self.properties()])

	def toJSON(self):
		return json.dumps(self.to_dict())

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
