# Discord Rich Presence for Plex

A Python script that displays your [Plex](https://www.plex.tv) status on [Discord](https://discordapp.com) using [Rich Presence](https://discordapp.com/developers/docs/rich-presence/how-to).

## Requirements

* [Python 3.6.4+](https://www.python.org/downloads)
* [plexapi](https://github.com/pkkid/python-plexapi)
* The script must be running on the same machine as the Discord client.

## Variables

You will have to change the following variables in `discordRichPresencePlex.py`:

* Line 82: `plexServerName` - Name of the Plex Media Server to connect to
* Line 83: `plexUsername` - Username of the account the server is signed in as
* Line 84: `plexPasswordOrToken` - Password or a X-Plex-Token associated with the above account
* Line 85: `usingToken` - Set this to `True` if the above is a X-Plex-Token
* Line 86: `listenForUser` - Your username, leave blank if it's the same as `plexUsername`
