package plex

import (
	"context"
	"drpp/server/logger"
	"encoding/json/v2"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	clientId         = "discord-rich-presence-plex"
	maxRespBodyBytes = 8 * 1024 * 1024 // 8 MB
	wsPingInterval   = 10 * time.Second
	wsPongTimeout    = 15 * time.Second
	wsWriteTimeout   = 15 * time.Second
)

type Client struct {
	BaseUrl    string
	Token      string
	httpClient *http.Client
}

func NewClient(baseUrl string, token string, timeout time.Duration) *Client {
	return &Client{
		BaseUrl:    baseUrl,
		Token:      token,
		httpClient: &http.Client{Timeout: timeout},
	}
}

func httpGetJson[T any](ctx context.Context, httpClient *http.Client, token string, url string) (T, error) {
	var result T
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return result, fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Plex-Client-Identifier", clientId)
	// TODO: Add more headers - https://developer.plex.tv/pms/#section/API-Info/Headers
	req.Header.Set("X-Plex-Token", token)
	resp, err := httpClient.Do(req)
	if err != nil {
		return result, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, maxRespBodyBytes))
		logger.Debug("GET %s, HTTP %d, %s, %s", url, resp.StatusCode, resp.Header, string(bodyBytes))
		return result, fmt.Errorf("HTTP status %s", resp.Status)
	}
	if err := json.UnmarshalRead(io.LimitReader(resp.Body, maxRespBodyBytes), &result); err != nil {
		return result, fmt.Errorf("unmarshal: %w", err)
	}
	return result, nil
}

type Account struct {
	User User `json:"user"`
}

type User struct {
	Title string `json:"title"`
}

func (c *Client) GetAccount(ctx context.Context) (*Account, error) {
	return httpGetJson[*Account](ctx, c.httpClient, c.Token, "https://plex.tv/users/account.json")
}

type Resource struct {
	Name        string       `json:"name"`
	Provides    string       `json:"provides"`
	Owned       bool         `json:"owned"`
	AccessToken string       `json:"accessToken"`
	Connections []Connection `json:"connections"`
}

type Connection struct {
	Protocol string `json:"protocol"`
	Address  string `json:"address"`
	Port     int64  `json:"port"`
	Uri      string `json:"uri"`
	Local    bool   `json:"local"`
	Relay    bool   `json:"relay"`
}

func (c *Client) GetServers(ctx context.Context) ([]Resource, error) {
	resources, err := httpGetJson[[]Resource](ctx, c.httpClient, c.Token, "https://clients.plex.tv/api/v2/resources?includeHttps=1&includeRelay=1&includeIPv6=1")
	if err != nil {
		return nil, err
	}
	var servers []Resource
	for _, resource := range resources {
		if resource.Provides == "server" {
			servers = append(servers, resource)
		}
	}
	return servers, nil
}

func (c *Client) TestServerConnection(ctx context.Context) error {
	_, err := httpGetJson[any](ctx, c.httpClient, c.Token, c.BaseUrl+"/identity")
	return err
}

func (c *Client) GetThumbUrl(thumb string) (string, map[string]string) {
	return c.BaseUrl + thumb, map[string]string{"X-Plex-Token": c.Token}
}

type MetadataResponse struct {
	MediaContainer MediaContainer `json:"MediaContainer"`
}

type MediaContainer struct {
	Metadata []Metadata `json:"Metadata"`
}

type Metadata struct {
	Type                 string  `json:"type,omitzero"`
	LibrarySectionTitle  string  `json:"librarySectionTitle,omitzero"`
	Key                  string  `json:"key,omitzero"`
	RatingKey            string  `json:"ratingKey,omitzero"`
	ParentRatingKey      string  `json:"parentRatingKey,omitzero"`
	GrandparentRatingKey string  `json:"grandparentRatingKey,omitzero"`
	Title                string  `json:"title,omitzero"`
	OriginalTitle        string  `json:"originalTitle,omitzero"`
	Year                 int64   `json:"year,omitzero"`
	Index                int64   `json:"index,omitzero"`
	Thumb                string  `json:"thumb,omitzero"`
	DurationMs           int64   `json:"duration,omitzero"`
	Genres               []Genre `json:"Genre,omitempty"`
	Guids                []Guid  `json:"Guid,omitempty"`
	SessionKey           string  `json:"sessionKey,omitzero"`
	User                 User    `json:"User,omitzero"`
	GrandparentThumb     string  `json:"grandparentThumb,omitzero"` // For live episodes
	GrandparentTitle     string  `json:"grandparentTitle,omitzero"` // For live episodes
}

