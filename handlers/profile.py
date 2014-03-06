import json
import logging
import os
import webapp2

from google.appengine.ext import ndb
from webapp2_extras import jinja2

import models.common
import handlers.base

class Index(handlers.base.Base):
	@ndb.toplevel
	def get(self):
		logging.debug('running Index.get')
		data = self.commonData()
		self.noCache()
		
		#logging.debug('json ' + self.toJSON(data));
		if data['user'] == None:
			return self.redir(data['loginURL'])
		else:
			self.render_template('profile/index.htm', **data)

class Save(handlers.base.Base):
	@ndb.toplevel
	def post(self):
		post = self.request.get('json')
		logging.debug('profile save json ' + post)
		
		obj = json.loads(post)
		
		mKey = ndb.Key(urlsafe=obj['key'])
		m = mKey.get()
		
		for c in obj['changes']:
			logging.debug('c ' + c['name'] +', ' + c['value'] +', ' + str(c['stamp']))
			val = getattr(m, c['name'])
			if val != None:
				logging.debug(c['name'] + ': ' + val )
			else:
				logging.debug(c['name'] + ': None')
			setattr(m, c['name'], c['value'])
		
		m.put()
		
		self.response.headers['Content-Type'] = 'application/json'
		self.response.write(post);
			
app = webapp2.WSGIApplication([
	('/profile/save', Save),
	('/profile/.*', Index),
], debug=True)