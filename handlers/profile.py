import logging
import os
import webapp2
from webapp2_extras import jinja2

import models.common
import handlers.base

class Index(handlers.base.Base):
	def get(self):
		data = self.commonData()
		#logging.debug('json ' + self.toJSON(data));
		if data['user'] == None:
			return self.redir(data['loginURL'])
		else:
			self.render_template('profile/index.htm', **data)

app = webapp2.WSGIApplication([
	('/profile/', Index)
], debug=True)