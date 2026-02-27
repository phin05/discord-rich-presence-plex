export interface PlexPin {
	id: number;
	code: string;
	authToken: string;
}

export interface PlexResource {
	name: string;
	product: string;
	productVersion: string;
	platform: string;
	platformVersion: string;
	device: string;
	clientIdentifier: string;
	provides: string;
	createdAt: string;
	lastSeenAt: string;
	owned: boolean;
}
