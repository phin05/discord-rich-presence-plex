from app import constants, logger
from typing import Any
from typing import TypedDict
import json
import os
import sys
import yaml

class Logging(TypedDict):
	debug: bool
	writeToFile: bool

class Posters(TypedDict):
	enabled: bool
	imgurClientID: str
	maxSize: int
	fit: bool

class Button(TypedDict):
	label: str
	url: str
	mediaTypes: list[str]

class Display(TypedDict):
	duration: bool
	genres: bool
	album: bool
	albumImage: bool
	artist: bool
	artistImage: bool
	year: bool
	statusIcon: bool
	progressMode: str
	paused: bool
	posters: Posters
	buttons: list[Button]

class Server(TypedDict, total = False):
	name: str
	listenForUser: str
	blacklistedLibraries: list[str]
	whitelistedLibraries: list[str]
	ipcPipeNumber: int

class User(TypedDict):
	token: str
	servers: list[Server]

class Config(TypedDict):
	logging: Logging
	display: Display
	users: list[User]

config: Config = {
	"logging": {
		"debug": True,
		"writeToFile": False,
	},
	"display": {
		"duration": False,
		"genres": True,
		"album": True,
		"albumImage": True,
		"artist": True,
		"artistImage": True,
		"year": True,
		"statusIcon": False,
		"progressMode": "bar",
		"paused": False,
		"posters": {
			"enabled": True,
			"imgurClientID": "",
			"maxSize": 256,
			"fit": True,
		},
		"buttons": [],
	},
	"users": [],
}
supportedConfigFileExtensions = {
	"yaml": "yaml",
	"yml": "yaml",
	"json": "json",
}
configFileExtension = ""
configFileType = ""
configFilePath = ""

def load() -> None:
	global configFileExtension, configFileType, configFilePath
	doesFileExist = False
	for i, (fileExtension, fileType) in enumerate(supportedConfigFileExtensions.items()):
		doesFileExist = os.path.isfile(f"{constants.configFilePathBase}.{fileExtension}")
		isFirstItem = i == 0
		if doesFileExist or isFirstItem:
			configFileExtension = fileExtension
			configFileType = fileType
			configFilePath = f"{constants.configFilePathBase}.{configFileExtension}"
			if doesFileExist:
				break
	if doesFileExist:
		try:
			with open(configFilePath, "r", encoding = "UTF-8") as configFile:
				if configFileType == "yaml":
					loadedConfig = yaml.safe_load(configFile) or {} # pyright: ignore[reportUnknownVariableType]
				else:
					loadedConfig = json.load(configFile) or {} # pyright: ignore[reportUnknownVariableType]
		except:
			logger.exception("Failed to parse the config file")
			sys.exit(1)
		else:
			copyDict(loadedConfig, config)
		if "hideTotalTime" in config["display"]:
			config["display"]["duration"] = not config["display"]["hideTotalTime"] # pyright: ignore[reportGeneralTypeIssues]
			del config["display"]["hideTotalTime"] # pyright: ignore[reportGeneralTypeIssues]
		if "useRemainingTime" in config["display"]:
			del config["display"]["useRemainingTime"] # pyright: ignore[reportGeneralTypeIssues]
		if "remainingTime" in config["display"]:
			del config["display"]["remainingTime"] # pyright: ignore[reportGeneralTypeIssues]
		if config["display"]["progressMode"] not in ["off", "elapsed", "remaining", "bar"]:
			config["display"]["progressMode"] = "bar"
	save()

class YamlSafeDumper(yaml.SafeDumper):
    def increase_indent(self, flow: bool = False, indentless: bool = False) -> None:
        return super().increase_indent(flow, False)

def save() -> None:
	try:
		with open(configFilePath, "w", encoding = "UTF-8") as configFile:
			if configFileType == "yaml":
				yaml.dump(config, configFile, sort_keys = False, Dumper = YamlSafeDumper, allow_unicode = True)
			else:
				json.dump(config, configFile, indent = "\t")
				configFile.write("\n")
	except:
		logger.exception("Failed to write to the config file")

def copyDict(source: Any, target: Any) -> None:
	for key, value in source.items():
		if isinstance(value, dict):
			copyDict(value, target.setdefault(key, {}))
		else:
			target[key] = value
