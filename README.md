# Discord Rich Presence for Plex

A Python script that displays your [Plex](https://www.plex.tv) status on [Discord](https://discordapp.com) using [Rich Presence](https://discordapp.com/developers/docs/rich-presence/how-to).

## Requirements

* [Python 3.6.7](https://www.python.org/downloads/release/python-367/)
* [plexapi](https://github.com/pkkid/python-plexapi)
* Use [websocket-client](https://github.com/websocket-client/websocket-client) version 0.48.0 (`pip install websocket-client==0.48.0`) as an issue with newer versions breaks the plexapi module's alert listener.
* The script must be running on the same machine as your Discord client.

## Configuration

Add your configuration(s) into the `plexConfigs` list on line 30.

#### Example

```python
plexConfigs = [
	plexConfig(serverName = "ABC", username = "xyz", password = "0tYD4UIC4Tb8X0nt"),
	plexConfig(serverName = "DEF", username = "pqr@pqr.pqr", token = "70iU3GZrI54S76Tn", listenForUser = "xyz"),
	plexConfig(serverName = "GHI", username = "xyz", password = "0tYD4UIC4Tb8X0nt", blacklistedLibraries = ["TV Shows", "Music"])
]
```

#### Parameters

* `serverName` - Name of the Plex Media Server to connect to.
* `username` - Your account's username or e-mail.
* `password` (not required if `token` is set) - The password associated with the above account.
* `token` (not required if `password` is set) - A [X-Plex-Token](https://support.plex.tv/articles/204059436-finding-an-authentication-token-x-plex-token) associated with the above account. In some cases, `myPlexAccessToken` from Plex Web App's HTML5 Local Storage must be used. To retrieve this token in Google Chrome, open Plex Web App, press F12, go to "Application", expand "Local Storage" and select the relevant entry. Ignores `password` if set.
* `listenForUser` (optional) - The script will respond to alerts originating only from this username. Defaults to `username` if not set.
* `blacklistedLibraries` (list, optional) - Alerts originating from blacklisted libraries are ignored.
* `whitelistedLibraries` (list, optional) - If set, alerts originating from libraries that are not in the whitelist are ignored.

### Other Variables

* Line 16: `extraLogging` - The script outputs more information if this is set to `True`.
* Line 17: `timeRemaining` - Set this to `True` to display time remaining instead of time elapsed while media is playing.

## Auto-start function for Windows

This auto-start function edits the Windows Registry from within the Python script. Make sure to back up your machine prior to enabling this script, editing the Windows Registry can break OS-level functions and cause you to lose your data. Proceed at your own risk. 

In order to use the autostart function for Windows you'll need to create a shortcut to your **plex_discord_rpc.py** file in `C:\Users\`**your_username**`\AppData\Roaming\Microsoft\Windows\Start Menu\Programs\Startup\`. From there you'll need to edit the following parameter within the .py file itself. 

### Parameter for auto-start

* `__file__` = the raw file path to your **plex_discord_rpc.py** file. 

## License

This project is licensed under the MIT License. See the [LICENSE](https://github.com/Phineas05/discord-rich-presence-plex/blob/master/LICENSE) file for details.
