import os
import sys

name = "Discord Rich Presence for Plex"
version = "2.2.2"

plexClientID = "discord-rich-presence-plex"
discordClientID = "413407336082833418"
configFilePath = "config.json"
cacheFilePath = "cache.json"

isUnix = sys.platform in ["linux", "darwin"]
processID = os.getpid()
