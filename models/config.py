from typing import TypedDict

class Logging(TypedDict):
	debug: bool
	writeToFile: bool

class Posters(TypedDict):
	enabled: bool
	imgurClientID: str

class Display(TypedDict):
	useRemainingTime: bool
	posters: Posters

class Server(TypedDict, total = False):
	name: str
	listenForUser: str
	blacklistedLibraries: list[str]
	whitelistedLibraries: list[str]

class User(TypedDict):
	token: str
	servers: list[Server]

class Config(TypedDict):
	logging: Logging
	display: Display
	users: list[User]
