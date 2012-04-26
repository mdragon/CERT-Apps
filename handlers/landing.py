import logging
import os
import webapp2
from webapp2_extras import jinja2

import models.common
import handlers.base

class Index(handlers.base.Base):
	def get(self):
		data = self.commonData()
		
		self.render_template('index.htm', **data)

app = webapp2.WSGIApplication([
	('/', Index)
], debug=True)