# Discord Rich Presence for Plex

A Python script that displays your [Plex](https://www.plex.tv) status on [Discord](https://discord.com) using [Rich Presence](https://discord.com/developers/docs/rich-presence/how-to).

Current Version: 2.0.3

## Getting Started

1. Install [Python 3.10](https://www.python.org/downloads/)
2. Download [this repository's contents](https://github.com/phin05/discord-rich-presence-plex/archive/refs/heads/master.zip)
3. Extract the folder contained in the above ZIP file
4. Navigate a command-line interface (cmd.exe, PowerShell, bash, etc.) into the above-extracted directory
5. Install the required Python modules by running `python -m pip install -r requirements.txt`
6. Start the script by running `python main.py`

When the script runs for the first time, a `config.json` file will be created in the working directory and you will be prompted to complete the authentication flow to allow the script to retrieve an access token for your Plex account.

The script must be running on the same machine as your Discord client.

## Configuration - `config.json`

### Reference

* `logging`
  * `debug` (default: `true`) - Outputs additional debug-helpful information to the console if enabled.
* `display`
  * `useRemainingTime` (default: `false`) - Displays your media's remaining time instead of elapsed time in your Rich Presence  if enabled.
* `users` (list)
  * `token` - An access token associated with your Plex account. ([X-Plex-Token](https://support.plex.tv/articles/204059436-finding-an-authentication-token-x-plex-token), [Authenticating with Plex](https://forums.plex.tv/t/authenticating-with-plex/609370))
  * `servers` (list)
    * `name` - Name of the Plex Media Server you wish to connect to.
    * `blacklistedLibraries` (optional list) - Alerts originating from libraries in this list are ignored.
    * `whitelistedLibraries` (optional list) - If set, alerts originating from libraries that are not in this list are ignored.

### Example

```json
{
  "logging": {
    "debug": true
  },
  "display": {
    "useRemainingTime": false
  },
  "users": [
    {
      "token": "HPbrz2NhfLRjU888Rrdt",
      "servers": [
        {
          "name": "Bob's Home Media Server"
        },
        {
          "name": "A Friend's Server",
          "whitelistedLibraries": ["Movies"]
        }
      ]
    }
  ]
}
```

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Credits

* [Discord](https://discord.com)
* [Plex](https://www.plex.tv)
* [plexapi](https://github.com/pkkid/python-plexapi)
* [requests](https://github.com/psf/requests)
* [websocket-client](https://github.com/websocket-client/websocket-client)
