from config.constants import isUnix, containerDemotionUidGid
import os
import sys

if isUnix and containerDemotionUidGid:
	uidGid = int(containerDemotionUidGid)
	os.system(f"chown -R {uidGid}:{uidGid} {os.path.dirname(os.path.realpath(__file__))}")
	os.setgid(uidGid) # pyright: ignore[reportGeneralTypeIssues,reportUnknownMemberType]
	os.setuid(uidGid) # pyright: ignore[reportGeneralTypeIssues,reportUnknownMemberType]
else:
	try:
		import subprocess
		def parsePipPackages(packagesStr: str) -> dict[str, str]:
			return { packageSplit[0]: packageSplit[1] if len(packageSplit) > 1 else "" for packageSplit in [package.split("==") for package in packagesStr.splitlines()] }
		pipFreezeResult = subprocess.run([sys.executable, "-m", "pip", "freeze"], stdout = subprocess.PIPE, text = True, check = True)
		installedPackages = parsePipPackages(pipFreezeResult.stdout)
		with open("requirements.txt", "r", encoding = "UTF-8") as requirementsFile:
			requiredPackages = parsePipPackages(requirementsFile.read())
		for packageName, packageVersion in requiredPackages.items():
			if packageName not in installedPackages:
				package = f"{packageName}{f'=={packageVersion}' if packageVersion else ''}"
				print(f"Installing missing dependency: {package}")
				subprocess.run([sys.executable, "-m", "pip", "install", "-U", package], check = True)
	except Exception as e:
		import traceback
		traceback.print_exception(e)
		print("An unexpected error occured during automatic installation of dependencies. Install them manually by running the following command: python -m pip install -U -r requirements.txt")

from config.constants import dataDirectoryPath, logFilePath, name, version, isInteractive
from core.config import config, loadConfig, saveConfig
from core.discord import DiscordIpcService
from core.plex import PlexAlertListener, initiateAuth, getAuthToken
from typing import Optional
from utils.cache import loadCache
from utils.logging import formatter, logger
from utils.text import formatSeconds
import logging
import models.config
import time

def init() -> None:
	if not os.path.exists(dataDirectoryPath):
		os.mkdir(dataDirectoryPath)
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
	init()
	if not config["users"]:
		logger.info("No users found in the config file")
		user = authNewUser()
		if not user:
			exit()
		config["users"].append(user)
		saveConfig()
	plexAlertListeners = [PlexAlertListener(user["token"], server) for user in config["users"] for server in user["servers"]]
	try:
		if isInteractive:
			while True:
				userInput = input()
				if userInput in ["exit", "quit"]:
					raise KeyboardInterrupt
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
			serverName = os.environ.get("PLEX_SERVER_NAME")
			if not serverName:
				serverName = input("Enter the name of the Plex Media Server you wish to connect to: ") if isInteractive else "ServerName"
			return { "token": authToken, "servers": [{ "name": serverName }] }
		time.sleep(5)
	else:
		logger.info(f"Authentication timed out ({formatSeconds(180)})")

def testIpc(ipcPipeNumber: int) -> None:
	init()
	logger.info("Testing Discord IPC connection")
	discordIpcService = DiscordIpcService(ipcPipeNumber)
	discordIpcService.connect()
	discordIpcService.setActivity({
		"details": "details",
		"state": "state",
		"assets": {
			"large_text": "large_text",
			"large_image": "logo",
			"small_text": "small_text",
			"small_image": "playing",
		},
	})
	time.sleep(15)
	discordIpcService.disconnect()

if __name__ == "__main__":
	mode = sys.argv[1] if len(sys.argv) > 1 else ""
	try:
		if not mode:
			main()
		elif mode == "test-ipc":
			testIpc(int(sys.argv[2]) if len(sys.argv) > 2 else -1)
		else:
			print(f"Invalid mode: {mode}")
	except KeyboardInterrupt:
		pass
