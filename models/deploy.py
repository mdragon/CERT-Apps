import datetime
import json
import time
import models.common

from google.appengine.ext import ndb
from google.appengine.api import users
		
class Deployment(models.common.Base):
	name = ndb.StringProperty()
	description = ndb.StringProperty()
	location = ndb.StringProperty()

	start = ndb.DateTimeProperty()
	end = ndb.DateTimeProperty()

class Response(models.common.Base):
	deployment = ndb.KeyProperty(kind=Deployment)
	member = ndb.KeyProperty(kind=models.common.Member)

	attending = ndb.StringProperty()
	start = ndb.DateTimeProperty()
	end = ndb.DateTimeProperty()
