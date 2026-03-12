import { Code } from "@mantine/core";
import type { ReactNode } from "react";

export interface TemplateVariable {
	name: string;
	description: ReactNode;
}

export interface TemplateVariableGroup {
	mediaType: string;
	label: string;
	variables: TemplateVariable[];
}

export const commonTemplateVariables: TemplateVariable[] = [
	{
		name: "MediaType",
		description: (
			<span>
				Media type (<Code>movie</Code>, <Code>episode</Code>, <Code>track</Code>, <Code>clip</Code>, <Code>liveEpisode</Code>)
			</span>
		),
	},
	{
		name: "State",
		description: (
			<span>
				Playback state (<Code>playing</Code>, <Code>paused</Code>, <Code>stopped</Code>)
			</span>
		),
	},
	{ name: "LibraryName", description: "Library name" },
	{ name: "ElapsedDurationMs", description: "Elapsed playback time in milliseconds" },
	{ name: "Item", description: "Raw item object from Plex" },
	{ name: "ParentItem", description: "Raw parent item object from Plex" },
	{ name: "GrandparentItem", description: "Raw grandparent item object from Plex" },
];

export const templateVariableGroups: TemplateVariableGroup[] = [
	{
		mediaType: "movie",
		label: "Movie",
		variables: [
			{ name: "Title", description: "Title" },
			{ name: "Year", description: "Release year" },
			{ name: "Duration", description: "Duration (e.g. 2h30m)" },
			{ name: "Genres", description: "Genres (up to 3, comma-separated)" },
			{ name: "Poster", description: "Poster image URL" },
			{ name: "ImdbUrl", description: "IMDb URL" },
			{ name: "TmdbUrl", description: "TMDB URL" },
			{ name: "TraktUrl", description: "Trakt URL" },
			{ name: "LetterboxdUrl", description: "Letterboxd URL" },
			{ name: "TvdbUrl", description: "TheTVDB URL" },
		],
	},
	{
		mediaType: "episode",
		label: "Episode",
		variables: [
			{ name: "ShowTitle", description: "Show title" },
			{ name: "ShowYear", description: "Show release year" },
			{ name: "EpisodeDuration", description: "Episode duration (e.g. 45m)" },
			{ name: "ShowGenres", description: "Show genres (up to 3, comma-separated)" },
			{ name: "ShowPoster", description: "Show poster image URL" },
			{ name: "SeasonNumber", description: "Season number" },
			{ name: "EpisodeNumber", description: "Episode number" },
			{ name: "EpisodeTitle", description: "Episode title" },
			{ name: "EpisodeImdbUrl", description: "Episode IMDb URL" },
			{ name: "EpisodeTmdbUrl", description: "Episode TMDB URL" },
			{ name: "EpisodeTraktUrl", description: "Episode Trakt URL" },
			{ name: "EpisodeTvdbUrl", description: "Episode TheTVDB URL" },
			{ name: "ShowImdbUrl", description: "Show IMDb URL" },
			{ name: "ShowTmdbUrl", description: "Show TMDB URL" },
			{ name: "ShowTraktUrl", description: "Show Trakt URL" },
			{ name: "ShowTvdbUrl", description: "Show TheTVDB URL" },
		],
	},
	{
		mediaType: "track",
		label: "Track",
		variables: [
			{ name: "Title", description: "Track title" },
			{ name: "Artist", description: "Track artist" },
			{ name: "Album", description: "Album title" },
			{ name: "Year", description: "Album release year" },
			{ name: "AlbumArtist", description: "Album artist" },
			{ name: "AlbumPoster", description: "Album cover image URL" },
			{ name: "ArtistPoster", description: "Artist image URL" },
			{ name: "Duration", description: "Track duration (e.g. 3m30s)" },
			{ name: "AlbumGenres", description: "Album genres (up to 3, comma-separated)" },
			{ name: "ArtistGenres", description: "Artist genres (up to 3, comma-separated)" },
			{ name: "TrackMusicBrainzUrl", description: "Track MusicBrainz URL" },
			{ name: "AlbumMusicBrainzUrl", description: "Album MusicBrainz URL" },
			{ name: "ArtistMusicBrainzUrl", description: "Artist MusicBrainz URL" },
		],
	},
	{
		mediaType: "clip",
		label: "Clip",
		variables: [
			{ name: "Title", description: "Title" },
			{ name: "Duration", description: "Duration (e.g. 1m30s)" },
			{ name: "Poster", description: "Poster image URL" },
		],
	},
	{
		mediaType: "liveEpisode",
		label: "Live Episode",
		variables: [
			{ name: "ShowTitle", description: "Show title" },
			{ name: "EpisodeTitle", description: "Episode title" },
			{ name: "ShowPoster", description: "Show poster image URL" },
		],
	},
];

export interface TemplateFunction {
	name: string;
	signature: string;
	example: string;
	description: ReactNode;
}

export const templateFunctions: TemplateFunction[] = [
	{
		name: "formatDuration",
		signature: "formatDuration(milliseconds int64, format string)",
		example: '{{ formatDuration .ElapsedDurationMs "%d:%02d:%02d" }}',
		description: (
			<span>
				Formats a duration given in milliseconds. <Code>format</Code> should be a <Code>fmt.Sprintf</Code> format string receiving hours, minutes and seconds as integer arguments (e.g. <Code>&quot;%d:%02d:%02d&quot;</Code> → <Code>&quot;1:30:00&quot;</Code>). If <Code>format</Code> is an empty string, Go&apos;s default duration format is used (e.g. <Code>&quot;1h30m0s&quot;</Code>).
			</span>
		),
	},
	{
		name: "formatGenres",
		signature: "formatGenres(genres []plex.Genre, delimiter string, maxGenres int)",
		example: '{{ formatGenres .Item.Genres ", " 3 }}',
		description: (
			<span>
				Joins up to <Code>maxGenres</Code> genre names from the <Code>genres</Code> slice using the <Code>delimiter</Code>.
			</span>
		),
	},
	{
		name: "toSentenceCase",
		signature: "toSentenceCase(text string)",
		example: "{{ toSentenceCase .State }}",
		description: "Capitalises the first character of the given string.",
	},
	{
		name: "adjustLength",
		signature: "adjustLength(text string, maxLength int, ellipsis string, minLength int, padding string)",
		example: '{{ adjustLength .Title 32 "…" 2 " " }}',
		description: (
			<span>
				Truncates the given string to <Code>maxLength</Code> characters with an <Code>ellipsis</Code>, or right-pads with <Code>padding</Code> if shorter than <Code>minLength</Code>.
			</span>
		),
	},
	{
		name: "stripNonAscii",
		signature: "stripNonAscii(text string)",
		example: "{{ stripNonAscii .Title }}",
		description: "Removes all non-ASCII characters from the given string.",
	},
];
