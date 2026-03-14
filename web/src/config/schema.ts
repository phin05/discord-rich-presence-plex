import { type Fields, type ObjectSchema } from "@/common/schema";
import type { Config, DisplayRule } from "@/config/types";

export const configSchema: ObjectSchema<Config> = {
	type: "object",
	label: "Configuration",
	fields: {
		version: {
			type: "number",
			label: "Configuration Version",
			hide: true,
		},
		web: {
			type: "object",
			label: "Web Settings",
			fields: {
				launchOnStartup: {
					type: "boolean",
					label: "Launch web interface on app startup",
					defaultValue: true,
					hideDefaultValue: true,
					hide: CONTAINERISED_BUILD,
				},
				bindAddress: {
					type: "string",
					label: "Bind Address",
					defaultValue: "127.0.0.1",
				},
				bindPort: {
					type: "number",
					label: "Bind Port",
					defaultValue: 8040,
				},
				allowedNetworks: {
					type: "array",
					label: "Allowed Networks",
					description: "Networks (in CIDR notation) allowed to connect to the web interface/API. Localhost is always allowed.",
					itemSchema: {
						type: "string",
						label: "Network",
					},
				},
				trustedProxies: {
					type: "array",
					label: "Trusted Proxies",
					description: "Networks (in CIDR notation) of trusted proxies, used for identifying the client's real IP from the X-Forwarded-For header",
					itemSchema: {
						type: "string",
						label: "Network",
					},
				},
			},
		},
		logger: {
			type: "object",
			label: "Logger Settings",
			fields: {
				enableDebugOutput: {
					type: "boolean",
					label: "Enable Debug Output",
				},
			},
		},
		discord: {
			type: "object",
			label: "Discord Settings",
			fields: {
				clientId: {
					type: "string",
					label: "Client ID",
					description: "Discord application/client ID",
					defaultValue: "413407336082833418",
				},
				ipcPipeNumber: {
					type: "number",
					label: "IPC Pipe Number",
					description: "A number in the range of 0-9 to specify the Discord IPC pipe to connect to. Use -1 to auto-detect the first available pipe.",
					defaultValue: -1,
				},
				ipcTimeoutSeconds: {
					type: "number",
					label: "IPC Timeout",
					description: "Seconds to wait before giving up on sending an activity update to Discord",
					defaultValue: 10,
				},
				rateLimit: {
					type: "number",
					label: "Rate Limit",
					description: "Maximum number of activity update calls to Discord per 15 seconds",
					defaultValue: 5,
				},
				stopTimeoutSeconds: {
					type: "number",
					label: "Stop Timeout",
					description: "Seconds to wait before clearing Rich Presence after playback stops",
					defaultValue: 3,
				},
				idleTimeoutSeconds: {
					type: "number",
					label: "Idle Timeout",
					description: "Seconds to wait before clearing Rich Presence when no activity is detected (when Plex unexpectedly stops sending updates)",
					defaultValue: 25,
				},
				displayRules: {
					type: "object",
					label: "Display Rules",
					fields: {
						movie: {
							type: "object",
							label: "Movie Rules",
							fields: getDisplayRuleFields(),
						},
						episode: {
							type: "object",
							label: "Episode Rules",
							fields: getDisplayRuleFields(),
						},
						track: {
							type: "object",
							label: "Track Rules",
							fields: getDisplayRuleFields(),
						},
						clip: {
							type: "object",
							label: "Clip Rules",
							fields: getDisplayRuleFields(),
						},
						liveEpisode: {
							type: "object",
							label: "Live Episode Rules",
							fields: getDisplayRuleFields(),
						},
					},
				},
			},
		},
		images: {
			type: "object",
			label: "Image Settings",
			fields: {
				fitInSquare: {
					type: "boolean",
					label: "Fit In Square",
					description: "Fits images inside a square while maintaining aspect ratio (otherwise Discord crops them)",
					defaultValue: true,
				},
				maxSize: {
					type: "number",
					label: "Max Size",
					description: "Maximum width and height (in pixels) to use while downscaling images before uploading",
					defaultValue: 256,
				},
				uploadTimeoutSeconds: {
					type: "number",
					label: "Upload Timeout",
					description: "Seconds to wait before giving up on image uploads",
					defaultValue: 10,
				},
				uploaders: {
					type: "object",
					label: "Upload Providers",
					description: "The first enabled upload provider will be used. Images are disabled if no upload providers are enabled.",
					fields: {
						litterbox: {
							type: "object",
							label: "Litterbox",
							link: "https://litterbox.catbox.moe/",
							fields: {
								enabled: {
									type: "boolean",
									label: "Enabled",
									defaultValue: true,
									hideDefaultValue: true,
								},
								expiryHours: {
									type: "number",
									label: "Expiry Hours",
									description: "Hours until uploaded images expire (1, 12, 24, or 72)",
									defaultValue: 72,
								},
							},
						},
						imgBb: {
							type: "object",
							label: "ImgBB",
							link: "https://api.imgbb.com/",
							fields: {
								enabled: {
									type: "boolean",
									label: "Enabled",
								},
								apiKey: {
									type: "string",
									label: "API Key",
									description: "Get an API key at https://api.imgbb.com/",
									masked: true,
								},
								expiryMinutes: {
									type: "number",
									label: "Expiry Minutes",
									description: "Minutes until uploaded images expire. Set to 0 for no expiry.",
									defaultValue: 72 * 60,
								},
							},
						},
						imgur: {
							type: "object",
							label: "Imgur",
							link: "https://api.imgur.com/oauth2/addclient",
							fields: {
								enabled: {
									type: "boolean",
									label: "Enabled",
								},
								clientId: {
									type: "string",
									label: "Client ID",
									description: 'Get a client ID by registering an application at https://api.imgur.com/oauth2/addclient (pick "OAuth 2 authorization without a callback URL" as the authorisation type)',
									masked: true,
								},
							},
						},
						copyparty: {
							type: "object",
							label: "Copyparty",
							link: "https://github.com/9001/copyparty",
							fields: {
								enabled: {
									type: "boolean",
									label: "Enabled",
								},
								url: {
									type: "string",
									label: "URL",
									description: "URL of the directory to upload images to",
								},
								password: {
									type: "string",
									label: "Password",
									description: 'Password to be sent in the "PW" header',
									masked: true,
								},
								expiryMinutes: {
									type: "number",
									label: "Expiry Minutes",
									description: "Minutes until uploaded images expire. Has an effect only if the volume has a lifetime. Set to 0 for no expiry.",
									defaultValue: 72 * 60,
								},
							},
						},
					},
				},
			},
		},
		plex: {
			type: "object",
			label: "Plex Settings",
			fields: {
				users: {
					type: "array",
					label: "Users",
					itemSchema: {
						type: "object",
						label: "User",
						fields: {
							enabled: {
								type: "boolean",
								label: "Enabled",
								defaultValue: true,
								hideDefaultValue: true,
							},
							name: {
								type: "string",
								label: "Name",
								description: "Name for this user (for your reference)",
							},
							token: {
								type: "string",
								label: "Token",
								description: "Plex authentication token (X-Plex-Token)",
								masked: true,
							},
							servers: {
								type: "array",
								label: "Servers",
								itemSchema: {
									type: "object",
									label: "Server",
									fields: {
										enabled: {
											type: "boolean",
											label: "Enabled",
											defaultValue: true,
											hideDefaultValue: true,
										},
										name: {
											type: "string",
											label: "Name",
											description: "Friendly name of the server. Can be found/set in the Plex Web App under Settings > Server Settings > General > Friendly Name.",
										},
										url: {
											type: "string",
											label: "URL",
											description: "URL of the server. Auto-detected if left empty. Example: http://127.0.0.1:32400",
										},
										listenForUser: {
											type: "string",
											label: "Target Username",
											description: "If the current user is the server owner, only alerts from this username will be processed. The current user's username is used if this is empty.",
										},
										blacklistedLibraries: {
											type: "array",
											label: "Blacklisted Libraries",
											description: "Alerts from these libraries will be ignored",
											itemSchema: {
												type: "string",
												label: "Library Name",
											},
										},
										whitelistedLibraries: {
											type: "array",
											label: "Whitelisted Libraries",
											description: "If set, only alerts from these libraries will be processed",
											itemSchema: {
												type: "string",
												label: "Library Name",
											},
										},
										requestTimeoutSeconds: {
											type: "number",
											label: "Request Timeout",
											description: "Seconds to wait before giving up on requests to this server",
											defaultValue: 10,
										},
										retryIntervalSeconds: {
											type: "number",
											label: "Retry Interval",
											description: "Seconds to wait before retrying connection after a failure",
											defaultValue: 5,
										},
										maxRetriesBeforeExit: {
											type: "number",
											label: "Max Retries Before Exit",
											description: "The maximum number of connection retries before the application exits. Use -1 to retry indefinitely.",
											defaultValue: -1,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	},
};

function getDisplayRuleFields() {
	const fields: Fields<DisplayRule> = {
		details: {
			type: "string",
			label: "Details Text",
			template: true,
			description: "First line in Rich Presence",
		},
		state: {
			type: "string",
			label: "State Text",
			template: true,
			description: "Second line in Rich Presence",
		},
		statusType: {
			type: "autocomplete",
			label: "Status Type",
			template: true,
			description: 'Field to display in the status in the member list: "details" (details text), "state" (state text), or "name" (Plex). For instance, if set to "details" and the details text above resolves to "XYZ", the status displayed will be "Watching XYZ" or "Listening to XYZ".',
			options: ["details", "state", "name"],
		},
		largeImage: {
			type: "string",
			label: "Large Image",
			template: true,
			description: "URL or asset name for the large image",
		},
		largeText: {
			type: "string",
			label: "Large Text",
			template: true,
			description: 'Text to show when hovering over the large image. Also shown on the third line in Rich Presence for the "Listening" activity type (tracks).',
		},
		smallImage: {
			type: "string",
			label: "Small Image",
			template: true,
			description: "URL or asset name for the small image shown at the bottom-right corner of the large image",
		},
		smallText: {
			type: "string",
			label: "Small Text",
			template: true,
			description: "Text to show when hovering over the small image",
		},
		detailsUrl: {
			type: "string",
			label: "Details URL",
			template: true,
			description: "Link to open when the details text is clicked",
		},
		stateUrl: {
			type: "string",
			label: "State URL",
			template: true,
			description: "Link to open when the state text is clicked",
		},
		largeUrl: {
			type: "string",
			label: "Large URL",
			template: true,
			description: "Link to open when the large image is clicked",
		},
		smallUrl: {
			type: "string",
			label: "Small URL",
			template: true,
			description: "Link to open when the small image is clicked",
		},
		progressMode: {
			type: "autocomplete",
			label: "Progress Mode",
			template: true,
			description: 'Progress/timestamp display mode: "off" (disabled), "elapsed" (elapsed time), "remaining" (remaining time), or "bar" (progress bar). The "off" mode is currently broken due to a Discord bug/limitation.',
			options: ["off", "elapsed", "remaining", "bar"],
		},
		pauseTimeoutSeconds: {
			type: "number",
			label: "Pause Timeout",
			description: "Seconds to wait before clearing Rich Presence after playback is paused. Use -1 to show indefinitely and 0 to clear immediately. Progress/timestamp display while paused is currently broken due to a Discord bug/limitation.",
			defaultValue: 0,
		},
		buttons: {
			type: "array",
			label: "Buttons",
			description: "Discord can show up to 2 buttons. If a button's URL resolves to an empty string, the button is skipped. Buttons are visible to only other users and not yourself due to a Discord bug/limitation.",
			itemSchema: {
				type: "object",
				label: "Button",
				fields: {
					label: {
						type: "string",
						label: "Label",
						template: true,
						description: 'Examples: "{{ .ShowTitle }} on IMDb", "GitHub"',
					},
					url: {
						type: "string",
						label: "URL",
						template: true,
						description: 'Examples: "{{ .ShowImdbUrl }}", "https://github.com/"',
					},
				},
			},
		},
	};
	return fields;
}
