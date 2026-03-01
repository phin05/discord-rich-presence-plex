package mediator

import (
	"bytes"
	"context"
	"drpp/server/cache"
	"drpp/server/config"
	"drpp/server/discord"
	"drpp/server/images"
	"drpp/server/logger"
	"drpp/server/plex"
	"encoding/json/v2"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"text/template"
	"time"
)

const maxUploadAttempts = 5

type ImageService interface {
	Upload(ctx context.Context, pngBytes []byte) (string, error)
	Lifespan() time.Duration
}

type Service struct {
	discordService *discord.Service
	plexServices   []*plex.Service
	cacheService   *cache.Service
	imageService   ImageService
	imagesConfig   config.Images
	displayRules   config.DisplayRules
	ipcTimeout     time.Duration
	stopTimeout    time.Duration
	idleTimeout    time.Duration

	lastState string
	stopTimer *time.Timer

	mu      sync.Mutex
	running bool
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

func NewService(
	discordService *discord.Service,
	plexServices []*plex.Service,
	cacheService *cache.Service,
	imageService ImageService,
	imagesConfig config.Images,
	discordConfig config.Discord,
) *Service {
	return &Service{
		discordService: discordService,
		plexServices:   plexServices,
		cacheService:   cacheService,
		imageService:   imageService,
		imagesConfig:   imagesConfig,
		displayRules:   discordConfig.DisplayRules,
		ipcTimeout:     time.Duration(discordConfig.IpcTimeoutSeconds) * time.Second,
		stopTimeout:    time.Duration(discordConfig.StopTimeoutSeconds) * time.Second,
		idleTimeout:    time.Duration(discordConfig.IdleTimeoutSeconds) * time.Second,
	}
}

func (s *Service) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running {
		return
	}
	s.running = true
	for _, plexService := range s.plexServices {
		plexService.Start(s.plexCallback)
	}
}

func (s *Service) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.running {
		return
	}
	s.running = false
	for _, plexService := range s.plexServices {
		plexService.Stop()
	}
	if s.cancel == nil {
		return
	}
	s.cancel()
	s.mu.Unlock()
	s.wg.Wait()
	s.mu.Lock()
	s.stopActivity()
}

func (s *Service) stopActivity() {
	s.lastState = ""
	s.clearStopTimer()
	s.discordService.Disconnect()
}

func (s *Service) setStopTimer(duration time.Duration) {
	if s.stopTimer != nil {
		return
	}
	s.stopTimer = time.AfterFunc(duration, func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		if !s.running {
			return
		}
		s.stopActivity()
	})
}

func (s *Service) clearStopTimer() {
	if s.stopTimer == nil {
		return
	}
	s.stopTimer.Stop()
	s.stopTimer = nil
}

func (s *Service) plexCallback(activity *plex.Activity) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.running {
		return
	}
	if s.cancel != nil {
		s.cancel()
		s.mu.Unlock()
		s.wg.Wait()
		s.mu.Lock()
		if !s.running {
			return
		}
	}
	var ctx context.Context
	ctx, s.cancel = context.WithCancel(context.Background())
	s.wg.Go(func() {
		s.handlePlexActivity(ctx, activity)
	})
}

