import json
import logging
import os
import webapp2

import datetime
import time

SIMPLE_TYPES = (int, long, float, bool, dict, basestring, list)

from google.appengine.api import users
from webapp2_extras import jinja2

import models.common

class Base(webapp2.RequestHandler):
	@webapp2.cached_property
	def jinja2(self):
		return jinja2.get_jinja2(app=self.app)

	def render_template(self, filename, **template_args):
		self.response.write(self.jinja2.render_template(filename, **template_args))
		
	def commonData(self):
		user = users.get_current_user()
		data = {'user': user}
		if user != None:
			q = models.common.Member.all()
			q.filter('user =', user)
			r = q.fetch(1)
			member = None
			if len(r) > 0:
				member = r[0]
			
			if member == None:
				member = models.common.Member()
				member.firstName = user.nickname()
				member.email = user.email()
				member.user = user
				member.save()
			data['member'] = member
			
			data['logoutURL'] = users.create_logout_url(self.request.path)
		else:
			data['loginURL'] = users.create_login_url(self.request.path)
		
		return data
	
	def to_dict(self, obj):
		if type(obj) == dict:
			return obj

		print ('type', type(obj))
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
		
	def toJSON(self, obj):
		return json.dumps(self.to_dict(obj))
		
	def redir(self, url):
		return webapp2.redirect(url, False)
		
	def noCache(self):
		self.response.headers['Cache-Control'] = 'no-cache'