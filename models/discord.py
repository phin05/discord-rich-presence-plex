from typing import TypedDict

class ActivityAssets(TypedDict):
	large_text: str
	large_image: str
	small_text: str
	small_image: str

class ActivityTimestamps(TypedDict, total = False):
	start: int
	end: int

class Activity(TypedDict, total = False):
	details: str
	state: str
	assets: ActivityAssets
	timestamps: ActivityTimestamps
