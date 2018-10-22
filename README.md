# Discord Rich Presence for Plex

A Python script that displays your [Plex](https://www.plex.tv) status on [Discord](https://discordapp.com) using [Rich Presence](https://discordapp.com/developers/docs/rich-presence/how-to).

## Requirements

* [Python 3.6.4+](https://www.python.org/downloads)
* [plexapi](https://github.com/pkkid/python-plexapi)
* [websocket-client](https://github.com/websocket-client/websocket-client)
	* Use version 0.48.0 (`pip install websocket-client==0.48.0`) as an issue with newer versions of websocket-client breaks the plexapi module's alert listener
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

* `serverName` - Name of the Plex Media Server to connect to
* `username` - Username of the account the above server is signed in as
* `password` - Password associated with the above account
* `token` - A [X-Plex-Token](https://support.plex.tv/articles/204059436-finding-an-authentication-token-x-plex-token) associated with the above account, ignores `password` if set
* `listenForUser` - Your username, defaults to `username` if not set

### Other Variables

* Line 16: `extraLogging` - The program outputs more information if this is set to `True`

## License

This project is licensed under the MIT License. See the [LICENSE](https://github.com/phin05/discord-rich-presence-plex/blob/master/LICENSE) file for details.
