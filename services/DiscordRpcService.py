# type: ignore

from store.constants import isUnix, processID
from utils.logs import logger
import asyncio
import json
import os
import struct
import time

class DiscordRpcService:

	clientID = "413407336082833418"
	ipcPipe = ((os.environ.get("XDG_RUNTIME_DIR", None) or os.environ.get("TMPDIR", None) or os.environ.get("TMP", None) or os.environ.get("TEMP", None) or "/tmp") + "/discord-ipc-0") if isUnix else r"\\?\pipe\discord-ipc-0"

	def __init__(self):
		self.loop = None
		self.pipeReader = None
		self.pipeWriter = None
		self.connected = False

	def connect(self):
		logger.info("Connecting Discord IPC Pipe")
		self.loop = asyncio.new_event_loop() if isUnix else asyncio.ProactorEventLoop()
		self.loop.run_until_complete(self.handshake())

	async def handshake(self):
		try:
			if isUnix:
				self.pipeReader, self.pipeWriter = await asyncio.open_unix_connection(self.ipcPipe, loop = self.loop)
			else:
				self.pipeReader = asyncio.StreamReader(loop = self.loop)
				self.pipeWriter, _ = await self.loop.create_pipe_connection(lambda: asyncio.StreamReaderProtocol(self.pipeReader, loop = self.loop), self.ipcPipe)
			self.write(0, { "v": 1, "client_id": self.clientID })
			if await self.read():
				self.connected = True
		except:
			logger.exception("An unexpected error occured during a RPC handshake operation")

	async def read(self):
		try:
			dataBytes = await self.pipeReader.read(1024)
			data = json.loads(dataBytes[8:].decode("utf-8"))
			logger.debug("[READ] %s", data)
			return data
		except:
			logger.exception("An unexpected error occured during a RPC read operation")
			self.disconnect()

	def write(self, op, payload):
		try:
			logger.debug("[WRITE] %s", payload)
			payload = json.dumps(payload)
			self.pipeWriter.write(struct.pack("<ii", op, len(payload)) + payload.encode("utf-8"))
		except:
			logger.exception("An unexpected error occured during a RPC write operation")
			self.disconnect()

	def disconnect(self):
		logger.info("Disconnecting Discord IPC Pipe")
		if (self.pipeWriter):
			try:
				self.pipeWriter.close()
			except:
				logger.exception("An unexpected error occured while closing an IPC pipe writer")
			self.pipeWriter = None
		if (self.pipeReader):
			try:
				self.loop.run_until_complete(self.pipeReader.read())
			except:
				logger.exception("An unexpected error occured while closing an IPC pipe reader")
			self.pipeReader = None
		try:
			self.loop.close()
		except:
			logger.exception("An unexpected error occured while closing an asyncio event loop")
		self.connected = False

	def sendActivity(self, activity):
		logger.info("Activity update: %s", activity)
		payload = {
			"cmd": "SET_ACTIVITY",
			"args": {
				"pid": processID,
				"activity": activity,
			},
			"nonce": "{0:.20f}".format(time.time())
		}
		self.write(1, payload)
		self.loop.run_until_complete(self.read())
