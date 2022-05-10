from datetime import datetime
from models.config import Config
from utils.logs import logger
import json
import os

class ConfigService:

	config: Config

	def __init__(self, configFilePath: str) -> None:
		self.configFilePath = configFilePath
		if os.path.isfile(self.configFilePath):
			try:
				with open(self.configFilePath, "r", encoding = "UTF-8") as configFile:
					self.config = json.load(configFile)
			except:
				os.rename(configFilePath, configFilePath.replace(".json", f"-{datetime.now().timestamp():.0f}.json"))
				logger.exception("Failed to parse the application's config file. A new one will be created.")
				self.resetConfig()
		else:
			self.resetConfig()

	def resetConfig(self) -> None:
		self.config = {
			"logging": {
				"debug": True
			},
			"display": {
				"useRemainingTime": False
			},
			"users": []
		}
		self.saveConfig()

	def saveConfig(self) -> None:
		try:
			with open(self.configFilePath, "w", encoding = "UTF-8") as configFile:
				json.dump(self.config, configFile, indent = "\t")
		except:
			logger.exception("Failed to write to the application's config file.\n%s")
