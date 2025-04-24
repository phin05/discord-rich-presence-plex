from typing import TypedDict

class Logging(TypedDict):
	debug: bool
	writeToFile: bool

class Posters(TypedDict):
	enabled: bool
	imgurClientID: str
	maxSize: int
	fit: bool

class Button(TypedDict):
	label: str
	url: str
	mediaTypes: list[str]

class Display(TypedDict):
	duration: bool
	genres: bool
	album: bool
	albumImage: bool
	artist: bool
	artistImage: bool
	year: bool
	statusIcon: bool
	progressMode: str
	paused: bool
	posters: Posters
	buttons: list[Button]

class Server(TypedDict, total = False):
	name: str
	listenForUser: str
	blacklistedLibraries: list[str]
	whitelistedLibraries: list[str]
	ipcPipeNumber: int

class User(TypedDict):
	token: str
	servers: list[Server]

class Config(TypedDict):
	logging: Logging
	display: Display
	users: list[User]
