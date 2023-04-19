import os
import sys

name = "Discord Rich Presence for Plexamp"
version = "2.3.2"

plexClientID = "discord-rich-presence-plexamp"
discordClientID = "1098083071116447815"
configFilePath = "config.json"
cacheFilePath = "cache.json"
logFilePath = "console.log"

isUnix = sys.platform in ["linux", "darwin"]
processID = os.getpid()
