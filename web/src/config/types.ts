export interface Config {
	web: Web;
	logger: Logger;
	discord: Discord;
	images: Images;
	plex: Plex;
}

export interface Web {
	launchOnStartup: boolean;
	bindAddress: string;
	bindPort: number;
	allowedNetworks: string[];
	trustedProxies: string[];
}

export interface Logger {
	enableDebugOutput: boolean;
}

export interface Discord {
	clientId: string;
	ipcPipeNumber: number;
	ipcTimeoutSeconds: number;
	rateLimit: number;
	stopTimeoutSeconds: number;
	idleTimeoutSeconds: number;
	displayRules: DisplayRules;
}

export interface DisplayRules {
	movie: DisplayRule;
	episode: DisplayRule;
	track: DisplayRule;
	clip: DisplayRule;
	liveEpisode: DisplayRule;
}

export interface DisplayRule {
	details: string;
	state: string;
	statusType: string;
	largeImage: string;
	largeText: string;
	smallImage: string;
	smallText: string;
	detailsUrl: string;
	stateUrl: string;
	largeUrl: string;
	smallUrl: string;
	progressMode: string;
	pauseTimeoutSeconds: number;
	buttons: Button[];
}

export interface Button {
	label: string;
	url: string;
}

export interface Images {
	fitInSquare: boolean;
	maxSize: number;
	uploadTimeoutSeconds: number;
	uploaders: Uploaders;
}

export interface Uploaders {
	litterbox: Litterbox;
	imgBb: ImgBb;
	imgur: Imgur;
	copyparty: Copyparty;
}

export interface Litterbox {
	enabled: boolean;
	expiryHours: 1 | 12 | 24 | 72;
}

export interface ImgBb {
	enabled: boolean;
	apiKey: string;
	expiryMinutes: number;
}

export interface Imgur {
	enabled: boolean;
	clientId: string;
}

export interface Copyparty {
	enabled: boolean;
	url: string;
	password: string;
	expiryMinutes: number;
}

export interface Plex {
	users: User[];
}

export interface User {
	enabled: boolean;
	name: string;
	token: string;
	servers: Server[];
}

export interface Server {
	enabled: boolean;
	name: string;
	listenForUser: string;
	blacklistedLibraries: string[];
	whitelistedLibraries: string[];
	requestTimeoutSeconds: number;
	retryIntervalSeconds: number;
	maxRetriesBeforeExit: number;
}
