package plex

import (
	"context"
	"drpp/server/config"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"
)

const connCheckInterval = 1 * time.Minute

type Activity struct {
	MediaType         string                                            `json:"mediaType"`
	State             string                                            `json:"state"`
	ElapsedDurationMs int64                                             `json:"elapsedDurationMs"`
	Item              Metadata                                          `json:"item"`
	ParentItem        Metadata                                          `json:"parentItem,omitzero"`
	GrandparentItem   Metadata                                          `json:"grandparentItem,omitzero"`
	GetThumbUrl       func(endpoint string) (string, map[string]string) `json:"-"`
}

type Service struct {
	userToken     string
	serverConfig  config.Server
	fatalShutdown context.CancelFunc
	logger        *prefixedLogger

	mu     sync.Mutex
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewService(userToken string, serverConfig config.Server, fatalShutdown context.CancelFunc) *Service {
	return &Service{
		userToken:     userToken,
		serverConfig:  serverConfig,
		fatalShutdown: fatalShutdown,
		logger:        newPrefixedLogger(serverConfig.Name),
	}
}

func (s *Service) Start(callback func(activity *Activity)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cancel != nil {
		return
	}
	var ctx context.Context
	ctx, s.cancel = context.WithCancel(context.Background())
	s.wg.Go(func() {
		var retries int
		for {
			err := s.run(ctx, callback)
			if ctx.Err() != nil {
				return
			}
			if err == nil {
				s.logger.Info("Reconnecting in %d seconds", s.serverConfig.RetryIntervalSeconds)
				retries = 0
			} else {
				s.logger.Error(err, "Failed to initialise")
				if s.serverConfig.MaxRetriesBeforeExit > -1 && retries >= s.serverConfig.MaxRetriesBeforeExit {
					s.logger.Error(nil, "Max retries exceeded, shutting down")
					s.fatalShutdown()
					return
				}
				s.logger.Info("Retrying in %d seconds", s.serverConfig.RetryIntervalSeconds)
				retries++
			}
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Duration(s.serverConfig.RetryIntervalSeconds) * time.Second):
			}
		}
	})
}

func (s *Service) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cancel == nil {
		return
	}
	s.cancel()
	s.cancel = nil
	s.wg.Wait()
	s.logger.Info("Disconnected")
}

