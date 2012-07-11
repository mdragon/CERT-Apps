import datetime
import json
import time
import models.common

from google.appengine.ext import db
from google.appengine.api import users
		
class Deployment(models.common.Base):
	name = db.StringProperty()
	description = db.StringProperty()
	location = db.PostalAddressProperty()

	start = db.DateTimeProperty()
	end = db.DateTimeProperty()

class Response(models.common.Base):
	deployment = db.ReferenceProperty(Deployment)
	member = db.ReferenceProperty(models.common.Member)

	attending = db.StringProperty()
	start = db.DateTimeProperty()
	end = db.DateTimeProperty()