func (s *Service) handlePlexActivity(ctx context.Context, activity *plex.Activity) {
	activityJson, _ := json.Marshal(activity)
	logger.Info("Activity: %s", activityJson)
	var rule config.DisplayRule
	ok := func() bool {
		s.mu.Lock()
		defer s.mu.Unlock()
		if activity.State == "stopped" {
			if s.lastState == "" || s.lastState == "stopped" {
				// Stopped without a previous state or timer already set, ignore
				return false
			}
			s.lastState = "stopped"
			s.clearStopTimer()
			s.setStopTimer(s.stopTimeout)
			return false
		}
		switch activity.MediaType {
		case "movie":
			rule = s.displayRules.Movie
		case "episode":
			rule = s.displayRules.Episode
		case "track":
			rule = s.displayRules.Track
		case "clip":
			rule = s.displayRules.Clip
		case "liveEpisode":
			rule = s.displayRules.LiveEpisode
		default:
			return false
		}
		if activity.State == "paused" && rule.PauseTimeoutSeconds >= 0 {
			if s.lastState == "" {
				// Paused without a previous state, with pause timeout set, ignore
				return false
			}
			if rule.PauseTimeoutSeconds == 0 {
				// Paused, with pause timeout set to 0, stop immediately
				s.stopActivity()
				return false
			}
		}
		// Playing, or transitioned to paused, or pause timeout set to -1, so clear idle timer
		if activity.State != "paused" || s.lastState != "paused" || rule.PauseTimeoutSeconds < 0 {
			s.clearStopTimer()
		}
		return true
	}()
	if !ok {
		return
	}
	templateData := make(map[string]any)
	templateData["MediaType"] = activity.MediaType
	templateData["State"] = activity.State
	templateData["LibraryName"] = activity.Item.LibrarySectionTitle
	templateData["ElapsedDurationMs"] = activity.ElapsedDurationMs
	addGuidUrls := func(guids []plex.Guid, prefix string) {
		guidMap := make(map[string]string)
		for _, guid := range guids {
			parts := strings.SplitN(guid.Id, "://", 2)
			if len(parts) == 2 {
				guidMap[parts[0]] = parts[1]
			}
		}
		if id, ok := guidMap["imdb"]; ok {
			templateData[prefix+"ImdbUrl"] = "https://www.imdb.com/title/" + id
		}
		if id, ok := guidMap["tmdb"]; ok {
			var tmdbPathSegment, traktIdType string
			if activity.MediaType == "movie" { //nolint:gocritic
				tmdbPathSegment = "movie"
				traktIdType = "movie"
			} else if prefix == "Episode" {
				tmdbPathSegment = "episode"
				traktIdType = "episode"
			} else {
				tmdbPathSegment = "tv"
				traktIdType = "show"
			}
			templateData[prefix+"TmdbUrl"] = "https://www.themoviedb.org/" + tmdbPathSegment + "/" + id
			templateData[prefix+"TraktUrl"] = "https://trakt.tv/search/tmdb/" + id + "?id_type=" + traktIdType
			if activity.MediaType == "movie" {
				templateData[prefix+"LetterboxdUrl"] = "https://letterboxd.com/tmdb/" + id
			}
		}
		if id, ok := guidMap["tvdb"]; ok {
			var tvdbPathSegment string
			if activity.MediaType == "movie" { //nolint:gocritic
				tvdbPathSegment = "movie"
			} else if prefix == "Episode" {
				tvdbPathSegment = "episode"
			} else {
				tvdbPathSegment = "series"
			}
			templateData[prefix+"TvdbUrl"] = "https://www.thetvdb.com/dereferrer/" + tvdbPathSegment + "/" + id
		}
		if id, ok := guidMap["mbid"]; ok {
			templateData[prefix+"MusicBrainzUrl"] = "https://musicbrainz.org/" + strings.ToLower(prefix) + "/" + id
		}
	}
	switch activity.MediaType {
	case "movie":
		templateData["Title"] = activity.Item.Title
		templateData["Year"] = activity.Item.Year
		templateData["Duration"] = formatDuration(activity.Item.DurationMs, "")
		templateData["Genres"] = formatGenres(activity.Item.Genres, ", ", 3)
		templateData["Poster"] = activity.Item.Thumb
		addGuidUrls(activity.Item.Guids, "")
	case "episode":
		templateData["ShowTitle"] = activity.GrandparentItem.Title
		templateData["ShowYear"] = activity.GrandparentItem.Year
		templateData["EpisodeDuration"] = formatDuration(activity.Item.DurationMs, "")
		templateData["ShowGenres"] = formatGenres(activity.GrandparentItem.Genres, ", ", 3)
		templateData["ShowPoster"] = activity.GrandparentItem.Thumb
		templateData["SeasonNumber"] = activity.ParentItem.Index
		templateData["EpisodeNumber"] = activity.Item.Index
		templateData["EpisodeTitle"] = activity.Item.Title
		addGuidUrls(activity.Item.Guids, "Episode")
		addGuidUrls(activity.GrandparentItem.Guids, "Show")
	case "track":
		templateData["Title"] = activity.Item.Title
		if activity.Item.OriginalTitle != "" {
			templateData["Artist"] = activity.Item.OriginalTitle
		} else {
			templateData["Artist"] = activity.GrandparentItem.Title
		}
		templateData["Album"] = activity.ParentItem.Title
		templateData["Year"] = activity.ParentItem.Year
		templateData["AlbumArtist"] = activity.GrandparentItem.Title
		templateData["AlbumPoster"] = activity.ParentItem.Thumb
		templateData["ArtistPoster"] = activity.GrandparentItem.Thumb
		templateData["Duration"] = formatDuration(activity.Item.DurationMs, "")
		templateData["AlbumGenres"] = formatGenres(activity.ParentItem.Genres, ", ", 3)
		templateData["ArtistGenres"] = formatGenres(activity.GrandparentItem.Genres, ", ", 3)
		addGuidUrls(activity.Item.Guids, "Track")
		addGuidUrls(activity.ParentItem.Guids, "Album")
		addGuidUrls(activity.GrandparentItem.Guids, "Artist")
	case "clip":
		templateData["Title"] = activity.Item.Title
		templateData["Duration"] = formatDuration(activity.Item.DurationMs, "")
		templateData["Poster"] = activity.Item.Thumb
	case "liveEpisode":
		templateData["ShowTitle"] = activity.Item.GrandparentTitle
		templateData["EpisodeTitle"] = activity.Item.Title
		templateData["ShowPoster"] = activity.Item.GrandparentThumb
	}
	templateData["Item"] = activity.Item
	templateData["ParentItem"] = activity.ParentItem
	templateData["GrandparentItem"] = activity.GrandparentItem
	logger.Debug("Template: %#v", templateData)
	resolveImage := func(tmpl string) string {
		thumb := renderTemplate(tmpl, templateData)
		if thumb == "" {
			return ""
		}
		if thumb == activity.Item.Thumb || thumb == activity.ParentItem.Thumb || thumb == activity.GrandparentItem.Thumb {
			var sourceUrl string
			var headers map[string]string
			if parsed, err := url.Parse(thumb); err != nil || !parsed.IsAbs() || (parsed.Scheme != "http" && parsed.Scheme != "https") {
				sourceUrl, headers = activity.GetThumbUrl(thumb)
			} else {
				sourceUrl = thumb
			}
			return s.getUploadedImageUrl(ctx, thumb, sourceUrl, headers)
		}
		return thumb
	}
	var largeImage, smallImage string
	var imageWg sync.WaitGroup
	imageWg.Go(func() { largeImage = resolveImage(rule.LargeImage) })
	imageWg.Go(func() { smallImage = resolveImage(rule.SmallImage) })
	imageWg.Wait()
	discordActivity := &discord.Activity{
		Type:              mapActivityType(activity.MediaType),
		StatusDisplayType: mapStatusDisplayType(rule.StatusType),
		Details:           adjustLength(renderTemplate(rule.Details, templateData), 128, 2),
		DetailsUrl:        adjustLength(renderTemplate(rule.DetailsUrl, templateData), 256, 0),
		State:             adjustLength(renderTemplate(rule.State, templateData), 128, 2),
		StateUrl:          adjustLength(renderTemplate(rule.StateUrl, templateData), 256, 0),
		Assets: &discord.ActivityAssets{
			LargeImage: adjustLength(largeImage, 300, 0),
			LargeText:  adjustLength(renderTemplate(rule.LargeText, templateData), 128, 2),
			LargeUrl:   adjustLength(renderTemplate(rule.LargeUrl, templateData), 256, 0),
			SmallImage: adjustLength(smallImage, 300, 0),
			SmallText:  adjustLength(renderTemplate(rule.SmallText, templateData), 128, 2),
			SmallUrl:   adjustLength(renderTemplate(rule.SmallUrl, templateData), 256, 0),
		},
	}
	for _, button := range rule.Buttons {
		label := adjustLength(stripNonAscii(renderTemplate(button.Label, templateData)), 32, 2)
		url := adjustLength(renderTemplate(button.Url, templateData), 512, 0)
		if url == "" {
			continue
		}
		discordActivity.Buttons = append(discordActivity.Buttons, &discord.ActivityButton{Label: label, Url: url})
		if len(discordActivity.Buttons) == 2 {
			break
		}
	}
	progressMode := renderTemplate(rule.ProgressMode, templateData)
	switch progressMode {
	case "off", "elapsed", "remaining", "bar":
	default:
		logger.Error(nil, "Invalid progress mode %q, defaulting to %q", progressMode, "off")
		progressMode = "off"
	}
	if progressMode != "off" {
		now := time.Now().UnixMilli()
		discordActivity.Timestamps = new(discord.ActivityTimestamps)
		if progressMode == "bar" || progressMode == "elapsed" {
			discordActivity.Timestamps.StartMs = now - activity.ElapsedDurationMs
		}
		if progressMode == "bar" || progressMode == "remaining" {
			discordActivity.Timestamps.EndMs = now + activity.Item.DurationMs - activity.ElapsedDurationMs
		}
	}
	ipcTimeout := s.ipcTimeout
	if ctx.Err() != nil {
		ipcTimeout = 2500 * time.Millisecond
	}
	if err := s.discordService.SetActivity(discordActivity, ipcTimeout); err != nil {
		logger.Error(err, "Failed to set Discord activity")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if activity.State == "paused" && rule.PauseTimeoutSeconds > 0 {
		s.setStopTimer(time.Duration(rule.PauseTimeoutSeconds) * time.Second)
	} else {
		s.setStopTimer(s.idleTimeout)
	}
	s.lastState = activity.State
}

type pendingUpload struct {
	done chan struct{}
	url  string
}

// Can use singleflight instead of this custom implementation
var pendingUploads = map[string]*pendingUpload{}
var pendingUploadsMu sync.Mutex

func (s *Service) getUploadedImageUrl(ctx context.Context, thumb string, sourceUrl string, headers map[string]string) string { //nolint:contextcheck // Image upload has to continue in the background so don't inherit the context
	if s.imageService == nil {
		return ""
	}
	cacheKey := fmt.Sprintf("%s:%t:%d", thumb, s.imagesConfig.FitInSquare, s.imagesConfig.MaxSize)
	for attempt := 1; attempt <= maxUploadAttempts; attempt++ {
		if cached := s.cacheService.Get(cacheKey); cached != "" {
			return cached
		}
		pendingUploadsMu.Lock()
		result, ok := pendingUploads[cacheKey]
		if !ok {
			logger.Debug("Initiating upload for image %q (attempt %d)", thumb, attempt)
			result = &pendingUpload{done: make(chan struct{}), url: ""}
			pendingUploads[cacheKey] = result
			imgCtx, cancel := context.WithTimeout(context.Background(), time.Duration(s.imagesConfig.UploadTimeoutSeconds)*time.Second)
			go func() {
				defer cancel()
				defer func() {
					close(result.done)
					pendingUploadsMu.Lock()
					delete(pendingUploads, cacheKey)
					pendingUploadsMu.Unlock()
				}()
				pngBytes, err := images.GetPngBytes(imgCtx, sourceUrl, s.imagesConfig.FitInSquare, s.imagesConfig.MaxSize, headers)
				if err != nil {
					logger.Error(err, "Failed to get image %q for uploading", thumb)
					return
				}
				newUrl, err := s.imageService.Upload(imgCtx, pngBytes)
				if err != nil || newUrl == "" {
					logger.Error(err, "Failed to upload image %q", thumb)
					return
				}
				if err := s.cacheService.Set(cacheKey, newUrl, s.imageService.Lifespan()); err != nil {
					logger.Error(err, "Failed to add uploaded image %q URL to cache", thumb)
				}
				result.url = newUrl
			}()
		}
		pendingUploadsMu.Unlock()
		select {
		case <-ctx.Done():
			return ""
		case <-result.done:
			if result.url != "" {
				return result.url
			}
		}
	}
	return ""
}

func mapActivityType(mediaType string) discord.ActivityType {
	switch mediaType {
	case "track":
		return discord.ActivityTypeListening
	default:
		return discord.ActivityTypeWatching
	}
}

func mapStatusDisplayType(statusType string) discord.ActivityStatusDisplayType {
	switch statusType {
	case "state":
		return discord.ActivityStatusDisplayTypeState
	case "details":
		return discord.ActivityStatusDisplayTypeDetails
	default:
		return discord.ActivityStatusDisplayTypeName
	}
}

var templateCache sync.Map // map[string]*template.Template

func renderTemplate(tmpl string, data map[string]any) string {
	var t *template.Template
	if val, ok := templateCache.Load(tmpl); ok {
		t = val.(*template.Template) //nolint:errcheck,forcetypeassert
	} else {
		var err error
		t, err = template.New("").Funcs(template.FuncMap{
			"formatDuration": formatDuration,
			"formatGenres":   formatGenres,
			"toSentenceCase": toSentenceCase,
			"stripNonAscii":  stripNonAscii,
			"adjustLength":   adjustLength,
		}).Parse(tmpl)
		if err != nil {
			logger.Error(err, "Failed to parse template %q", tmpl)
			return tmpl
		}
		templateCache.Store(tmpl, t)
	}
	var buffer bytes.Buffer
	if err := t.Execute(&buffer, data); err != nil {
		logger.Error(err, "Failed to execute template %q", tmpl)
		return tmpl
	}
	return strings.ReplaceAll(buffer.String(), "<no value>", "")
}

func formatDuration(milliseconds int64, format string) string {
	duration := time.Duration(milliseconds) * time.Millisecond
	if format == "" {
		return duration.Round(time.Second).String()
	}
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	seconds := int(duration.Seconds()) % 60
	return fmt.Sprintf(format, hours, minutes, seconds)
}

func formatGenres(genres []plex.Genre, delimiter string, maxGenres int) string {
	var genreNames []string
	for _, genre := range genres {
		genreNames = append(genreNames, genre.Tag)
		if len(genreNames) == maxGenres {
			break
		}
	}
	return strings.Join(genreNames, delimiter)
}

func toSentenceCase(text string) string {
	if text == "" {
		return text
	}
	chars := []rune(text)
	chars[0] = []rune(strings.ToUpper(string(chars[0])))[0]
	return string(chars)
}

func adjustLength(text string, maxLength int, minLength int) string {
	if text == "" {
		return text
	}
	chars := []rune(text)
	if len(chars) > maxLength {
		marginLength := min(maxLength, 3)
		return string(chars[:maxLength-marginLength]) + strings.Repeat(".", marginLength)
	}
	if len(chars) < minLength {
		return text + strings.Repeat(" ", minLength-len(chars))
	}
	return text
}

var nonAsciiRegex = regexp.MustCompile(`[^\x00-\x7f]`)

func stripNonAscii(text string) string {
	return nonAsciiRegex.ReplaceAllString(text, "")
}
