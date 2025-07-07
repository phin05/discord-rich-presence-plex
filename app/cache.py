from app import logger, constants
from typing import  Optional, TypedDict
import json
import os
import sys

class CacheEntry(TypedDict):
	value: str
	expiry: int

cache: dict[str, CacheEntry] = {}

def load() -> None:
	if not os.path.isfile(constants.cacheFilePath):
		return
	try:
		with open(constants.cacheFilePath, "r", encoding = "UTF-8") as cacheFile:
			cache.update(json.load(cacheFile))
	except:
		logger.exception("Failed to parse the cache file")
		sys.exit(1)

def get(key: str) -> Optional[CacheEntry]:
	return cache.get(key)

def set(key: str, value: str, expiry: int) -> None:
	cache[key] = { "value": value, "expiry": expiry }
	try:
		with open(constants.cacheFilePath, "w", encoding = "UTF-8") as cacheFile:
			json.dump(cache, cacheFile, indent = "\t")
	except:
		logger.exception("Failed to write to the cache file")
