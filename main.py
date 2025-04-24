from config.constants import isInContainer, runtimeDirectory, uid, gid, containerCwd, noRuntimeDirChown
from utils.logging import logger
import os
import sys

if isInContainer:
	if not os.path.isdir(runtimeDirectory):
		logger.error(f"Runtime directory does not exist. Ensure that it is mounted into the container at {runtimeDirectory}")
		sys.exit(1)
	if os.geteuid() == 0: # pyright: ignore[reportAttributeAccessIssue,reportUnknownMemberType]
		if uid == -1 or gid == -1:
			logger.warning(f"Environment variable(s) DRPP_UID and/or DRPP_GID are/is not set. Manually ensure appropriate ownership of {runtimeDirectory}")
			statResult = os.stat(runtimeDirectory)
			uid, gid = statResult.st_uid, statResult.st_gid
		else:
			if noRuntimeDirChown:
				logger.warning(f"Environment variable DRPP_NO_RUNTIME_DIR_CHOWN is set to true. Manually ensure appropriate ownership of {runtimeDirectory}")
			else:
				os.system(f"chmod 700 {runtimeDirectory}")
				os.system(f"chown -R {uid}:{gid} {runtimeDirectory}")
		os.system(f"chown -R {uid}:{gid} {containerCwd}")
		os.setgid(gid) # pyright: ignore[reportAttributeAccessIssue,reportUnknownMemberType]
		os.setuid(uid) # pyright: ignore[reportAttributeAccessIssue,reportUnknownMemberType]
	else:
		logger.warning("Not running as the superuser. Manually ensure appropriate ownership of mounted contents")

from config.constants import noPipInstall

if not noPipInstall:
	try:
		import subprocess
		def parsePipPackages(packagesStr: str) -> dict[str, str]:
			return { packageSplit[0].lower(): packageSplit[1] if len(packageSplit) > 1 else "" for packageSplit in [package.split("==") for package in packagesStr.splitlines()] }
		pipFreezeResult = subprocess.run([sys.executable, "-m", "pip", "freeze"], stdout = subprocess.PIPE, text = True, check = True)
		installedPackages = parsePipPackages(pipFreezeResult.stdout)
		with open("requirements.txt", "r", encoding = "UTF-8") as requirementsFile:
			requiredPackages = parsePipPackages(requirementsFile.read())
		for packageName, requiredPackageVersion in requiredPackages.items():
			installedPackageVersion = installedPackages.get(packageName, "none")
			if installedPackageVersion != requiredPackageVersion:
				logger.info(f"Installing dependency: {packageName} (required: {requiredPackageVersion}, installed: {installedPackageVersion})")
				subprocess.run([sys.executable, "-m", "pip", "install", "-U", f"{packageName}=={requiredPackageVersion}"], check = True)
	except:
		logger.exception("An unexpected error occured during automatic installation of dependencies. Install them manually by running the following command: python -m pip install -U -r requirements.txt")

from config.constants import dataDirectoryPath, logFilePath, name, version, isInteractive, plexServerNameInput
from core.config import config, loadConfig, saveConfig
from core.discord import DiscordIpcService
from core.imgur import uploadToImgur
from core.plex import PlexAlertListener, initiateAuth, getAuthToken
from models.discord import ActivityType
from typing import Optional
from utils.cache import loadCache
from utils.logging import formatter
from utils.text import formatSeconds
import logging
import models.config
import time

def init() -> None:
	if not os.path.isdir(dataDirectoryPath):
		os.makedirs(dataDirectoryPath)
	for oldFilePath in ["config.json", "cache.json", "console.log"]:
		if os.path.isfile(oldFilePath):
			os.rename(oldFilePath, os.path.join(dataDirectoryPath, oldFilePath))
	loadConfig()
	if config["logging"]["debug"]:
		logger.setLevel(logging.DEBUG)
	if config["logging"]["writeToFile"]:
		fileHandler = logging.FileHandler(logFilePath)
		fileHandler.setFormatter(formatter)
		logger.addHandler(fileHandler)
	logger.info("%s - v%s", name, version)
	loadCache()

def main() -> None:
	if not config["users"]:
		logger.info("No users found in the config file")
		user = authNewUser()
		if not user:
			sys.exit(1)
		config["users"].append(user)
		saveConfig()
	plexAlertListeners = [PlexAlertListener(user["token"], server) for user in config["users"] for server in user["servers"]]
	try:
		if isInteractive:
			while True:
				userInput = input()
				if userInput in ["exit", "quit"]:
					raise KeyboardInterrupt
				elif userInput == "reload-config":
					loadConfig()
					print("Config reloaded from file")
				else:
					print("Unrecognised command")
		else:
			while True:
				time.sleep(3600)
	except KeyboardInterrupt:
		for plexAlertListener in plexAlertListeners:
			plexAlertListener.disconnect()

def authNewUser() -> Optional[models.config.User]:
	id, code, url = initiateAuth()
	logger.info("Open the below URL in your web browser and sign in:")
	logger.info(url)
	time.sleep(5)
	for i in range(35):
		logger.info(f"Checking whether authentication is successful ({formatSeconds((i + 1) * 5)})")
		authToken = getAuthToken(id, code)
		if authToken:
			logger.info("Authentication successful")
			serverName = plexServerNameInput
			if not serverName:
				if isInteractive:
					serverName = input("Enter the name of the Plex Media Server to connect to: ")
				else:
					serverName = "ServerName"
					logger.warning("Environment variable DRPP_PLEX_SERVER_NAME_INPUT is not set and the environment is non-interactive")
					logger.warning("\"ServerName\" will be used as a placeholder for the name of the Plex Media Server to connect to")
					logger.warning("Change this by editing the config file and restarting the script")
			return { "token": authToken, "servers": [{ "name": serverName }] }
		time.sleep(5)
	else:
		logger.info(f"Authentication timed out ({formatSeconds(180)})")

def testIpc(pipeNumber: int) -> None:
	init()
	logger.info("Testing Discord IPC connection")
	currentTimestamp = int(time.time() * 1000)
	discordIpcService = DiscordIpcService(pipeNumber)
	discordIpcService.connect()
	discordIpcService.setActivity({
		"type": ActivityType.WATCHING,
		"details": "details",
		"state": "state",
		"assets": {
			"large_text": "large_text",
			"large_image": uploadToImgur("https://placehold.co/256x256/EEE/333.png") or "large_text",
			"small_text": "small_text",
			"small_image": "playing",
		},
		"timestamps": {
			"start": currentTimestamp,
			"end": currentTimestamp + 15000,
		},
		"buttons": [
			{
				"label": "Label",
				"url": "https://placehold.co/"
			},
		],
	})
	time.sleep(15)
	discordIpcService.disconnect()

if __name__ == "__main__":
	mode = sys.argv[1] if len(sys.argv) > 1 else ""
	try:
		if not mode:
			init()
			main()
		elif mode == "test-ipc":
			testIpc(int(sys.argv[2]) if len(sys.argv) > 2 else -1)
		else:
			logger.error(f"Invalid mode: {mode}")
	except KeyboardInterrupt:
		pass
