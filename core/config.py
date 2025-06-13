from config.constants import configFilePathBase
from utils.dict import copyDict
from utils.logging import logger
import json
import models.config
import os
import sys
import yaml

config: models.config.Config = {
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
			"enabled": False,
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

def loadConfig() -> None:
	global configFileExtension, configFileType, configFilePath
	doesFileExist = False
	for i, (fileExtension, fileType) in enumerate(supportedConfigFileExtensions.items()):
		doesFileExist = os.path.isfile(f"{configFilePathBase}.{fileExtension}")
		isFirstItem = i == 0
		if doesFileExist or isFirstItem:
			configFileExtension = fileExtension
			configFileType = fileType
			configFilePath = f"{configFilePathBase}.{configFileExtension}"
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
	saveConfig()

class YamlSafeDumper(yaml.SafeDumper):
    def increase_indent(self, flow: bool = False, indentless: bool = False) -> None:
        return super().increase_indent(flow, False)

def saveConfig() -> None:
	try:
		with open(configFilePath, "w", encoding = "UTF-8") as configFile:
			if configFileType == "yaml":
				yaml.dump(config, configFile, sort_keys = False, Dumper = YamlSafeDumper, allow_unicode = True)
			else:
				json.dump(config, configFile, indent = "\t")
				configFile.write("\n")
	except:
		logger.exception("Failed to write to the config file")
