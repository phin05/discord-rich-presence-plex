from .logging import logger
from config.constants import cacheFilePath
from typing import Any
import json
import os
import time

cache: dict[str, Any] = {}

def loadCache() -> None:
	if not os.path.isfile(cacheFilePath):
		return
	try:
		with open(cacheFilePath, "r", encoding = "UTF-8") as cacheFile:
			cache.update(json.load(cacheFile))
	except:
		root, ext = os.path.splitext(cacheFilePath)
		os.rename(cacheFilePath, f"{root}-{time.time():.0f}.{ext}")
		logger.exception("Failed to parse the cache file. A new one will be created.")

def getCacheKey(key: str) -> Any:
	return cache.get(key)

def setCacheKey(key: str, value: Any) -> None:
	cache[key] = value
	try:
		with open(cacheFilePath, "w", encoding = "UTF-8") as cacheFile:
			json.dump(cache, cacheFile, separators = (",", ":"))
	except:
		logger.exception("Failed to write to the cache file")
