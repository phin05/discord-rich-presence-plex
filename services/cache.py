from typing import Any
from store.constants import cacheFilePath
from utils.logging import logger
import json
import os

cache: dict[str, Any] = {}

def loadCache() -> None:
	global cache
	if os.path.isfile(cacheFilePath):
		try:
			with open(cacheFilePath, "r", encoding = "UTF-8") as cacheFile:
				cache = json.load(cacheFile)
		except:
			logger.exception("Failed to parse the application's cache file.")

def getKey(key: str) -> Any:
	return cache.get(key)

def setKey(key: str, value: Any) -> None:
	cache[key] = value
	try:
		with open(cacheFilePath, "w", encoding = "UTF-8") as cacheFile:
			json.dump(cache, cacheFile, separators = (",", ":"))
	except:
		logger.exception("Failed to write to the application's cache file.")
