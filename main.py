from config.constants import isUnix, containerDemotionUid
import os
import subprocess
import sys

if isUnix and containerDemotionUid:
	uid = int(containerDemotionUid)
	os.system(f"chown -R {uid}:{uid} /app")
	os.setuid(uid) # pyright: ignore[reportGeneralTypeIssues,reportUnknownMemberType]
else:
	def parsePipPackages(packagesStr: str) -> dict[str, str]:
		return { packageSplit[0]: packageSplit[1] for packageSplit in [package.split("==") for package in packagesStr.splitlines()] }
	pipFreezeResult = subprocess.run([sys.executable, "-m", "pip", "freeze"], capture_output = True, text = True)
	installedPackages = parsePipPackages(pipFreezeResult.stdout)
	with open("requirements.txt", "r", encoding = "UTF-8") as requirementsFile:
		requiredPackages = parsePipPackages(requirementsFile.read())
	for packageName, packageVersion in requiredPackages.items():
		if packageName not in installedPackages:
			print(f"Required package '{packageName}' not found, installing...")
			subprocess.run([sys.executable, "-m", "pip", "install", f"{packageName}=={packageVersion}"], check = True)

from config.constants import dataDirectoryPath, logFilePath, name, plexClientID, version, isInteractive
from core.config import config, loadConfig, saveConfig
from core.discord import DiscordIpcService
from core.plex import PlexAlertListener
from typing import Optional
from utils.cache import loadCache
from utils.logging import formatter, logger
from utils.text import formatSeconds
import logging
import models.config
import requests
import time
import urllib.parse

def main() -> None:
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
	if not config["users"]:
		logger.info("No users found in the config file")
		user = authUser()
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

def authUser() -> Optional[models.config.User]:
	response = requests.post("https://plex.tv/api/v2/pins.json?strong=true", headers = {
		"X-Plex-Product": name,
		"X-Plex-Client-Identifier": plexClientID,
	}).json()
	logger.info("Open the below URL in your web browser and sign in:")
	logger.info("https://app.plex.tv/auth#?clientID=%s&code=%s&context%%5Bdevice%%5D%%5Bproduct%%5D=%s", plexClientID, response["code"], urllib.parse.quote(name))
	time.sleep(5)
	for i in range(35):
		logger.info(f"Checking whether authentication is successful... ({formatSeconds((i + 1) * 5)})")
		authCheckResponse = requests.get(f"https://plex.tv/api/v2/pins/{response['id']}.json?code={response['code']}", headers = {
			"X-Plex-Client-Identifier": plexClientID,
		}).json()
		if authCheckResponse["authToken"]:
			logger.info("Authentication successful")
			serverName = os.environ.get("PLEX_SERVER_NAME")
			if not serverName:
				serverName = input("Enter the name of the Plex Media Server you wish to connect to: ") if isInteractive else "ServerName"
			return { "token": authCheckResponse["authToken"], "servers": [{ "name": serverName }] }
		time.sleep(5)
	else:
		logger.info(f"Authentication timed out ({formatSeconds(180)})")

def testIpc() -> None:
	logger.info("Testing Discord IPC connection")
	discordIpcService = DiscordIpcService()
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
	if len(sys.argv) > 1 and sys.argv[1] == "test-ipc":
		testIpc()
	else:
		main()
