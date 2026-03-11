# Discord Rich Presence for Plex

Discord Rich Presence for Plex (DRPP) is an application that displays your [Plex](https://www.plex.tv/) status on [Discord](https://discord.com/) using [Rich Presence](https://docs.discord.com/developers/rich-presence/overview).

[![Latest Release](https://img.shields.io/github/v/release/phin05/discord-rich-presence-plex?label=Latest%20Release)](https://github.com/phin05/discord-rich-presence-plex/releases/latest)
[![Build Status](https://img.shields.io/github/actions/workflow/status/phin05/discord-rich-presence-plex/release.yml?label=Build&logo=github)](https://github.com/phin05/discord-rich-presence-plex/actions/workflows/release.yml)

<img width="646" height="276" alt="Showcase" src="https://github.com/user-attachments/assets/a8064002-0b3e-43d4-96d5-74864e514cf3" />

## Features

- Automatically displays your Plex playback activity (movies, TV shows, music, clips) on Discord with poster artwork, progress, buttons, external links, and metadata.
- Web-based user interface for managing all settings, interactive Plex authentication, and viewing live logs.
- Ability to customise all fields shown in your Rich Presence for each media type using template strings. [[Example](https://github.com/user-attachments/assets/eb3fa89d-eca3-4b81-81c4-60e33517b57b)]
- Support for Windows, Linux (including Docker), and macOS.

## Usage

### Getting Started

If you're on Linux, you can [run DRPP with Docker](#docker).

1. Download the [latest release](https://github.com/phin05/discord-rich-presence-plex/releases/latest) for your platform and extract it.
2. Run the executable file.
3. Open the web interface to configure DRPP:
   - The web interface will launch automatically in your default browser, or
   - Navigate to [http://localhost:8040](http://localhost:8040), or
   - Click the DRPP icon in your system tray and select **Web UI**.
4. Click **Add User** and complete the interactive Plex authentication flow.

On Windows and Linux, DRPP has an icon in the system tray. Clicking this allows you to launch the web interface in your default browser.

Information about each config property and templating is available in the web interface.

The web interface is meant for local access only and must not be exposed to the internet. Even if you expose it, all connections from external networks are blocked by default.

### Discord Setup

Discord must be running on the same machine as DRPP.

The **"Share my activity"** setting must be enabled in Discord for Rich Presence to work.

Navigate to **Discord Settings → Activity Settings → Activity Privacy** to enable it.

### CLI Flags

| Flag                | Environment Variable   | Default Value                   | Description              |
| ------------------- | ---------------------- | ------------------------------- | ------------------------ |
| `--data-dir`        | `DRPP_DATA_DIR`        | `data` (relative to executable) | Path to data directory   |
| `--config-file`     | `DRPP_CONFIG_FILE`     | `config.yml` (inside data dir)  | Path to config file      |
| `--cache-file`      | `DRPP_CACHE_FILE`      | `cache.json` (inside data dir)  | Path to cache file       |
| `--log-file`        | `DRPP_LOG_FILE`        | (disabled)                      | Path to log file         |
| `--disable-web-ui`  | `DRPP_DISABLE_WEB_UI`  | `false`                         | Disable web interface    |
| `--disable-systray` | `DRPP_DISABLE_SYSTRAY` | `false`                         | Disable system tray icon |

Environment variables take precedence over default values but are overridden by explicitly passed flags.

## Docker

### Image

[ghcr.io/phin05/discord-rich-presence-plex](https://ghcr.io/phin05/discord-rich-presence-plex)

### Docker Compose Example

```yaml
services:
  drpp:
    container_name: drpp
    image: ghcr.io/phin05/discord-rich-presence-plex:latest
    restart: unless-stopped
    ports:
      - 127.0.0.1:8040:8040
    environment:
      DRPP_UID: 1000
      DRPP_GID: 1000
    volumes:
      - ./data:/app/data
      - /run/user/1000:/run/app
```

Once DRPP is running, navigate to [http://localhost:8040](http://localhost:8040). Click **Add User** and complete the interactive Plex authentication flow.

Make sure to bind the port only to `127.0.0.1` as shown above.

### Volumes

| Path        | Purpose                                                     |
| ----------- | ----------------------------------------------------------- |
| `/app/data` | Directory for persistent config and cache storage           |
| `/run/app`  | Discord's runtime directory for inter-process communication |

Discord's runtime directory is determined by the first set environment variable from the list below, checked in the environment where Discord is running:

1. `XDG_RUNTIME_DIR`
2. `TMPDIR`
3. `TMP`
4. `TEMP`

If none are set, `/tmp` is used.

For example, if `XDG_RUNTIME_DIR` is `/run/user/1000`, mount that directory into the container at `/run/app`.

### Environment Variables

| Variable                    | Description                                            |
| --------------------------- | ------------------------------------------------------ |
| `DRPP_UID`                  | UID of the user running Discord (find by running `id`) |
| `DRPP_GID`                  | GID of the user running Discord (find by running `id`) |
| `DRPP_NO_RUNTIME_DIR_CHOWN` | Set to `true` to skip ownership change of `/run/app`   |

When both `DRPP_UID` and `DRPP_GID` are set, DRPP changes ownership of `/run/app` and `/app` to match the specified UID and GID to prevent permission issues.

### Containerised Discord

To run Discord in a container as well, mount a shared directory from the host into both the Discord container (as Discord's runtime directory) and the DRPP container (at `/run/app`). Ensure that the shared directory is owned by the user the containerised Discord process runs as.

<details>

<summary>Example</summary>

**Docker Compose example using [kasmweb/discord](https://hub.docker.com/r/kasmweb/discord)**

```yaml
services:
  kasmweb-discord:
    container_name: kasmweb-discord
    image: kasmweb/discord:1.14.0
    restart: unless-stopped
    ports:
      - 127.0.0.1:6901:6901
    shm_size: 512m
    environment:
      VNC_PW: password
      XDG_RUNTIME_DIR: /run/user/1000
    volumes:
      - ./discord:/home/kasm-user/.config/discord
      - ./runtime:/run/user/1000
    user: 0:0
    entrypoint: sh -c "chown -R kasm-user:kasm-user /home/kasm-user && chmod 700 /run/user/1000 && chown -R kasm-user:kasm-user /run/user/1000 && su kasm-user -c '/dockerstartup/kasm_default_profile.sh /dockerstartup/vnc_startup.sh /dockerstartup/kasm_startup.sh'"
  drpp:
    container_name: drpp
    image: ghcr.io/phin05/discord-rich-presence-plex:latest
    restart: unless-stopped
    ports:
      - 127.0.0.1:8040:8040
    volumes:
      - ./data:/app/data
      - ./runtime:/run/app:ro
    depends_on:
      - kasmweb-discord
```

</details>

### Docker on Windows and macOS

DRPP's container image is Linux-based. Docker uses virtualisation to run Linux containers on Windows and macOS. In such cases, for DRPP to communicate with Discord, Discord needs to run in a Linux container as well, as per the instructions above.

## Other Information

### Image Upload Providers

DRPP downloads poster images from Plex and uploads them to an external image host for displaying in Discord. The following providers are available:

- [Litterbox](https://litterbox.catbox.moe/) (default)
- [ImgBB](https://imgbb.com/)
- [Imgur](https://imgur.com/)
- [Copyparty](https://github.com/9001/copyparty) (self-hosted)
