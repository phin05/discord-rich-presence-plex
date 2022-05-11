from services import ConfigService, PlexAlertListener
from store.constants import isUnix, name, plexClientID, version
from utils.logs import logger
import logging
import os
import requests
import time
import urllib.parse

os.system("clear" if isUnix else "cls")
logger.info("%s - v%s", name, version)
configService = ConfigService("config.json")
config = configService.config
if config["logging"]["debug"]:
	logger.setLevel(logging.DEBUG)
PlexAlertListener.useRemainingTime = config["display"]["useRemainingTime"]

if len(config["users"]) == 0:
	logger.info("No users found in the config file. Initiating authentication flow.")
	response = requests.post("https://plex.tv/api/v2/pins.json?strong=true", headers = {
		"X-Plex-Product": name,
		"X-Plex-Client-Identifier": plexClientID,
	}).json()
	logger.info("Open the below URL in your web browser and sign in:")
	logger.info("https://app.plex.tv/auth#?clientID=%s&code=%s&context%%5Bdevice%%5D%%5Bproduct%%5D=%s", plexClientID, response["code"], urllib.parse.quote(name))
	time.sleep(5)
	logger.info("Checking whether authentication is successful...")
	for _ in range(120):
		authCheckResponse = requests.get(f"https://plex.tv/api/v2/pins/{response['id']}.json?code={response['code']}", headers = {
			"X-Plex-Client-Identifier": plexClientID,
		}).json()
		if authCheckResponse["authToken"]:
			logger.info("Authentication successful.")
			serverName = input("Enter the name of the Plex Media Server you wish to connect to: ")
			config["users"].append({ "token": authCheckResponse["authToken"], "servers": [{ "name": serverName }] })
			configService.saveConfig()
			break
		time.sleep(5)
	else:
		logger.info("Authentication failed.")
		exit()

plexAlertListeners: list[PlexAlertListener] = []
try:
	for user in config["users"]:
		for server in user["servers"]:
			plexAlertListeners.append(PlexAlertListener(user["token"], server))
	while True:
		userInput = input()
		if userInput in ["exit", "quit"]:
			break
except KeyboardInterrupt:
	for plexAlertListener in plexAlertListeners:
		plexAlertListener.disconnect()
except:
	logger.exception("An unexpected error occured")
