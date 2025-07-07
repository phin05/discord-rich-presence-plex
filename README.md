# Discord Rich Presence for Plex

Discord Rich Presence for Plex is a Python script which displays your [Plex](https://www.plex.tv/) status on [Discord](https://discord.com/) using [Rich Presence](https://discord.com/developers/docs/rich-presence/overview).

[![Latest Release](https://img.shields.io/github/v/release/phin05/discord-rich-presence-plex?label=Latest%20Release)](https://github.com/phin05/discord-rich-presence-plex/releases/latest)
[![Build Status](https://img.shields.io/github/actions/workflow/status/phin05/discord-rich-presence-plex/release.yml?label=Build&logo=github)](https://github.com/phin05/discord-rich-presence-plex/actions/workflows/release.yml)

![Showcase](assets/readme/showcase.png)

## Installation

If you're using a Linux-based operating system, you can [run this script with Docker](#run-with-docker). Otherwise, follow these instructions:

1. Install [Python](https://www.python.org/downloads/) (version 3.10 or newer) - Make sure to tick "Add Python to PATH" during the installation.
2. Download the [latest release](https://github.com/phin05/discord-rich-presence-plex/releases/latest) of this script.
3. Extract the directory contained in the above ZIP file.
4. Navigate a command-line interface (cmd, PowerShell, bash, etc.) into the above-extracted directory.
5. Start the script by running `python main.py`.

When the script runs for the first time, a directory named `data` will be created in the current working directory along with a `config.yaml` file inside of it. You will be prompted to complete authentication to allow the script to retrieve an access token for your Plex account.

The script must be running on the same machine as your Discord client.

## Configuration

The config file is stored in a directory named `data`.

### Supported Formats

- YAML - `config.yaml` / `config.yml`
- JSON - `config.json`

### Reference

- `logging`
  - `debug` (boolean, default: `true`) - Outputs additional debug-helpful information to the console.
  - `writeToFile` (boolean, default: `false`) - Writes console output to a `console.log` file in the `data` directory.
- `display` - Display settings for Rich Presence
  - `duration` (boolean, default: `false`) - Displays the total duration. Applicable to movies and TV shows only.
  - `genres` (boolean, default: `true`) - Displays the genre. Applicable to movies only.
  - `album` (boolean, default: `true`) - Displays the album name. Applicable to music only.
  - `albumImage` (boolean, default: `true`) - Displays the album image. Applicable to music only.
  - `artist` (boolean, default: `true`) - Displays the artist name. Applicable to music only.
  - `artistImage` (boolean, default: `true`) - Displays the artist image. Applicable to music only.
  - `year` (boolean, default: `true`) - Displays the release year.
  - `statusIcon` (boolean, default: `false`) - Displays a status icon (playing, paused, buffering) at the bottom-right corner of the poster. Applicable to movies and TV shows only.
  - `progressMode` (string, default: `bar`) - Progress/timestamp display mode. Valid modes are `off`, `elapsed` (displays elapsed time), `remaining` (displays remaining time) and `bar` (displays a progress bar). The `off` and `remaining` modes are currently broken due to a Discord bug/limitation.
  - `paused` (boolean, default: `false`) - Displays Rich Presence even while media is paused. Progress/timestamp display while paused is currently broken due to a Discord bug/limitation.
  - `posters`
    - `enabled` (boolean, default: `true`) - Displays media posters (including album art and artist images). [Imgur](https://imgur.com/) is used if `imgurClientID` is set, otherwise [Litterbox](https://litterbox.catbox.moe/) is used.
    - `imgurClientID` (string, default: `""`) - [Obtention Instructions](#obtaining-an-imgur-client-id)
    - `maxSize` (int, default: `256`) - Maximum width and maximum height to use while downscaling posters before uploading them.
    - `fit` (boolean, default: `true`) - Fits posters inside a square while maintaining the original aspect ratio. Otherwise, Discord crops posters into a square.
  - `buttons` (list) - [Information](#buttons)
    - `label` (string) - The label to be displayed on the button.
    - `url` (string) - A web address or a [dynamic URL placeholder](#dynamic-button-urls).
    - `mediaTypes` (list, optional) - If set, the button is displayed only for the specified media types. Valid media types are `movie`, `episode`, `live_episode`, `track` and `clip`.
- `users` (list)
  - `token` (string) - An access token associated with your Plex account. ([X-Plex-Token](https://support.plex.tv/articles/204059436-finding-an-authentication-token-x-plex-token/), [Authenticating with Plex](https://forums.plex.tv/t/authenticating-with-plex/609370))
  - `servers` (list)
    - `name` (string) - Name of the Plex Media Server to connect to.
    - `listenForUser` (string, optional) - The script reacts to alerts originating only from this username. Defaults to the parent user's username if not set.
    - `blacklistedLibraries` (list, optional) - Alerts originating from libraries in this list are ignored.
    - `whitelistedLibraries` (list, optional) - If set, alerts originating from libraries that are not in this list are ignored.
    - `ipcPipeNumber` (int, optional) - A number in the range of `0-9` to specify the Discord IPC pipe to connect to. Defaults to `-1`, which specifies that the first existing pipe in the range should be used. When a Discord client is launched, it binds to the first unbound pipe number, which is typically `0`.

### Obtaining an Imgur client ID

1. Go to Imgur's [application registration page](https://api.imgur.com/oauth2/addclient).
2. Enter any name for the application and pick "OAuth 2 authorization without a callback URL" as the authorisation type.
3. Submit the form to obtain your application's client ID.

### Buttons

Discord can display up to 2 buttons in your Rich Presence. Buttons are visible to only other users and not yourself (Discord bug/limitation).

#### Dynamic Button Labels

Instances of `{title}` in button labels will be replaced with the top-level title of the media being played.

#### Dynamic Button URLs

During runtime, the following dynamic URL placeholders will get replaced with real URLs based on the media being played:

- `dynamic:imdb`
- `dynamic:tmdb`
- `dynamic:thetvdb`
- `dynamic:trakt`
- `dynamic:letterboxd`
- `dynamic:musicbrainz`

### Example (YAML)

```yaml
logging:
  debug: true
  writeToFile: false
display:
  duration: false
  genres: true
  album: true
  albumImage: true
  artist: true
  artistImage: true
  year: true
  statusIcon: false
  progressMode: bar
  paused: false
  posters:
    enabled: true
    imgurClientID: 9e9sf637S8bRp4z
    maxSize: 256
    fit: true
  buttons:
    - label: "IMDb: {title}"
      url: dynamic:imdb
    - label: Music Stats
      url: https://github.com
      mediaTypes:
        - track
users:
  - token: HPbrz2NhfLRjU888Rrdt
    servers:
      - name: Bob's Home Media Server
      - name: A Friend's Server
        whitelistedLibraries:
          - Movies
```

## Configuration - Discord

The "Share your detected activities with others" setting must be enabled under Discord Settings → Activity Settings → Activity Privacy.

![Discord Activity Privacy](assets/readme/discord-activity-privacy.png)

## Configuration - Environment Variables

- `DRPP_PLEX_SERVER_NAME_INPUT` - This is used only during the initial setup (when there are no users in the config) as the name of the Plex server to be added to the config file after user authentication. If this isn't set, in interactive environments, the user is prompted for an input, and in non-interactive environments, "ServerName" is used as a placeholder, which can later be changed by editing the config file and restarting the script.
- `DRPP_NO_PIP_INSTALL` - Set this to `true` to skip automatic invocation of pip on script startup to install missing dependencies.

## Run with Docker

### Image

[ghcr.io/phin05/discord-rich-presence-plex](https://ghcr.io/phin05/discord-rich-presence-plex)

Images are available for the following platforms:

- linux/amd64
- linux/arm64
- linux/386
- linux/arm/v7

### Volumes

Mount a directory for persistent data (config file, cache file and log file) at `/app/data`.

The runtime directory where Discord stores its inter-process communication Unix socket file needs to be mounted into the container at `/run/app`. The path for this would be the first non-null value from the values of the following environment variables in the environment Discord is running in: ([source](https://github.com/discord/discord-rpc/blob/963aa9f3e5ce81a4682c6ca3d136cddda614db33/src/connection_unix.cpp#L29C33-L29C33))

- XDG_RUNTIME_DIR
- TMPDIR
- TMP
- TEMP

If all four environment variables aren't set, `/tmp` is used.

For example, if the environment variable `XDG_RUNTIME_DIR` is set to `/run/user/1000`, that would be the runtime directory that needs to be mounted into the container at `/run/app`. If none of the environment variables are set, you need to mount `/tmp` into the container at `/run/app`.

### UID and GID

The environment variables `DRPP_UID` and `DRPP_GID` can be used to specify the UID and GID of the user Discord is running as. You can determine these by running `id` in your terminal as such user.

If both of the above environment variables are set, the script will change the ownership of `/run/app` and its contents to be in line with the specified UID and GID to prevent issues caused due to insufficient permissions. To skip this ownership change, set the environment variable `DRPP_NO_RUNTIME_DIR_CHOWN` to `true`. Skipping this is necessary only in cases where the runtime directory isn't exclusively dedicated for a single user.

The ownership of `/app` and its contents will be changed as well. If both of the above environment variables are set, they will determine the ownership. Otherwise, the existing ownership information of `/run/app` will be used.

### Other Info

If you're running the container for the first time (when there are no users in the config), set the `DRPP_PLEX_SERVER_NAME_INPUT` environment variable to the name of the Plex server to be added to the config file after user authentication, and check the container logs for the authentication link.

### Docker Compose example

```yaml
services:
  drpp:
    container_name: drpp
    image: ghcr.io/phin05/discord-rich-presence-plex:latest
    restart: unless-stopped
    environment:
      DRPP_UID: 1000
      DRPP_GID: 1000
    volumes:
      - ./data:/app/data
      - /run/user/1000:/run/app
```

### Containerised Discord

If you wish to run Discord in a container as well, you need to mount a designated directory from the host machine into your Discord container at the path where Discord would store its Unix socket file. You can determine this path by checking the environment variables inside the container as per the [volumes](#volumes) section above, or you can set one of the environment variables yourself. That same host directory needs to be mounted into this script's container at `/run/app`. Ensure that the designated directory being mounted into the containers is owned by the user the containerised Discord process is running as.

Depending on the Discord container image you're using, there might be a lot of resource usage overhead and other complications.

#### Docker Compose example using [kasmweb/discord](https://hub.docker.com/r/kasmweb/discord)

```yaml
services:
  kasmcord:
    container_name: kasmcord
    image: kasmweb/discord:1.14.0
    restart: unless-stopped
    ports:
      - 127.0.0.1:6901:6901
    shm_size: 512m
    environment:
      VNC_PW: password
      XDG_RUNTIME_DIR: /run/user/1000
    volumes:
      - ./kasmcord:/run/user/1000
    user: "0"
    entrypoint: sh -c "chmod 700 /run/user/1000 && chown -R kasm-user:kasm-user /run/user/1000 && su kasm-user -c '/dockerstartup/kasm_default_profile.sh /dockerstartup/vnc_startup.sh /dockerstartup/kasm_startup.sh'"
  drpp:
    container_name: drpp
    image: ghcr.io/phin05/discord-rich-presence-plex:latest
    restart: unless-stopped
    volumes:
      - ./drpp:/app/data
      - ./kasmcord:/run/app:ro
    depends_on:
      - kasmcord
```

### Docker on Windows and macOS

The container image for this script is based on Linux. Docker uses virtualisation to run Linux containers on Windows and macOS. In such cases, if you want to run this script's container, you need to run Discord in a Linux container as well, as per the instructions above.
