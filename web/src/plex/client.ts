import { type PlexPin, type PlexResource } from "@/plex/types";

const productName = "Discord Rich Presence for Plex";
const clientId = "discord-rich-presence-plex";
const defaultHeaders = {
	"Accept": "application/json",
	"X-Plex-Client-Identifier": clientId,
};
const fetchTimeout = 10000;

export async function generatePlexPin() {
	const response = await fetch("https://plex.tv/api/v2/pins?strong=true", {
		method: "POST",
		signal: AbortSignal.timeout(fetchTimeout),
		headers: {
			...defaultHeaders,
			"X-Plex-Product": productName,
		},
	});
	if (!response.ok) {
		throw new Error(`Failed to generate Plex PIN: HTTP status ${response.status} ${response.statusText}`);
	}
	return (await response.json()) as PlexPin;
}

export function getPlexAuthUrl(pin: PlexPin) {
	const forwardUrl = `${window.location.origin}/?plexAuthCallback`;
	return `https://app.plex.tv/auth#?clientID=${clientId}&code=${pin.code}&forwardUrl=${encodeURIComponent(forwardUrl)}&context%5Bdevice%5D%5Bproduct%5D=${encodeURIComponent(productName)}`;
}

export async function getPlexAuthToken(pin: PlexPin) {
	const response = await fetch(`https://plex.tv/api/v2/pins/${pin.id}?code=${pin.code}`, {
		signal: AbortSignal.timeout(fetchTimeout),
		headers: defaultHeaders,
	});
	if (!response.ok) {
		throw new Error(`Failed to get Plex auth token: HTTP status ${response.status} ${response.statusText}`);
	}
	pin = (await response.json()) as PlexPin;
	return pin.authToken;
}

export async function getPlexResources(authToken: string) {
	const response = await fetch("https://clients.plex.tv/api/v2/resources?includeHttps=1&includeRelay=1&includeIPv6=1", {
		signal: AbortSignal.timeout(fetchTimeout),
		headers: {
			...defaultHeaders,
			"X-Plex-Token": authToken,
		},
	});
	if (!response.ok) {
		throw new Error(`Failed to get Plex resources: HTTP status ${response.status} ${response.statusText}`);
	}
	const data = (await response.json()) as PlexResource[];
	return data;
}