type Genre struct {
	Tag string `json:"tag"`
}

type Guid struct {
	Id string `json:"id"`
}

func (c *Client) GetSessions(ctx context.Context) ([]Metadata, error) {
	resp, err := httpGetJson[*MetadataResponse](ctx, c.httpClient, c.Token, c.BaseUrl+"/status/sessions")
	if err != nil {
		return nil, err
	}
	return resp.MediaContainer.Metadata, nil
}

func (c *Client) GetMetadata(ctx context.Context, ratingKey string) (*Metadata, error) {
	resp, err := httpGetJson[*MetadataResponse](ctx, c.httpClient, c.Token, fmt.Sprintf("%s/library/metadata/%s", c.BaseUrl, ratingKey))
	if err != nil {
		return nil, fmt.Errorf("fetch metadata: %w", err)
	}
	if len(resp.MediaContainer.Metadata) == 0 {
		return nil, errors.New("no metadata returned")
	}
	return &resp.MediaContainer.Metadata[0], nil
}

type Notification struct {
	NotificationContainer NotificationContainer `json:"NotificationContainer"`
}

type NotificationContainer struct {
	Type                         string                         `json:"type"`
	PlaySessionStateNotification []PlaySessionStateNotification `json:"PlaySessionStateNotification"`
}

type PlaySessionStateNotification struct {
	State        string `json:"state"`
	SessionKey   string `json:"sessionKey"`
	RatingKey    string `json:"ratingKey"`
	ViewOffsetMs int64  `json:"viewOffset"`
}

func (c *Client) StartNotificationListener(ctx context.Context, wg *sync.WaitGroup, handler func(*NotificationContainer), errorHandler func(error)) error {
	baseUrl, err := url.Parse(c.BaseUrl)
	if err != nil {
		return fmt.Errorf("parse URL: %w", err)
	}
	var scheme string
	switch baseUrl.Scheme {
	case "http":
		scheme = "ws"
	case "https":
		scheme = "wss"
	default:
		return fmt.Errorf("unsupported URL scheme %q", baseUrl.Scheme)
	}
	wsUrl := url.URL{Scheme: scheme, Host: baseUrl.Host, Path: "/:/websockets/notifications"}
	wsHeaders := http.Header{
		"X-Plex-Client-Identifier": []string{clientId},
		"X-Plex-Token":             []string{c.Token},
	}
	conn, resp, err := websocket.DefaultDialer.DialContext(ctx, wsUrl.String(), wsHeaders)
	if err != nil {
		return fmt.Errorf("connect websocket: %w", err)
	}
	defer resp.Body.Close() // gorilla/websocket returns a response with a NopCloser body, so this doesn't close the websocket connection
	wg.Go(func() {
		for {
			_, data, err := conn.ReadMessage()
			if ctx.Err() != nil {
				return
			}
			if err != nil {
				if !websocket.IsCloseError(err) || websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					errorHandler(fmt.Errorf("read: %w", err))
				}
				return
			}
			var n Notification
			if err := json.Unmarshal(data, &n); err != nil {
				errorHandler(fmt.Errorf("unmarshal: %w", err))
				return
			}
			handler(&n.NotificationContainer)
		}
	})
	wg.Go(func() {
		pingWriter := time.NewTicker(wsPingInterval)
		defer pingWriter.Stop()
		for {
			select {
			case <-ctx.Done():
				_ = conn.SetWriteDeadline(time.Now().Add(wsWriteTimeout))
				if err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")); err != nil {
					errorHandler(fmt.Errorf("write close: %w", err))
				}
				_ = conn.Close()
				return
			case <-pingWriter.C:
				_ = conn.SetWriteDeadline(time.Now().Add(wsWriteTimeout))
				if err := conn.WriteMessage(websocket.PingMessage, []byte("ping")); err != nil {
					errorHandler(fmt.Errorf("ping: %w", err))
				}
			}
		}
	})
	_ = conn.SetReadDeadline(time.Now().Add(wsPongTimeout))
	conn.SetPongHandler(func(data string) error {
		// TODO: Check if pong data matches the ping data
		_ = conn.SetReadDeadline(time.Now().Add(wsPongTimeout))
		return nil
	})
	return nil
}
