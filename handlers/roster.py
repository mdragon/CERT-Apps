import json
import logging
import os
import webapp2

from google.appengine.ext import db
from webapp2_extras import jinja2

import models.common
import handlers.base

class Index(handlers.base.Base):
	def get(self):
		logging.debug('running Index.get')
		
		data = self.commonData()
		
		self.render_template('roster/list.htm', **data)

class List(handlers.base.Base):
	def get(self):
		logging.debug('running List.get')
		data = self.commonData()
		self.noCache()
		
		#logging.debug('json ' + self.toJSON(data));
		if data['user'] == None:
			return self.redir(data['loginURL'])
		else:
			membersQ = models.common.Member.all()
			members = []
			
			for m in membersQ:
				members.append(m.filterForRoster(data['member']))
							
			data['members']	= members
			data['json']	= self.toJSON(members)
		
			self.renderJSON(data);

class Save(handlers.base.Base):
	def post(self):
		post = self.request.get('json')
		logging.debug('profile save json ' + post)
		
		obj = json.loads(post)
		
		m = db.get(obj['key'])
		
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
	('/roster/list', List),
	('/roster/.*', Index),
], debug=True)