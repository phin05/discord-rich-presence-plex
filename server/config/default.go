package config

import "os"

var Containerised = os.Getenv("DRPP_CONTAINERISED") == "true"

func newDefaultConfig() Config {
	var launchOnStartup bool
	var bindAddress string
	var allowedNetworks []string
	if Containerised {
		// TODO: Detect the container's interface and subnet
		bindAddress = "0.0.0.0"
		allowedNetworks = []string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"}
	} else {
		bindAddress = "127.0.0.1"
		launchOnStartup = true
	}
	return Config{
		Version: currentVersion,
		Web: Web{
			LaunchOnStartup: launchOnStartup,
			BindAddress:     bindAddress,
			BindPort:        8040, // TODO: Auto-select an available port if this is unavailable
			AllowedNetworks: allowedNetworks,
		},
		Discord: Discord{
			ClientId:           "413407336082833418",
			IpcPipeNumber:      -1,
			IpcTimeoutSeconds:  10,
			RateLimit:          5,
			StopTimeoutSeconds: 3,
			IdleTimeoutSeconds: 25,
			DisplayRules: DisplayRules{
				Movie: DisplayRule{
					Name:         "{{ .Title }}",
					Details:      "{{ .Title }} ({{ .Year }})",
					State:        "{{ .Genres }}",
					StatusType:   "name",
					LargeImage:   "{{ .Poster }}",
					SmallImage:   `{{ if eq .State "paused" }}paused{{ end }}`,
					SmallText:    `{{ if eq .State "paused" }}Paused{{ end }}`,
					ProgressMode: "bar",
				},
				Episode: DisplayRule{
					Name:         "{{ .ShowTitle }}",
					Details:      "{{ .ShowTitle }} ({{ .ShowYear }})",
					State:        "S{{ .SeasonNumber }}E{{ .EpisodeNumber }} · {{ .EpisodeTitle }}",
					StatusType:   "name",
					LargeImage:   "{{ .ShowPoster }}",
					SmallImage:   `{{ if eq .State "paused" }}paused{{ end }}`,
					SmallText:    `{{ if eq .State "paused" }}Paused{{ end }}`,
					ProgressMode: "bar",
				},
				Track: DisplayRule{
					Name:         "{{ .Artist }}",
					Details:      "{{ .Title }}",
					State:        "{{ .Artist }}",
					StatusType:   "name",
					LargeImage:   "{{ .AlbumPoster }}",
					LargeText:    "{{ .Album }} ({{ .Year }})",
					SmallImage:   `{{ if eq .State "paused" }}paused{{ else }}{{ .ArtistPoster }}{{ end }}`,
					SmallText:    `{{ if eq .State "paused" }}Paused{{ else }}{{ .AlbumArtist }}{{ end }}`,
					ProgressMode: "bar",
				},
				Clip: DisplayRule{
					Name:         "{{ .Title }}",
					Details:      "{{ .Title }}",
					StatusType:   "name",
					LargeImage:   "{{ .Poster }}",
					SmallImage:   `{{ if eq .State "paused" }}paused{{ end }}`,
					SmallText:    `{{ if eq .State "paused" }}Paused{{ end }}`,
					ProgressMode: "bar",
				},
				LiveEpisode: DisplayRule{
					Name:         "{{ .ShowTitle }}",
					Details:      "{{ .ShowTitle }}",
					State:        "{{ .EpisodeTitle }}",
					StatusType:   "name",
					LargeImage:   "{{ .ShowPoster }}",
					SmallImage:   `{{ if eq .State "paused" }}paused{{ end }}`,
					SmallText:    `{{ if eq .State "paused" }}Paused{{ end }}`,
					ProgressMode: "elapsed",
				},
			},
		},
		Images: Images{
			FitInSquare:          true,
			MaxSize:              256,
			UploadTimeoutSeconds: 10,
			Uploaders: Uploaders{
				Litterbox: Litterbox{
					Enabled:     true,
					ExpiryHours: 72,
				},
				ImgBb: ImgBb{
					ExpiryMinutes: 72 * 60,
				},
				Copyparty: Copyparty{
					ExpiryMinutes: 72 * 60,
				},
			},
		},
	}
}
