package discord

type handshake struct {
	V        int    `json:"v"`
	ClientId string `json:"client_id"`
}

type frame struct {
	Cmd   frameCmd   `json:"cmd"`
	Args  *frameArgs `json:"args"`
	Nonce string     `json:"nonce"`
}

type frameCmd string

const (
	frameCmdSetActivity frameCmd = "SET_ACTIVITY"
)

type frameArgs struct {
	Pid      int       `json:"pid"`
	Activity *Activity `json:"activity"`
}

type ipcResponse struct {
	Evt     string `json:"evt"`
	Message string `json:"message"`
	Data    struct {
		Message string `json:"message"`
		User    struct {
			Username string `json:"username"`
		} `json:"user"`
	} `json:"data"`
}

// https://docs.discord.com/developers/events/gateway-events#activity-object
// https://discord.com/developers/docs/social-sdk/classdiscordpp_1_1Activity.html
type Activity struct {
	Name              string                    `json:"name,omitzero"`
	Type              ActivityType              `json:"type"`
	StatusDisplayType ActivityStatusDisplayType `json:"status_display_type"`
	Details           string                    `json:"details,omitzero"`
	DetailsUrl        string                    `json:"details_url,omitzero"`
	State             string                    `json:"state,omitzero"`
	StateUrl          string                    `json:"state_url,omitzero"`
	Assets            ActivityAssets            `json:"assets,omitzero"`
	Timestamps        ActivityTimestamps        `json:"timestamps,omitzero"`
	Buttons           []ActivityButton          `json:"buttons,omitempty"`
}

type ActivityType uint8

const (
	ActivityTypeListening ActivityType = iota + 2
	ActivityTypeWatching
)

type ActivityStatusDisplayType uint8

const (
	ActivityStatusDisplayTypeName ActivityStatusDisplayType = iota
	ActivityStatusDisplayTypeState
	ActivityStatusDisplayTypeDetails
)

type ActivityAssets struct {
	LargeImage string `json:"large_image,omitzero"`
	LargeText  string `json:"large_text,omitzero"`
	LargeUrl   string `json:"large_url,omitzero"`
	SmallImage string `json:"small_image,omitzero"`
	SmallText  string `json:"small_text,omitzero"`
	SmallUrl   string `json:"small_url,omitzero"`
}

type ActivityTimestamps struct {
	StartMs int64 `json:"start,omitzero"`
	EndMs   int64 `json:"end,omitzero"`
}

type ActivityButton struct {
	Label string `json:"label,omitzero"`
	Url   string `json:"url,omitzero"`
}
