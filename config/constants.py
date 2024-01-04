import os
import sys

name = "Discord Rich Presence for Plex"
version = "2.4.3"

plexClientID = "discord-rich-presence-plex"
discordClientID = "413407336082833418"

dataDirectoryPath = "data"
configFilePathRoot = os.path.join(dataDirectoryPath, "config")
cacheFilePath = os.path.join(dataDirectoryPath, "cache.json")
logFilePath = os.path.join(dataDirectoryPath, "console.log")

isUnix = sys.platform in ["linux", "darwin"]
processID = os.getpid()
isInteractive = sys.stdin and sys.stdin.isatty()
containerDemotionUidGid = os.environ.get("DRPP_CONTAINER_DEMOTION_UID_GID", "")
