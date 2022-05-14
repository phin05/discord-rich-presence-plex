from store.constants import cacheFilePath
from typing import Any
from utils.logging import logger
import json
import os
import time

cache: dict[str, Any] = {}

def loadCache() -> None:
	if os.path.isfile(cacheFilePath):
		try:
			with open(cacheFilePath, "r", encoding = "UTF-8") as cacheFile:
				cache.update(json.load(cacheFile))
		except:
			os.rename(cacheFilePath, cacheFilePath.replace(".json", f"-{time.time():.0f}.json"))
			logger.exception("Failed to parse the application's cache file. A new one will be created.")

def getKey(key: str) -> Any:
	return cache.get(key)

def setKey(key: str, value: Any) -> None:
	cache[key] = value
	try:
		with open(cacheFilePath, "w", encoding = "UTF-8") as cacheFile:
			json.dump(cache, cacheFile, separators = (",", ":"))
	except:
		logger.exception("Failed to write to the application's cache file.")
