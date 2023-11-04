from config.constants import configFilePathRoot
from utils.dict import copyDict
from utils.logging import logger
import json
import models.config
import os
import time
import yaml

config: models.config.Config = {
	"logging": {
		"debug": True,
		"writeToFile": False,
	},
	"display": {
		"hideTotalTime": False,
		"useRemainingTime": False,
		"posters": {
			"enabled": False,
			"imgurClientID": "",
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
	global configFileType, configFileExtension, configFilePath
	for fileExtension, fileType in supportedConfigFileExtensions.items():
		filePath = f"{configFilePathRoot}.{fileExtension}"
		if os.path.isfile(filePath):
			configFileExtension = fileExtension
			configFileType = fileType
			break
	else:
		configFileExtension, configFileType = list(supportedConfigFileExtensions.items())[0]
		filePath = f"{configFilePathRoot}.{configFileExtension}"
	configFilePath = filePath
	if os.path.isfile(configFilePath):
		try:
			with open(configFilePath, "r", encoding = "UTF-8") as configFile:
				if configFileType == "yaml":
					loadedConfig = yaml.safe_load(configFile) or {}
				else:
					loadedConfig = json.load(configFile) or {}
		except:
			os.rename(configFilePath, f"{configFilePathRoot}-{time.time():.0f}.{configFileExtension}")
			logger.exception("Failed to parse the application's config file. A new one will be created.")
		else:
			copyDict(loadedConfig, config)
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
		logger.exception("Failed to write to the application's config file.")
