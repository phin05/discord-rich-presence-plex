package mediator

import (
	"bytes"
	"drpp/server/logger"
	"drpp/server/plex"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"text/template"
	"time"
	"unicode"
)

func buildTemplateData(activity *plex.Activity) map[string]any {
	data := make(map[string]any)
	data["MediaType"] = activity.MediaType
	data["State"] = activity.State
	data["LibraryName"] = activity.Item.LibrarySectionTitle
	data["ElapsedDurationMs"] = activity.ElapsedDurationMs
	addGuidUrls := func(guids []plex.Guid, prefix string) {
		guidMap := make(map[string]string)
		for _, guid := range guids {
			parts := strings.SplitN(guid.Id, "://", 2)
			if len(parts) == 2 {
				guidMap[parts[0]] = parts[1]
			}
		}
		if id, ok := guidMap["imdb"]; ok {
			data[prefix+"ImdbUrl"] = "https://www.imdb.com/title/" + id
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
			data[prefix+"TmdbUrl"] = "https://www.themoviedb.org/" + tmdbPathSegment + "/" + id
			data[prefix+"TraktUrl"] = "https://trakt.tv/search/tmdb/" + id + "?id_type=" + traktIdType
			if activity.MediaType == "movie" {
				data[prefix+"LetterboxdUrl"] = "https://letterboxd.com/tmdb/" + id
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
			data[prefix+"TvdbUrl"] = "https://www.thetvdb.com/dereferrer/" + tvdbPathSegment + "/" + id
		}
		if id, ok := guidMap["mbid"]; ok {
			data[prefix+"MusicBrainzUrl"] = "https://musicbrainz.org/" + strings.ToLower(prefix) + "/" + id
		}
	}
	switch activity.MediaType {
	case "movie":
		data["Title"] = activity.Item.Title
		data["Year"] = activity.Item.Year
		data["Duration"] = formatDuration(activity.Item.DurationMs, "")
		data["Genres"] = formatGenres(activity.Item.Genres, ", ", 3)
		data["Poster"] = activity.Item.Thumb
		addGuidUrls(activity.Item.Guids, "")
	case "episode":
		data["ShowTitle"] = activity.GrandparentItem.Title
		data["ShowYear"] = activity.GrandparentItem.Year
		data["EpisodeDuration"] = formatDuration(activity.Item.DurationMs, "")
		data["ShowGenres"] = formatGenres(activity.GrandparentItem.Genres, ", ", 3)
		data["ShowPoster"] = activity.GrandparentItem.Thumb
		data["SeasonNumber"] = activity.ParentItem.Index
		data["EpisodeNumber"] = activity.Item.Index
		data["EpisodeTitle"] = activity.Item.Title
		addGuidUrls(activity.Item.Guids, "Episode")
		addGuidUrls(activity.GrandparentItem.Guids, "Show")
	case "track":
		data["Title"] = activity.Item.Title
		if activity.Item.OriginalTitle != "" {
			data["Artist"] = activity.Item.OriginalTitle
		} else {
			data["Artist"] = activity.GrandparentItem.Title
		}
		data["Album"] = activity.ParentItem.Title
		data["Year"] = activity.ParentItem.Year
		data["AlbumArtist"] = activity.GrandparentItem.Title
		data["AlbumPoster"] = activity.ParentItem.Thumb
		data["ArtistPoster"] = activity.GrandparentItem.Thumb
		data["Duration"] = formatDuration(activity.Item.DurationMs, "")
		data["AlbumGenres"] = formatGenres(activity.ParentItem.Genres, ", ", 3)
		data["ArtistGenres"] = formatGenres(activity.GrandparentItem.Genres, ", ", 3)
		addGuidUrls(activity.Item.Guids, "Track")
		addGuidUrls(activity.ParentItem.Guids, "Album")
		addGuidUrls(activity.GrandparentItem.Guids, "Artist")
	case "clip":
		data["Title"] = activity.Item.Title
		data["Duration"] = formatDuration(activity.Item.DurationMs, "")
		data["Poster"] = activity.Item.Thumb
	case "liveEpisode":
		data["ShowTitle"] = activity.Item.GrandparentTitle
		data["EpisodeTitle"] = activity.Item.Title
		data["ShowPoster"] = activity.Item.GrandparentThumb
	}
	data["Item"] = activity.Item
	data["ParentItem"] = activity.ParentItem
	data["GrandparentItem"] = activity.GrandparentItem
	return data
}

var templateCache sync.Map // map[string]*template.Template

var templateFuncs = template.FuncMap{
	"formatDuration": formatDuration,
	"formatGenres":   formatGenres,
	"toSentenceCase": toSentenceCase,
	"adjustLength":   adjustLength,
	"stripNonAscii":  stripNonAscii,
}

func renderTemplate(tmpl string, data map[string]any) string {
	var t *template.Template
	if val, ok := templateCache.Load(tmpl); ok {
		t = val.(*template.Template) //nolint:errcheck,forcetypeassert
	} else {
		var err error
		t, err = template.New("").Funcs(templateFuncs).Parse(tmpl)
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
	chars[0] = unicode.ToUpper(chars[0])
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
