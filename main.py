from app import constants, logger
import os
import sys

if constants.isInContainer:
	if not os.path.isdir(constants.runtimeDirectory):
		logger.error(f"Runtime directory does not exist. Ensure that it is mounted into the container at {constants.runtimeDirectory}")
		sys.exit(1)
	if os.geteuid() == 0: # pyright: ignore[reportAttributeAccessIssue,reportUnknownMemberType]
		if constants.uid == -1 or constants.gid == -1:
			logger.warning(f"Environment variable(s) DRPP_UID and/or DRPP_GID are/is not set. Manually ensure appropriate ownership of {constants.runtimeDirectory}")
			statResult = os.stat(constants.runtimeDirectory)
			uid, gid = statResult.st_uid, statResult.st_gid
		else:
			if constants.noRuntimeDirChown:
				logger.warning(f"Environment variable DRPP_NO_RUNTIME_DIR_CHOWN is set to true. Manually ensure appropriate ownership of {constants.runtimeDirectory}")
			else:
				os.system(f"chmod 700 {constants.runtimeDirectory}")
				os.system(f"chown -R {constants.uid}:{constants.gid} {constants.runtimeDirectory}")
		os.system(f"chown -R {constants.uid}:{constants.gid} {constants.containerCwd}")
		os.setgid(constants.gid) # pyright: ignore[reportAttributeAccessIssue,reportUnknownMemberType]
		os.setuid(constants.uid) # pyright: ignore[reportAttributeAccessIssue,reportUnknownMemberType]
	else:
		logger.warning("Not running as the superuser. Manually ensure appropriate ownership of mounted contents")

if not constants.noPipInstall:
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

from app import cache, config, discord, images, plex
from typing import Optional
import logging
import time

def init() -> None:
	if not os.path.isdir(constants.dataDirectoryPath):
		os.makedirs(constants.dataDirectoryPath)
	for oldFilePath in ["config.json", "cache.json", "console.log"]:
		if os.path.isfile(oldFilePath):
			os.rename(oldFilePath, os.path.join(constants.dataDirectoryPath, oldFilePath))
	config.load()
	if config.config["logging"]["debug"]:
		logger.logger.setLevel(logging.DEBUG)
	if config.config["logging"]["writeToFile"]:
		fileHandler = logging.FileHandler(constants.logFilePath)
		fileHandler.setFormatter(logger.formatter)
		logger.logger.addHandler(fileHandler)
	logger.info("%s - v%s", constants.name, constants.version)
	cache.load()

def main() -> None:
	if not config.config["users"]:
		logger.info("No users found in the config file")
		user = authNewUser()
		if not user:
			sys.exit(1)
		config.config["users"].append(user)
		config.save()
	plexAlertListeners = [plex.PlexAlertListener(user["token"], server) for user in config.config["users"] for server in user["servers"]]
	try:
		if constants.isInteractive:
			while True:
				userInput = input()
				if userInput in ["exit", "quit"]:
					raise KeyboardInterrupt
				elif userInput == "reload-config":
					config.load()
					print("Config reloaded from file")
				else:
					print("Unrecognised command")
		else:
			while True:
				time.sleep(3600)
	except KeyboardInterrupt:
		for plexAlertListener in plexAlertListeners:
			plexAlertListener.disconnect()

def authNewUser() -> Optional[config.User]:
	id, code, url = plex.initiateAuth()
	logger.info("Open the below URL in your web browser and sign in:")
	logger.info(url)
	time.sleep(5)
	for i in range(35):
		logger.info(f"Checking whether authentication is successful ({(i + 1) * 5}s)")
		authToken = plex.getAuthToken(id, code)
		if authToken:
			logger.info("Authentication successful")
			serverName = constants.plexServerNameInput
			if not serverName:
				if constants.isInteractive:
					serverName = input("Enter the name of the Plex Media Server to connect to: ")
				else:
					serverName = "ServerName"
					logger.warning("Environment variable DRPP_PLEX_SERVER_NAME_INPUT is not set and the environment is non-interactive")
					logger.warning("\"ServerName\" will be used as a placeholder for the name of the Plex Media Server to connect to")
					logger.warning("Change this by editing the config file and restarting the script")
			return { "token": authToken, "servers": [{ "name": serverName }] }
		time.sleep(5)
	else:
		logger.info(f"Authentication timed out (180s)")

def testIpc(pipeNumber: int) -> None:
	init()
	logger.info("Testing Discord IPC connection")
	currentTimestamp = int(time.time() * 1000)
	discordIpcService = discord.DiscordIpcService(pipeNumber)
	for i in range(1, 501):
		if not discordIpcService.connected:
			discordIpcService.connect()
		discordIpcService.setActivity({
			"type": discord.ActivityType.WATCHING,
			"details": f"Iteration {i}",
			"state": "state",
			"assets": {
				"large_text": "large_text",
				"large_image": images.upload("key", "https://placehold.co/256x256/EEE/333.png") or "paused",
				"small_text": "small_text",
				"small_image": "playing",
			},
			"timestamps": {
				"start": currentTimestamp,
				"end": currentTimestamp + 3600000,
			},
			"buttons": [
				{
					"label": "Label",
					"url": "https://placehold.co/"
				},
			],
		})
		time.sleep(5)
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
