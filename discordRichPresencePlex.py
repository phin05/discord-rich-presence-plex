import asyncio
import json
import os
import plexapi
import plexapi.myplex
import struct
import subprocess
import sys
import tempfile
import threading
import time

isLinux = sys.platform in ["linux", "darwin"]

if (isLinux):
	os.system("clear")
else:
	os.system("cls")

class discordRichPresence:

	def __init__(self, clientID):
		self.IPCPipe = ((os.environ.get("XDG_RUNTIME_DIR", None) or os.environ.get("TMPDIR", None) or os.environ.get("TMP", None) or os.environ.get("TEMP", None) or "/tmp") + "/discord-ipc-0") if isLinux else "\\\\?\\pipe\\discord-ipc-0"
		self.clientID = clientID
		self.pipeReader = None
		self.pipeWriter = None
		self.process = None
		self.running = False

	async def read(self):
		print("\nReading:")
		data = await self.pipeReader.read(1024)
		print(json.loads(data[8:].decode("utf-8")))

	def write(self, op, payload):
		print("\nWriting:")
		payload = json.dumps(payload)
		print(payload)
		data = self.pipeWriter.write(struct.pack("<ii", op, len(payload)) + payload.encode("utf-8"))

	async def handshake(self):
		if (isLinux):
			self.pipeReader, self.pipeWriter = await asyncio.open_unix_connection(self.IPCPipe, loop = self.loop)
		else:
			self.pipeReader = asyncio.StreamReader(loop = self.loop)
			self.pipeWriter, _ = await self.loop.create_pipe_connection(lambda: asyncio.StreamReaderProtocol(self.pipeReader, loop = self.loop), self.IPCPipe)
		self.write(0, {"v": 1, "client_id": self.clientID})
		await self.read()
		self.running = True

	def start(self):
		print("\nOpening Discord IPC Pipe")
		emptyProcessFilePath = tempfile.gettempdir() + "\\discordRichPresencePlex-emptyProcess.py"
		if (not os.path.exists(emptyProcessFilePath)):
			with open(emptyProcessFilePath, "w") as emptyProcessFile:
				emptyProcessFile.write("import time\n\ntry:\n\twhile (True):\n\t\ttime.sleep(60)\nexcept:\n\tpass")
		self.process = subprocess.Popen(["pythonw", emptyProcessFilePath])
		self.loop = asyncio.get_event_loop() if isLinux else asyncio.ProactorEventLoop()
		self.loop.run_until_complete(self.handshake())

	def stop(self):
		print("\nClosing Discord IPC Pipe")
		self.process.kill()
		self.pipeWriter.close()
		self.loop.close()
		self.running = False

	def send(self, activity):
		payload = {
			"cmd": "SET_ACTIVITY",
			"args": {
				"activity": activity,
				"pid": self.process.pid
			},
			"nonce": "{0:.20f}".format(time.time())
		}
		sent = self.write(1, payload)
		self.loop.run_until_complete(self.read())

class discordRichPresencePlex(discordRichPresence):

	plexServerName = ""
	plexUsername = ""
	plexPasswordOrToken = ""
	usingToken = False
	listenForUser = ""
	listenForUser = plexUsername if listenForUser == "" else listenForUser

	productName = "Plex Media Server"
	lastState = None
	lastSessionKey = None
	lastRatingKey = None
	stopTimer = None

	def __init__(self):
		super().__init__("413407336082833418")

	def run(self):
		if (self.usingToken):
			self.plexAccount = plexapi.myplex.MyPlexAccount(self.plexUsername, token = self.plexPasswordOrToken)
		else:
			self.plexAccount = plexapi.myplex.MyPlexAccount(self.plexUsername, self.plexPasswordOrToken)
		print("Logged in as Plex User \"" + self.plexAccount.username + "\"")
		self.plexServer = None
		for resource in self.plexAccount.resources():
			if (resource.product == self.productName and resource.name == self.plexServerName):
				self.plexServer = resource.connect()
				self.plexServer.startAlertListener(self.onPlexServerAlert)
				print("Connected to " + self.productName + " \"" + self.plexServerName + "\"")
				print("Listening for PlaySessionStateNotification alerts from user \"" + self.listenForUser + "\"")
				break
		if (not self.plexServer):
			print(self.productName + " \"" + self.plexServerName + "\" not found")
			sys.exit(0)

	def onPlexServerAlert(self, data):
		try:
			if (data["type"] == "playing" and "PlaySessionStateNotification" in data):
				sessionData = data["PlaySessionStateNotification"][0]
				state = sessionData["state"]
				sessionKey = int(sessionData["sessionKey"])
				ratingKey = int(sessionData["ratingKey"])
				viewOffset = int(sessionData["viewOffset"])
				for session in self.plexServer.sessions():
					if (session.sessionKey == sessionKey):
						if (session.usernames[0].lower() == self.listenForUser.lower()):
							break
						else:
							return
				if (self.lastSessionKey == sessionKey and self.lastRatingKey == ratingKey):
					if (self.lastState == state):
						return
					elif (state == "stopped"):
						self.lastState, self.lastSessionKey, self.lastRatingKey = None, None, None
						self.stopTimer = threading.Timer(5, self.stop)
						self.stopTimer.start()
						return
				elif (state == "stopped"):
					return
				if (self.stopTimer):
					self.stopTimer.cancel()
					self.stopTimer = None
				self.lastState, self.lastSessionKey, self.lastRatingKey = state, sessionKey, ratingKey
				metadata = self.plexServer.fetchItem(ratingKey)
				mediaType = metadata.type
				if (mediaType == "movie"):
					title = metadata.title + " (" + str(metadata.year) + ")"
					if (state != "playing"):
						extra = str(time.strftime("%H:%M:%S", time.gmtime(viewOffset / 1000))) + "/" + str(time.strftime("%H:%M:%S", time.gmtime(metadata.duration / 1000)))
					else:
						extra = str(time.strftime("%Hh%Mm", time.gmtime(metadata.duration / 1000)))
					extra = extra + " · " + ", ".join([genre.tag for genre in metadata.genres[:3]])
				elif (mediaType == "episode"):
					title = metadata.grandparentTitle
					extra = "S" + str(metadata.parentIndex) + " · E" + str(metadata.index) + " - " + metadata.title
				elif (mediaType == "track"):
					title = metadata.title
					extra = metadata.grandparentTitle + " · " + metadata.parentTitle
				else:
					return
				activity = {
					"details": title,
					"state": extra,
					"assets": {
						"large_text": "Plex",
						"large_image": "logo",
						"small_text": state.capitalize(),
						"small_image": state
					},
				}
				if (state == "playing"):
					activity["timestamps"] = {"start": int(time.time()) - (viewOffset / 1000)}
				if (not self.running):
					self.start()
				self.send(activity)
		except Exception as e:
			print("Error: " + str(e))

discordRichPresencePlexInstance = discordRichPresencePlex()
try:
	discordRichPresencePlexInstance.run()
	while True:
		time.sleep(60)
		continue
except KeyboardInterrupt:
	if (discordRichPresencePlexInstance.running):
		discordRichPresencePlexInstance.stop()
except Exception as e:
	print("Error: " + str(e))
