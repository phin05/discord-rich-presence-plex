from enum import IntEnum
from typing import TypedDict

class ActivityType(IntEnum):
	LISTENING = 2
	WATCHING = 3

class ActivityAssets(TypedDict, total = False):
	large_text: str
	large_image: str
	small_text: str
	small_image: str

class ActivityTimestamps(TypedDict, total = False):
	start: int
	end: int

class ActivityButton(TypedDict):
	label: str
	url: str

class Activity(TypedDict, total = False):
	type: ActivityType
	details: str
	state: str
	assets: ActivityAssets
	timestamps: ActivityTimestamps
	buttons: list[ActivityButton]
