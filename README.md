# Discord Rich Presence for Plex

![Showcase](https://user-images.githubusercontent.com/59180111/168054648-af0590fd-9bd7-42d0-91b2-d7974643debd.png)

Discord Rich Presence for Plex is a Python script which displays your [Plex](https://www.plex.tv/) status on [Discord](https://discord.com/) using [Rich Presence](https://discord.com/developers/docs/rich-presence/how-to).

[![Latest Release](https://img.shields.io/github/v/release/phin05/discord-rich-presence-plex?label=Latest%20Release)](https://github.com/phin05/discord-rich-presence-plex/releases/latest)
[![Build Status](https://img.shields.io/github/actions/workflow/status/phin05/discord-rich-presence-plex/release.yml?label=Build&logo=github)](https://github.com/phin05/discord-rich-presence-plex/actions/workflows/release.yml)

## Usage

When the script runs for the first time, a `data` directory will be created in the current working directory along with a `config.json` file inside of it. You will be prompted to complete the authentication flow to allow the script to retrieve an access token for your Plex account.

The script must be running on the same machine as your Discord client.

### Instructions

1. Install [Python 3.10](https://www.python.org/downloads/) - Make sure to tick "Add Python 3.10 to PATH" during the installation.
2. Download the [latest release](https://github.com/phin05/discord-rich-presence-plex/releases/latest) of this script.
3. Extract the directory contained in the above ZIP file.
4. Navigate a command-line interface (cmd.exe, PowerShell, bash, etc.) into the above-extracted directory.
5. Install the required Python modules by running `python -m pip install -U -r requirements.txt`.
6. Start the script by running `python main.py`.

Alternatively, you can [run with Docker](#run-with-docker) if you're using a Linux-based operating system.

## Configuration - `config.json`

* `logging`
  * `debug` (boolean, default: `true`) - Outputs additional debug-helpful information to the console if enabled.
  * `writeToFile` (boolean, default: `false`) - Writes console output to a `console.log` file in the `data` directory if enabled.
* `display` - Display settings for Rich Presence
  * `hideTotalTime` (boolean, default: `false`) - Hides the total duration of the media if enabled.
  * `useRemainingTime` (boolean, default: `false`) - Displays the media's remaining time instead of elapsed time if enabled.
  * `posters`
    * `enabled` (boolean, default: `false`) - Displays media posters if enabled. Requires `imgurClientID`.
    * `imgurClientID` (string, default: `""`) - [Obtention Instructions](#obtaining-an-imgur-client-id)
  * `buttons` (list) - [Information](#buttons)
    * `label` (string) - The label to be displayed on the button.
    * `url` (string) - A web address or a [dynamic URL placeholder](#dynamic-button-urls).
* `users` (list)
  * `token` (string) - An access token associated with your Plex account. ([X-Plex-Token](https://support.plex.tv/articles/204059436-finding-an-authentication-token-x-plex-token/), [Authenticating with Plex](https://forums.plex.tv/t/authenticating-with-plex/609370))
  * `servers` (list)
    * `name` (string) - Name of the Plex Media Server you wish to connect to.
    * `listenForUser` (string, optional) - The script will respond to alerts originating only from this username. Defaults to the parent user's username if not set.
    * `blacklistedLibraries` (list, optional) - Alerts originating from libraries in this list are ignored.
    * `whitelistedLibraries` (list, optional) - If set, alerts originating from libraries that are not in this list are ignored.

### Obtaining an Imgur client ID

1. Go to Imgur's [application registration page](https://api.imgur.com/oauth2/addclient).
2. Enter any name for the application and pick OAuth2 without a callback URL as the authorisation type.
3. Submit the form to obtain your application's client ID.

### Buttons

Discord can display up to 2 buttons in your Rich Presence.

Due to a strange Discord bug, these buttons are unresponsive or exhibit strange behaviour towards your own clicks, but other users are able to click on them to open their corresponding URLs.

#### Dynamic Button URLs

During runtime, the following dynamic URL placeholders will get replaced with real URLs based on the media being played:
* `dynamic:imdb`
* `dynamic:tmdb`

### Example

```json
{
  "logging": {
    "debug": true,
    "writeToFile": false
  },
  "display": {
    "hideTotalTime": false,
    "useRemainingTime": false,
    "posters": {
      "enabled": true,
      "imgurClientID": "9e9sf637S8bRp4z"
    },
    "buttons": [
      {
        "label": "IMDb Link",
        "url": "dynamic:imdb"
      },
      {
        "label": "My YouTube Channel",
        "url": "https://www.youtube.com/channel/me"
      }
    ]
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

## Configuration - Discord

The "Display current activity as a status message" setting must be enabled in Discord Settings → Activity Settings → Activity Privacy.

![Discord Settings](https://user-images.githubusercontent.com/59180111/186830889-35af3895-ece0-4a7d-9efb-f68640116884.png)

## Run with Docker

### Image

[ghcr.io/phin05/discord-rich-presence-plex](https://ghcr.io/phin05/discord-rich-presence-plex)

### Volumes

Mount a directory for persistent data (config file, cache file and log file) at `/app/data`.

The directory where Discord stores its inter-process communication Unix socket file needs to be mounted into the container at `/run/app`. The path for this would be the first non-null value from the values of the following environment variables: ([source](https://github.com/discord/discord-rpc/blob/963aa9f3e5ce81a4682c6ca3d136cddda614db33/src/connection_unix.cpp#L29C33-L29C33))

* XDG_RUNTIME_DIR
* TMPDIR
* TMP
* TEMP
* Fallback path: /tmp

### Example

```
docker run \
  -v ./data:/app/data \
  -v /run/user/1000:/run/app:ro \
  --detach \
  --restart unless-stopped \
  --name discord-rich-presence-plex \
  ghcr.io/phin05/discord-rich-presence-plex:latest
```
