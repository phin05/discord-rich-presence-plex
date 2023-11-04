# pyright: reportUnknownArgumentType=none,reportUnknownMemberType=none,reportUnknownVariableType=none

from .config import config
from .discord import DiscordIpcService
from .imgur import uploadToImgur
from plexapi.alert import AlertListener
from plexapi.base import Playable, PlexPartialObject
from plexapi.media import Genre, GuidTag
from plexapi.myplex import MyPlexAccount, PlexServer
from typing import Optional
from utils.cache import getCacheKey, setCacheKey
from utils.logging import LoggerWithPrefix
from utils.text import formatSeconds
import models.config
import models.discord
import models.plex
import threading
import time

class PlexAlertListener(threading.Thread):

	productName = "Plex Media Server"
	updateTimeoutTimerInterval = 30
	connectionTimeoutTimerInterval = 60
	maximumIgnores = 2

	def __init__(self, token: str, serverConfig: models.config.Server):
		super().__init__()
		self.daemon = True
		self.token = token
		self.serverConfig = serverConfig
		self.logger = LoggerWithPrefix(f"[{self.serverConfig['name']}] ") # pyright: ignore[reportTypedDictNotRequiredAccess]
		self.discordIpcService = DiscordIpcService()
		self.updateTimeoutTimer: Optional[threading.Timer] = None
		self.connectionTimeoutTimer: Optional[threading.Timer] = None
		self.account: Optional[MyPlexAccount] = None
		self.server: Optional[PlexServer] = None
		self.alertListener: Optional[AlertListener] = None
		self.lastState, self.lastSessionKey, self.lastRatingKey = "", 0, 0
		self.listenForUser, self.isServerOwner, self.ignoreCount = "", False, 0
		self.start()

	def run(self) -> None:
		connected = False
		while not connected:
			try:
				self.logger.info("Signing into Plex")
				self.account = MyPlexAccount(token = self.token)
				self.logger.info("Signed in as Plex user \"%s\"", self.account.username)
				self.listenForUser = self.serverConfig.get("listenForUser", "") or self.account.username
				self.server = None
				for resource in self.account.resources():
					if resource.product == self.productName and resource.name.lower() == self.serverConfig["name"].lower():
						self.logger.info("Connecting to %s \"%s\"", self.productName, self.serverConfig["name"])
						self.server = resource.connect()
						try:
							self.server.account()
							self.isServerOwner = True
						except:
							pass
						self.logger.info("Connected to %s \"%s\"", self.productName, resource.name)
						self.alertListener = AlertListener(self.server, self.handleAlert, self.reconnect)
						self.alertListener.start()
						self.logger.info("Listening for alerts from user \"%s\"", self.listenForUser)
						self.connectionTimeoutTimer = threading.Timer(self.connectionTimeoutTimerInterval, self.connectionTimeout)
						self.connectionTimeoutTimer.start()
						connected = True
						break
				if not self.server:
					self.logger.error("%s \"%s\" not found", self.productName, self.serverConfig["name"])
					break
			except Exception as e:
				self.logger.error("Failed to connect to %s \"%s\": %s", self.productName, self.serverConfig["name"], e) # pyright: ignore[reportTypedDictNotRequiredAccess]
				self.logger.error("Reconnecting in 10 seconds")
				time.sleep(10)

	def disconnect(self) -> None:
		if self.alertListener:
			try:
				self.alertListener.stop()
			except:
				pass
		self.disconnectRpc()
		self.account, self.server, self.alertListener, self.listenForUser, self.isServerOwner, self.ignoreCount = None, None, None, "", False, 0
		self.logger.info("Stopped listening for alerts")

	def reconnect(self, exception: Exception) -> None:
		self.logger.error("Connection to Plex lost: %s", exception)
		self.disconnect()
		self.logger.error("Reconnecting")
		self.run()

	def disconnectRpc(self) -> None:
		self.lastState, self.lastSessionKey, self.lastRatingKey = "", 0, 0
		self.discordIpcService.disconnect()
		self.cancelTimers()

	def cancelTimers(self) -> None:
		if self.updateTimeoutTimer:
			self.updateTimeoutTimer.cancel()
		if self.connectionTimeoutTimer:
			self.connectionTimeoutTimer.cancel()
		self.updateTimeoutTimer, self.connectionTimeoutTimer = None, None

	def updateTimeout(self) -> None:
		self.logger.debug("No recent updates from session key %s", self.lastSessionKey)
		self.disconnectRpc()

	def connectionTimeout(self) -> None:
		try:
			assert self.server
			self.logger.debug("Request for list of clients to check connection: %s", self.server.clients())
		except Exception as e:
			self.reconnect(e)
		else:
			self.connectionTimeoutTimer = threading.Timer(self.connectionTimeoutTimerInterval, self.connectionTimeout)
			self.connectionTimeoutTimer.start()

	def handleAlert(self, alert: models.plex.Alert) -> None:
		try:
			if alert["type"] == "playing" and "PlaySessionStateNotification" in alert:
				stateNotification = alert["PlaySessionStateNotification"][0]
				state = stateNotification["state"]
				sessionKey = int(stateNotification["sessionKey"])
				ratingKey = int(stateNotification["ratingKey"])
				viewOffset = int(stateNotification["viewOffset"])
				self.logger.debug("Received alert: %s", stateNotification)
				assert self.server
				item: PlexPartialObject = self.server.fetchItem(ratingKey)
				libraryName: str = item.section().title
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
							self.disconnectRpc()
							return
				elif state == "stopped":
					self.logger.debug("Received \"stopped\" state alert from unknown session, ignoring")
					return
				if self.isServerOwner:
					self.logger.debug("Searching sessions for session key %s", sessionKey)
					sessions: list[Playable] = self.server.sessions()
					if len(sessions) < 1:
						self.logger.debug("Empty session list, ignoring")
						return
					for session in sessions:
						self.logger.debug("%s, Session Key: %s, Usernames: %s", session, session.sessionKey, session.usernames)
						if session.sessionKey == sessionKey:
							self.logger.debug("Session found")
							sessionUsername: str = session.usernames[0]
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
				mediaType: str = item.type
				title: str
				thumb: str
				if mediaType in ["movie", "episode"]:
					stateStrings: list[str] = [] if config["display"]["hideTotalTime"] else [formatSeconds(item.duration / 1000)]
					if mediaType == "movie":
						title = f"{item.title} ({item.year})"
						genres: list[Genre] = item.genres[:3]
						stateStrings.append(f"{', '.join(genre.tag for genre in genres)}")
						largeText = "Watching a movie"
						thumb = item.thumb
					else:
						title = item.grandparentTitle
						stateStrings.append(f"S{item.parentIndex:02}E{item.index:02}")
						stateStrings.append(item.title)
						largeText = "Watching a TV show"
						thumb = item.grandparentThumb
					if state != "playing":
						if config["display"]["useRemainingTime"]:
							stateStrings.append(f"{formatSeconds((item.duration - viewOffset) / 1000, ':')} left")
						else:
							stateStrings.append(f"{formatSeconds(viewOffset / 1000, ':')} elapsed")
					stateText = " · ".join(stateString for stateString in stateStrings if stateString)
				elif mediaType == "track":
					title = item.title
					stateText = f"{item.originalTitle or item.grandparentTitle} - {item.parentTitle} ({self.server.fetchItem(item.parentRatingKey).year})"
					largeText = "Listening to music"
					thumb = item.thumb
				else:
					self.logger.debug("Unsupported media type \"%s\", ignoring", mediaType)
					return
				thumbUrl = ""
				if thumb and config["display"]["posters"]["enabled"]:
					thumbUrl = getCacheKey(thumb)
					if not thumbUrl:
						self.logger.debug("Uploading image to Imgur")
						thumbUrl = uploadToImgur(self.server.url(thumb, True))
						setCacheKey(thumb, thumbUrl)
				activity: models.discord.Activity = {
					"details": title[:128],
					"state": stateText[:128],
					"assets": {
						"large_text": largeText,
						"large_image": thumbUrl or "logo",
						"small_text": state.capitalize(),
						"small_image": state,
					},
				}
				if config["display"]["buttons"]:
					guidTags: list[GuidTag] = []
					if mediaType == "movie":
						guidTags = item.guids
					elif mediaType == "episode":
						guidTags = self.server.fetchItem(item.grandparentRatingKey).guids
					guids: list[str] = [guid.id for guid in guidTags]
					buttons: list[models.discord.ActivityButton] = []
					for button in config["display"]["buttons"]:
						if button["url"].startswith("dynamic:"):
							if guids:
								newUrl = button["url"]
								if button["url"] == "dynamic:imdb":
									for guid in guids:
										if guid.startswith("imdb://"):
											newUrl = guid.replace("imdb://", "https://www.imdb.com/title/")
											break
								elif button["url"] == "dynamic:tmdb":
									for guid in guids:
										if guid.startswith("tmdb://"):
											tmdbPathSegment = "movie" if mediaType == "movie" else "tv"
											newUrl = guid.replace("tmdb://", f"https://www.themoviedb.org/{tmdbPathSegment}/")
											break
								if newUrl:
									buttons.append({ "label": button["label"], "url": newUrl })
						else:
							buttons.append(button)
					if buttons:
						activity["buttons"] = buttons[:2]
				if state == "playing":
					currentTimestamp = int(time.time())
					if config["display"]["useRemainingTime"]:
						activity["timestamps"] = {"end": round(currentTimestamp + ((item.duration - viewOffset) / 1000))}
					else:
						activity["timestamps"] = {"start": round(currentTimestamp - (viewOffset / 1000))}
				if not self.discordIpcService.connected:
					self.discordIpcService.connect()
				if self.discordIpcService.connected:
					self.discordIpcService.setActivity(activity)
		except:
			self.logger.exception("An unexpected error occured in the Plex alert handler")
