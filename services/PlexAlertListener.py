# type: ignore

from plexapi.alert import AlertListener
from plexapi.myplex import MyPlexAccount
from services import DiscordRpcService
from utils.logs import LoggerWithPrefix
from utils.text import formatSeconds
import hashlib
import threading
import time

class PlexAlertListener:

	productName = "Plex Media Server"
	updateTimeoutTimerInterval = 30
	connectionTimeoutTimerInterval = 60
	maximumIgnores = 2
	useRemainingTime = False

	def __init__(self, token, serverConfig):
		self.token = token
		self.serverConfig = serverConfig
		self.logger = LoggerWithPrefix(f"[{self.serverConfig['name']}/{hashlib.md5(str(id(self)).encode('UTF-8')).hexdigest()[:5].upper()}] ")
		self.discordRpcService = DiscordRpcService()
		self.updateTimeoutTimer = None
		self.connectionTimeoutTimer = None
		self.reset()
		self.connect()

	def reset(self):
		self.plexAccount = None
		self.listenForUser = ""
		self.plexServer = None
		self.isServerOwner = False
		self.plexAlertListener = None
		self.lastState = ""
		self.lastSessionKey = 0
		self.lastRatingKey = 0
		self.ignoreCount = 0

	def connect(self):
		connected = False
		while not connected:
			try:
				self.plexAccount = MyPlexAccount(token = self.token)
				self.logger.info("Signed in as Plex User \"%s\"", self.plexAccount.username)
				self.listenForUser = self.serverConfig.get("listenForUser", self.plexAccount.username)
				self.plexServer = None
				for resource in self.plexAccount.resources():
					if resource.product == self.productName and resource.name.lower() == self.serverConfig["name"].lower():
						self.logger.info("Connecting to %s \"%s\"", self.productName, self.serverConfig["name"])
						self.plexServer = resource.connect()
						try:
							self.plexServer.account()
							self.isServerOwner = True
						except:
							pass
						self.logger.info("Connected to %s \"%s\"", self.productName, resource.name)
						self.plexAlertListener = AlertListener(self.plexServer, self.handlePlexAlert, self.reconnect)
						self.plexAlertListener.start()
						self.logger.info("Listening for alerts from user \"%s\"", self.listenForUser)
						self.connectionTimeoutTimer = threading.Timer(self.connectionTimeoutTimerInterval, self.connectionTimeout)
						self.connectionTimeoutTimer.start()
						connected = True
						break
				if not self.plexServer:
					self.logger.error("%s \"%s\" not found", self.productName, self.serverConfig["name"])
					break
			except Exception as e:
				self.logger.error("Failed to connect to %s \"%s\": %s", self.productName, self.serverConfig["name"], e)
				self.logger.error("Reconnecting in 10 seconds")
				time.sleep(10)

	def disconnect(self):
		self.discordRpcService.disconnect()
		self.cancelTimers()
		try:
			self.plexAlertListener.stop()
		except:
			pass
		self.reset()
		self.logger.info("Stopped listening for alerts")

	def reconnect(self, exception):
		self.logger.error("Connection to Plex lost: %s", exception)
		self.disconnect()
		self.logger.error("Reconnecting")
		self.connect()

	def cancelTimers(self):
		if self.updateTimeoutTimer:
			self.updateTimeoutTimer.cancel()
			self.updateTimeoutTimer = None
		if self.connectionTimeoutTimer:
			self.connectionTimeoutTimer.cancel()
			self.connectionTimeoutTimer = None

	def updateTimeout(self):
		self.logger.debug("No recent updates from session key %s", self.lastSessionKey)
		self.discordRpcService.disconnect()
		self.cancelTimers()

	def connectionTimeout(self):
		try:
			self.logger.debug("Request for list of clients to check connection: %s", self.plexServer.clients())
		except Exception as e:
			self.reconnect(e)
		else:
			self.connectionTimeoutTimer = threading.Timer(self.connectionTimeoutTimerInterval, self.connectionTimeout)
			self.connectionTimeoutTimer.start()

	def handlePlexAlert(self, data):
		try:
			if data["type"] == "playing" and "PlaySessionStateNotification" in data:
				alert = data["PlaySessionStateNotification"][0]
				state = alert["state"]
				sessionKey = int(alert["sessionKey"])
				ratingKey = int(alert["ratingKey"])
				viewOffset = int(alert["viewOffset"])
				self.logger.debug("Received alert: %s", alert)
				item = self.plexServer.fetchItem(ratingKey)
				libraryName = item.section().title
				if "blacklistedLibraries" in self.serverConfig and libraryName in self.serverConfig["blacklistedLibraries"]:
					self.logger.debug("Library \"%s\" is blacklisted, ignoring", libraryName)
					return
				if "whitelistedLibraries" in self.serverConfig and libraryName not in self.serverConfig["whitelistedLibraries"]:
					self.logger.debug("Library \"%s\" is not whitelisted, ignoring", libraryName)
					return
				if self.lastSessionKey == sessionKey and self.lastRatingKey == ratingKey:
					if self.updateTimeoutTimer:
						self.updateTimeoutTimer.cancel()
						self.updateTimeoutTimer = None
					if self.lastState == state and self.ignoreCount < self.maximumIgnores:
						self.logger.debug("Nothing changed, ignoring")
						self.ignoreCount += 1
						self.updateTimeoutTimer = threading.Timer(self.updateTimeoutTimerInterval, self.updateTimeout)
						self.updateTimeoutTimer.start()
						return
					else:
						self.ignoreCount = 0
						if state == "stopped":
							self.lastState, self.lastSessionKey, self.lastRatingKey = None, None, None
							self.discordRpcService.disconnect()
							self.cancelTimers()
							return
				elif state == "stopped":
					self.logger.debug("Received \"stopped\" state alert from unknown session key, ignoring")
					return
				if self.isServerOwner:
					self.logger.debug("Searching sessions for session key %s", sessionKey)
					plexServerSessions = self.plexServer.sessions()
					if len(plexServerSessions) < 1:
						self.logger.debug("Empty session list, ignoring")
						return
					for session in plexServerSessions:
						self.logger.debug("%s, Session Key: %s, Usernames: %s", session, session.sessionKey, session.usernames)
						if session.sessionKey == sessionKey:
							self.logger.debug("Session found")
							sessionUsername = session.usernames[0]
							if sessionUsername.lower() == self.listenForUser.lower():
								self.logger.debug("Username \"%s\" matches \"%s\", continuing", sessionUsername, self.listenForUser)
								break
							self.logger.debug("Username \"%s\" doesn't match \"%s\", ignoring", sessionUsername, self.listenForUser)
							return
					else:
						self.logger.debug("No matching session found, ignoring")
						return
				if self.updateTimeoutTimer:
					self.updateTimeoutTimer.cancel()
				self.updateTimeoutTimer = threading.Timer(self.updateTimeoutTimerInterval, self.updateTimeout)
				self.updateTimeoutTimer.start()
				self.lastState, self.lastSessionKey, self.lastRatingKey = state, sessionKey, ratingKey
				if state != "playing":
					stateText = f"{formatSeconds(viewOffset / 1000, ':')} / {formatSeconds(item.duration / 1000, ':')}"
				else:
					stateText = formatSeconds(item.duration / 1000)
				mediaType = item.type
				if mediaType == "movie":
					title = f"{item.title} ({item.year})"
					if len(item.genres) > 0:
						stateText += f" · {', '.join(genre.tag for genre in item.genres[:3])}"
					largeText = "Watching a movie"
					# self.logger.debug("Poster: %s", item.thumbUrl)
				elif mediaType == "episode":
					title = item.grandparentTitle
					stateText += f" · S{item.parentIndex:02}E{item.index:02} - {item.title}"
					largeText = "Watching a TV show"
					# self.logger.debug("Poster: %s", self.plexServer.url(item.grandparentThumb, True))
				elif mediaType == "track":
					title = item.title
					artist = item.originalTitle
					if not artist:
						artist = item.grandparentTitle
					stateText = f"{artist} - {item.parentTitle}"
					largeText = "Listening to music"
					# self.logger.debug("Album Art: %s", item.thumbUrl)
				else:
					self.logger.debug("Unsupported media type \"%s\", ignoring", mediaType)
					return
				activity = {
					"details": title[:128],
					"state": stateText[:128],
					"assets": {
						"large_text": largeText,
						"large_image": "logo",
						"small_text": state.capitalize(),
						"small_image": state,
					},
				}
				if state == "playing":
					currentTimestamp = int(time.time())
					if self.useRemainingTime:
						activity["timestamps"] = {"end": round(currentTimestamp + ((item.duration - viewOffset) / 1000))}
					else:
						activity["timestamps"] = {"start": round(currentTimestamp - (viewOffset / 1000))}
				if not self.discordRpcService.connected:
					self.discordRpcService.connect()
				if self.discordRpcService.connected:
					self.discordRpcService.sendActivity(activity)
		except:
			self.logger.exception("An unexpected error occured in the alert handler")
