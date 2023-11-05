from config.constants import discordClientID, isUnix, processID
from typing import Any, Optional
from utils.logging import logger
import asyncio
import json
import models.discord
import os
import struct
import time

class DiscordIpcService:

	def __init__(self):
		self.ipcPipe = (("/run/app" if os.path.isdir("/run/app") else os.environ.get("XDG_RUNTIME_DIR", os.environ.get("TMPDIR", os.environ.get("TMP", os.environ.get("TEMP", "/tmp"))))) + "/discord-ipc-0") if isUnix else r"\\?\pipe\discord-ipc-0"
		self.loop: asyncio.AbstractEventLoop = None # pyright: ignore[reportGeneralTypeIssues]
		self.pipeReader: asyncio.StreamReader = None # pyright: ignore[reportGeneralTypeIssues]
		self.pipeWriter: asyncio.StreamWriter = None # pyright: ignore[reportGeneralTypeIssues]
		self.connected = False

	async def handshake(self) -> None:
		try:
			if isUnix:
				self.pipeReader, self.pipeWriter = await asyncio.open_unix_connection(self.ipcPipe) # pyright: ignore[reportGeneralTypeIssues,reportUnknownMemberType]
			else:
				self.pipeReader = asyncio.StreamReader()
				self.pipeWriter = (await self.loop.create_pipe_connection(lambda: asyncio.StreamReaderProtocol(self.pipeReader), self.ipcPipe))[0] # pyright: ignore[reportGeneralTypeIssues,reportUnknownMemberType]
			self.write(0, { "v": 1, "client_id": discordClientID })
			if await self.read():
				self.connected = True
		except:
			logger.exception("An unexpected error occured during an IPC handshake operation")

	async def read(self) -> Optional[Any]:
		try:
			dataBytes = await self.pipeReader.read(1024)
			data = json.loads(dataBytes[8:].decode("utf-8"))
			logger.debug("[READ] %s", data)
			return data
		except:
			logger.exception("An unexpected error occured during an IPC read operation")
			self.connected = False

	def write(self, op: int, payload: Any) -> None:
		try:
			logger.debug("[WRITE] %s", payload)
			payload = json.dumps(payload)
			self.pipeWriter.write(struct.pack("<ii", op, len(payload)) + payload.encode("utf-8"))
		except:
			logger.exception("An unexpected error occured during an IPC write operation")
			self.connected = False

	def connect(self) -> None:
		if self.connected:
			logger.debug("Attempt to connect Discord IPC pipe while already connected")
			return
		logger.info("Connecting Discord IPC pipe")
		self.loop = asyncio.new_event_loop()
		self.loop.run_until_complete(self.handshake())

	def disconnect(self) -> None:
		if not self.connected:
			logger.debug("Attempt to disconnect Discord IPC pipe while not connected")
			return
		logger.info("Disconnecting Discord IPC pipe")
		try:
			self.pipeWriter.close()
		except:
			logger.exception("An unexpected error occured while closing the IPC pipe writer")
		try:
			self.loop.run_until_complete(self.pipeReader.read())
		except:
			logger.exception("An unexpected error occured while closing the IPC pipe reader")
		try:
			self.loop.close()
		except:
			logger.exception("An unexpected error occured while closing the asyncio event loop")
		self.connected = False

	def setActivity(self, activity: models.discord.Activity) -> None:
		logger.info("Activity update: %s", activity)
		payload = {
			"cmd": "SET_ACTIVITY",
			"args": {
				"pid": processID,
				"activity": activity,
			},
			"nonce": "{0:.2f}".format(time.time()),
		}
		self.write(1, payload)
		self.loop.run_until_complete(self.read())
