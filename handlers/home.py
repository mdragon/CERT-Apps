import webapp2
import models
import logging

class MainPage(webapp2.RequestHandler):
	def get(self):
		self.response.headers['Content-Type'] = 'text/plain'
		self.response.out.write('Hello, webapp World!')
		m = models.Member()
		m.email = 'mdragon@gmail.com'
		m.address = '13 Sheridan Ave. West Orange, NJ, 07052'
		
		logging.debug('m', m.toJSON())

app = webapp2.WSGIApplication([('/', MainPage)], debug=True)