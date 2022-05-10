from services import ConfigService, PlexAlertListener
from store.constants import isUnix
from utils.logs import logger
import logging
import os

os.system("clear" if isUnix else "cls")

configService = ConfigService("config.json")
config = configService.config

if config["logging"]["debug"]:
	logger.setLevel(logging.DEBUG)

PlexAlertListener.useRemainingTime = config["display"]["useRemainingTime"]

if len(config["users"]) == 0:
	logger.info("No users in config. Initiating authorisation flow. ! TBD !") # TODO
	exit()

plexAlertListeners: list[PlexAlertListener] = []
try:
	for user in config["users"]:
		for server in user["servers"]:
			plexAlertListeners.append(PlexAlertListener(user["username"], user["token"], server))
	while True:
		userInput = input()
		if userInput in ["exit", "quit"]:
			break
except KeyboardInterrupt:
	for plexAlertListener in plexAlertListeners:
		plexAlertListener.disconnect()
except:
	logger.exception("An unexpected error occured")
