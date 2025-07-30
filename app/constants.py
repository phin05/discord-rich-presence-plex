import os
import sys

name = "Discord Rich Presence for Plex"
version = "2.15.0"

plexClientID = "discord-rich-presence-plex"
discordClientID = "413407336082833418"

dataDirectoryPath = "data"
configFilePathBase = os.path.join(dataDirectoryPath, "config")
cacheFilePath = os.path.join(dataDirectoryPath, "cache.json")
logFilePath = os.path.join(dataDirectoryPath, "console.log")

isUnix = sys.platform in ["linux", "darwin"]
processID = os.getpid()
isInteractive = sys.stdin and sys.stdin.isatty()
plexServerNameInput = os.environ.get("DRPP_PLEX_SERVER_NAME_INPUT")
noPipInstall = os.environ.get("DRPP_NO_PIP_INSTALL", "") == "true"
isInContainer = os.environ.get("DRPP_IS_IN_CONTAINER", "") == "true"
runtimeDirectory = "/run/app" if isInContainer else os.environ.get("XDG_RUNTIME_DIR", os.environ.get("TMPDIR", os.environ.get("TMP", os.environ.get("TEMP", "/tmp"))))
ipcPipeBase = runtimeDirectory if isUnix else r"\\?\pipe"
uid = int(os.environ.get("DRPP_UID", "-1"))
gid = int(os.environ.get("DRPP_GID", "-1"))
containerCwd = "/app"
noRuntimeDirChown = os.environ.get("DRPP_NO_RUNTIME_DIR_CHOWN", "") == "true"
