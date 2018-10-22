# Discord Rich Presence for Plex

A Python script that displays your [Plex](https://www.plex.tv) status on [Discord](https://discordapp.com) using [Rich Presence](https://discordapp.com/developers/docs/rich-presence/how-to).

## Requirements

* [Python 3.6.7](https://www.python.org/downloads/release/python-367/)
* [plexapi](https://github.com/pkkid/python-plexapi)
* Use [websocket-client](https://github.com/websocket-client/websocket-client) version 0.48.0 (`pip install websocket-client==0.48.0`) as an issue with newer versions of websocket-client breaks the plexapi module's alert listener.
* The script must be running on the same machine as the Discord client.

## Configuration

Add your configuration(s) into the `plexConfigs` list on line 26.

#### Example

```python
plexConfigs = [
	plexConfig(serverName = "ABC", username = "xyz", password = "0tYD4UIC4Tb8X0nt"),
	plexConfig(serverName = "DEF", username = "pqr@pqr.pqr", token = "70iU3GZrI54S76Tn", listenForUser = "xyz")
]
```

#### Parameters

* `serverName` - Name of the Plex Media Server to connect to.
* `username` - Your account's username or e-mail.
* `password` - The password associated with the above account.
* `token` (optional) - A [X-Plex-Token](https://support.plex.tv/articles/204059436-finding-an-authentication-token-x-plex-token) associated with the above account. In some cases, `myPlexAccessToken` from Plex Web App's HTML5 Local Storage must be used. To retrieve this token in Google Chrome, open Plex Web App, press F12, go to "Application", expand "Local Storage" and select the relevant entry. Ignores `password` if set.
* `listenForUser` - The script will respond to alerts originating only from this username. Defaults to `username` if not set.

### Other Variables

* Line 16: `extraLogging` - The script outputs more information if this is set to `True`

## License

This project is licensed under the MIT License. See the [LICENSE](https://github.com/Phineas05/discord-rich-presence-plex/blob/master/LICENSE) file for details.