func (s *Service) run(ctx context.Context, callback func(activity *Activity)) error {

	client := NewClient("", s.userToken, time.Duration(s.serverConfig.RequestTimeoutSeconds)*time.Second)

	localCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	account, err := client.GetAccount(localCtx)
	if err != nil {
		return fmt.Errorf("fetch account details: %w", err)
	}
	if s.serverConfig.ListenForUser == "" {
		s.serverConfig.ListenForUser = account.User.UsernameOrTitle()
	}
	s.logger.Info("Signed in as user %q, listening for user %q", account.User.UsernameOrTitle(), s.serverConfig.ListenForUser)

	servers, err := client.GetServers(localCtx)
	if err != nil {
		return fmt.Errorf("fetch servers: %w", err)
	}
	if len(servers) == 0 {
		return errors.New("no servers found")
	}

	serverIndex := slices.IndexFunc(servers, func(server Resource) bool {
		return strings.EqualFold(server.Name, s.serverConfig.Name)
	})
	if serverIndex == -1 {
		return errors.New("specified server not found")
	}
	server := servers[serverIndex]

	client.Token = server.AccessToken

	var serverUrls []string
	if s.serverConfig.Url == "" {
		slices.SortStableFunc(server.Connections, func(c1, c2 Connection) int {
			if c1.Local != c2.Local {
				if c1.Local {
					return -1
				}
				return 1
			}
			if c1.Relay != c2.Relay {
				if c1.Relay {
					return 1
				}
				return -1
			}
			return 0
		})
		for _, connection := range server.Connections {
			serverUrls = append(serverUrls, connection.Uri)
		}
	} else {
		serverUrls = append(serverUrls, s.serverConfig.Url)
	}

	activityCh := make(chan *Activity, 1)

	itemCache := make(map[string]*Metadata)
	getItem := func(ratingKey string, itemType string) *Metadata {
		if ratingKey == "" {
			return nil
		}
		if val, ok := itemCache[ratingKey]; ok {
			return val
		}
		item, err := client.GetMetadata(localCtx, ratingKey)
		if err != nil {
			s.logger.Error(err, "Failed to fetch metadata for %s item %q", itemType, ratingKey)
			return nil
		}
		itemCache[ratingKey] = item
		return item
	}

	var lastSessionKey, lastState string

	handler := func(n *NotificationContainer) {

		if n.Type != "playing" || len(n.PlaySessionStateNotification) == 0 {
			return
		}
		notification := n.PlaySessionStateNotification[0]
		s.logger.Debug("Notification: %#v", notification)

		switch notification.State {
		case "playing", "paused", "stopped":
		default:
			s.logger.Debug("Unrecognised state %q, ignoring", notification.State)
			return
		}

		if notification.SessionKey != lastSessionKey {
			if server.Owned {
				sessions, err := client.GetSessions(localCtx)
				if err != nil {
					s.logger.Error(err, "Failed to fetch sessions")
					return
				}
				sessionIndex := slices.IndexFunc(sessions, func(session Metadata) bool {
					return session.SessionKey == notification.SessionKey
				})
				if sessionIndex == -1 {
					s.logger.Debug("Session key %q not found in sessions list, ignoring", notification.SessionKey)
					return
				}
				sessionUsername := sessions[sessionIndex].User.UsernameOrTitle()
				isOwnSession := strings.EqualFold(sessionUsername, s.serverConfig.ListenForUser)
				if !isOwnSession {
					s.logger.Debug("Session key %q belongs to user %q, not %q, ignoring", notification.SessionKey, sessionUsername, s.serverConfig.ListenForUser)
					return
				}
			}
			if notification.State != "playing" && lastState == "playing" {
				s.logger.Debug("Ignoring transition to non-playing state while a different playing session is active")
				return
			}
			clear(itemCache)
		}

		lastSessionKey = notification.SessionKey
		lastState = notification.State

		item := getItem(notification.RatingKey, "main")
		if item == nil {
			return
		}

		var mediaType string
		if strings.HasPrefix(item.Key, "/livetv") {
			mediaType = "liveEpisode"
		} else {
			mediaType = item.Type
		}

		switch mediaType {
		case "movie", "episode", "track", "clip", "liveEpisode":
		default:
			s.logger.Debug("Unrecognised media type %q, ignoring", mediaType)
			return
		}

		libraryName := strings.TrimSpace(item.LibrarySectionTitle)
		if libraryName == "" {
			s.logger.Debug("Library name is empty, ignoring blacklist/whitelist")
		} else {
			if len(s.serverConfig.BlacklistedLibraries) > 0 {
				if slices.IndexFunc(s.serverConfig.BlacklistedLibraries, func(l string) bool {
					return strings.EqualFold(l, libraryName)
				}) != -1 {
					s.logger.Debug("Library %q is blacklisted, ignoring", libraryName)
					return
				}
			}
			if len(s.serverConfig.WhitelistedLibraries) > 0 {
				if slices.IndexFunc(s.serverConfig.WhitelistedLibraries, func(l string) bool {
					return strings.EqualFold(l, libraryName)
				}) == -1 {
					s.logger.Debug("Library %q is not whitelisted, ignoring", libraryName)
					return
				}
			}
		}

		activity := &Activity{
			MediaType:         mediaType,
			State:             notification.State,
			ElapsedDurationMs: notification.ViewOffsetMs,
			Item:              *item,
			GetThumbUrl:       client.GetThumbUrl,
		}
		parentItem := getItem(item.ParentRatingKey, "parent")
		if parentItem != nil {
			activity.ParentItem = *parentItem
		}
		grandparentItem := getItem(item.GrandparentRatingKey, "grandparent")
		if grandparentItem != nil {
			activity.GrandparentItem = *grandparentItem
		}
		// Drain any stale pending activity
		select {
		case <-activityCh:
		default:
		}
		activityCh <- activity

	}

	errorHandler := func(err error) {
		if localCtx.Err() != nil {
			return
		}
		s.logger.Error(err, "Restarting due to notification listener error")
		cancel()
	}

	var wg sync.WaitGroup

	var errs []error
	for _, serverUrl := range serverUrls {
		client.BaseUrl = strings.TrimRight(serverUrl, "/")
		s.logger.Debug("Connecting to %s", client.BaseUrl)
		if err := client.StartNotificationListener(localCtx, &wg, handler, errorHandler); err != nil {
			errs = append(errs, err)
			continue
		}
		errs = nil
		break
	}
	if len(errs) > 0 {
		return fmt.Errorf("start notification listener: %w", errors.Join(errs...))
	}
	s.logger.Info("Connected to server")

	wg.Go(func() {
		for {
			select {
			case <-localCtx.Done():
				return
			case activity := <-activityCh:
				callback(activity)
			}
		}
	})

	wg.Go(func() {
		connChecker := time.NewTicker(connCheckInterval)
		defer connChecker.Stop()
		for {
			select {
			case <-localCtx.Done():
				return
			case <-connChecker.C:
				if err := client.TestServerConnection(localCtx); err != nil {
					s.logger.Error(err, "Restarting due to connection check error")
					cancel()
					return
				}
			}
		}
	})

	wg.Wait()

	return nil

}
