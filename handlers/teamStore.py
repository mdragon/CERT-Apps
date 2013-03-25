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
		
		self.render_template('roster/list.htm', **data)

class List(handlers.base.Base):
	@ndb.toplevel
	def get(self):
		logging.debug('running List.get')
		data = self.commonData()
		self.noCache()
		member = data['member']
		
		#logging.debug('json ' + self.toJSON(data));
		if data['user'] == None:
			return self.redir(data['loginURL'])
		else:
			teams = models.common.Membership.query(models.common.Membership.member == member.key).fetch(999)
#			teams = member.membership_set
			team = None
			for t in teams:
				team = t.team
				break
			
			members = []
			membersJSON = []
			if not team is None:			
				membersQ = models.common.Membership.query()
				membersQ.filter(models.common.Membership.team == team)
				membersData = membersQ.fetch(999)
			
				#logging.debug('team name ' + team.name)
				logging.debug('membersData len ' + str(len(membersData)))
			
				membersQ.filter(models.common.Membership.member == member.key)
				membershipData = membersQ.fetch(999)			
				leaderCheck = None
				
				for ms in membershipData:
					logging.debug('ms' + ms.toJSON())
					leaderCheck = ms
					break
				
				for ms in membersData:
					msMember = ms.member.get()
					logging.debug('msMember ' + msMember.toJSON())
					filteredMember = msMember.filterForRoster(leaderCheck)
					filteredMember.id = filteredMember.key.id()
					members.append(filteredMember)
					membersJSON.append(filteredMember.toJSON())
					team.member_ids.append(msMember.key.urlsafe())
			else:
				logging.debug('team is none for memeber ' + member.toJSON())

			if team is None:
				teams = models.common.Team.query(models.common.Team.name=="Default Team").fetch(1)
				for t in teams:
						team = t
						break

				if team is None:
					team = models.common.Team()
					team.name = "Default Team"
					team.put()

				member.id = member.key.id()
				members.append(member)
				membersJSON.append(member.toJSON())
				team.member_ids.append('abc')
				team.member_ids.append(member.key.urlsafe())
				logging.debug(member.key.urlsafe())

			pieces = dict()
			pieces['teams'] = team
			pieces['members'] = members

			data['members']	= pieces	
			data['json']	= '{ "members": [' + ','.join(membersJSON) + '], "teams": [' + team.toJSON().replace('{"key"', '{  "member_ids": [' + str(member.key.id()) + '], "key"') +'] }'
			#data['json'] = self.toJSON(pieces)
			#data['json'] = '{ "teams": [' + team.toJSON() +'] }'
			#data['json'] = '{ "team": ' + team.toJSON() +' }'

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
	('/teams', List),
	('/roster/.*', Index),
], debug=True)