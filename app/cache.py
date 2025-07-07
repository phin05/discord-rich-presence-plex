import sys
from app import logger, constants
from typing import Any
import json
import os

cache: dict[str, Any] = {}

def load() -> None:
	if not os.path.isfile(constants.cacheFilePath):
		return
	try:
		with open(constants.cacheFilePath, "r", encoding = "UTF-8") as cacheFile:
			cache.update(json.load(cacheFile))
	except:
		logger.exception("Failed to parse the cache file")
		sys.exit(1)

def get(key: str) -> Any:
	return cache.get(key)

def set(key: str, value: Any) -> None:
	cache[key] = value
	try:
		with open(constants.cacheFilePath, "w", encoding = "UTF-8") as cacheFile:
			json.dump(cache, cacheFile, separators = (",", ":"))
	except:
		logger.exception("Failed to write to the cache file")
