from typing import TypedDict

class StateNotification(TypedDict):
	state: str
	sessionKey: int
	ratingKey: int
	viewOffset: int

class Alert(TypedDict):
	type: str
	PlaySessionStateNotification: list[StateNotification]
