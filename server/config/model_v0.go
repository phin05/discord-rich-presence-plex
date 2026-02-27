package config

type configV0 struct {
	// Logging loggingV0 `json:"logging" yaml:"logging"`
	Display displayV0 `json:"display" yaml:"display"`
	Users   []userV0  `json:"users" yaml:"users"`
}

// type loggingV0 struct {
// 	Debug       bool `json:"debug" yaml:"debug"`
// 	WriteToFile bool `json:"writeToFile" yaml:"writeToFile"`
// }

type displayV0 struct {
	// Duration       bool             `json:"duration" yaml:"duration"`
	// Genres         bool             `json:"genres" yaml:"genres"`
	// Album          bool             `json:"album" yaml:"album"`
	// AlbumImage     bool             `json:"albumImage" yaml:"albumImage"`
	// Artist         bool             `json:"artist" yaml:"artist"`
	// ArtistImage    bool             `json:"artistImage" yaml:"artistImage"`
	// Year           bool             `json:"year" yaml:"year"`
	// StatusIcon     bool             `json:"statusIcon" yaml:"statusIcon"`
	// ProgressMode   string           `json:"progressMode" yaml:"progressMode"`
	// StatusTextType statusTextTypeV0 `json:"statusTextType" yaml:"statusTextType"`
	// Paused         bool             `json:"paused" yaml:"paused"`
	Posters postersV0 `json:"posters" yaml:"posters"`
	// Buttons        []buttonV0       `json:"buttons" yaml:"buttons"`
}

// type statusTextTypeV0 struct {
// 	Watching  string `json:"watching" yaml:"watching"`
// 	Listening string `json:"listening" yaml:"listening"`
// }

type postersV0 struct {
	Enabled       bool   `json:"enabled" yaml:"enabled"`
	ImgurClientId string `json:"imgurClientID" yaml:"imgurClientID"`
	MaxSize       int    `json:"maxSize" yaml:"maxSize"`
	Fit           bool   `json:"fit" yaml:"fit"`
}

// type buttonV0 struct {
// 	Label      string   `json:"label" yaml:"label"`
// 	Url        string   `json:"url" yaml:"url"`
// 	MediaTypes []string `json:"mediaTypes" yaml:"mediaTypes"`
// }

type userV0 struct {
	Token   string     `json:"token" yaml:"token"`
	Servers []serverV0 `json:"servers" yaml:"servers"`
}

type serverV0 struct {
	Name                 string   `json:"name" yaml:"name"`
	ListenForUser        string   `json:"listenForUser" yaml:"listenForUser"`
	BlacklistedLibraries []string `json:"blacklistedLibraries" yaml:"blacklistedLibraries"`
	WhitelistedLibraries []string `json:"whitelistedLibraries" yaml:"whitelistedLibraries"`
	// IpcPipeNumber        int      `json:"ipcPipeNumber" yaml:"ipcPipeNumber"`
}
