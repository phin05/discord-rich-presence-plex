# Discord Rich Presence for Plex

A Python script that displays your [Plex](https://www.plex.tv) status on [Discord](https://discord.com) using [Rich Presence](https://discord.com/developers/docs/rich-presence/how-to).

Current Version: 2.2.0

## Getting Started

1. Install [Python 3.10](https://www.python.org/downloads/) - Make sure to tick "Add Python 3.10 to PATH" during the installation
2. Download [this repository's contents](https://github.com/phin05/discord-rich-presence-plex/archive/refs/heads/master.zip)
3. Extract the folder contained in the above ZIP file
4. Navigate a command-line interface (cmd.exe, PowerShell, bash, etc.) into the above-extracted directory
5. Install the required Python modules by running `python -m pip install -U -r requirements.txt`
6. Start the script by running `python main.py`

When the script runs for the first time, a `config.json` file will be created in the working directory and you will be prompted to complete the authentication flow to allow the script to retrieve an access token for your Plex account.

The script must be running on the same machine as your Discord client.

## Configuration - `config.json`

### Reference

* `logging`
  * `debug` (boolean, default: `true`) - Outputs additional debug-helpful information to the console if enabled.
* `display`
  * `useRemainingTime` (boolean, default: `false`) - Displays your media's remaining time instead of elapsed time in your Rich Presence if enabled.
  * `posters`
    * `enabled` (boolean, default: `false`) - Displays media posters in Rich Presence if enabled. Requires `imgurClientID`.
    * `imgurClientID` (string, default: `""`) - [Instructions](#obtaining-an-imgur-client-id)
* `users` (list)
  * `token` (string) - An access token associated with your Plex account. ([X-Plex-Token](https://support.plex.tv/articles/204059436-finding-an-authentication-token-x-plex-token), [Authenticating with Plex](https://forums.plex.tv/t/authenticating-with-plex/609370))
  * `servers` (list)
    * `name` (string) - Name of the Plex Media Server you wish to connect to.
    * `listenForUser` (string, optional) - The script will respond to alerts originating only from this username. Defaults to the parent user's username if not set.
    * `blacklistedLibraries` (list, optional) - Alerts originating from libraries in this list are ignored.
    * `whitelistedLibraries` (list, optional) - If set, alerts originating from libraries that are not in this list are ignored.

### Obtaining an Imgur client ID

1. Go to Imgur's [application registration page](https://api.imgur.com/oauth2/addclient)
2. Enter any name for the application and pick OAuth2 without a callback URL as the authorisation type
3. Submit the form to obtain your application's client ID

### Example

```json
{
  "logging": {
    "debug": true
  },
  "display": {
    "useRemainingTime": false,
    "posters": {
      "enabled": true,
      "imgurClientID": "9e9sf637S8bRp4z"
    }
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
          "listenForUser": "xyz",
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
* [Python-PlexAPI](https://github.com/pkkid/python-plexapi)
* [Requests](https://github.com/psf/requests)
* [websocket-client](https://github.com/websocket-client/websocket-client)
