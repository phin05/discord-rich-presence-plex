import { generatePlexPin, getPlexAuthToken, getPlexAuthUrl, getPlexResources } from "@/plex/client";
import { PlexAuthContext, type PlexAuthOnAdd, type PlexAuthOnCancel } from "@/plex/PlexAuthContext";
import { type PlexResource } from "@/plex/types";
import { Button, Divider, Flex, Kbd, Loader, Modal, Select, Switch, Text, TextInput } from "@mantine/core";
import { notifications } from "@mantine/notifications";
import { IconPlus } from "@tabler/icons-react";
import { type PropsWithChildren, useCallback, useRef, useState } from "react";

export function PlexAuthProvider({ children }: PropsWithChildren) {
	const [phase, setPhase] = useState<"idle" | "pin" | "popup" | "token" | "servers" | "server">("idle");
	const [token, setToken] = useState("");
	const [server, setServer] = useState("");
	const [servers, setServers] = useState<PlexResource[]>([]);
	const [manualEntry, setManualEntry] = useState(false);

	const onAddRef = useRef<PlexAuthOnAdd | null>(null);
	const onCancelRef = useRef<PlexAuthOnCancel | null>(null);

	const popupRef = useRef<Window | null>(null);
	const closedCheckerRef = useRef<number | null>(null);

	const cleanupPopup = useCallback(() => {
		if (popupRef.current !== null) {
			if (!popupRef.current.closed) {
				popupRef.current.close();
			}
			popupRef.current = null;
		}
		if (closedCheckerRef.current !== null) {
			clearInterval(closedCheckerRef.current);
			closedCheckerRef.current = null;
		}
	}, []);

	const cleanupState = useCallback(() => {
		setPhase("idle");
		setToken("");
		setServer("");
		setServers([]);
		setManualEntry(false);
	}, []);

	const cancel = useCallback(() => {
		cleanupPopup();
		cleanupState();
		if (onCancelRef.current) {
			onCancelRef.current();
		}
		notifications.show({
			color: "red",
			title: "Plex authentication cancelled",
			message: "",
		});
	}, [cleanupPopup, cleanupState]);

	const handleError = useCallback(
		(error: Error) => {
			cleanupPopup();
			cleanupState();
			notifications.show({
				color: "red",
				title: "Plex authentication failed",
				message: error.message,
			});
		},
		[cleanupPopup, cleanupState],
	);

	const auth = useCallback(
		async (onAdd: PlexAuthOnAdd, onCancel?: PlexAuthOnCancel) => {
			try {
				onAddRef.current = onAdd;
				onCancelRef.current = onCancel ?? null;

				const width = 850;
				const height = 700;
				const left = (window.innerWidth - width) / 2 + window.screenLeft;
				const top = (window.innerHeight - height) / 2 + window.screenTop;
				const popup = window.open(`${window.location.origin}/?plexAuthWaiting`, "_blank", `width=${width},height=${height},top=${top},left=${left}`);
				if (!popup) {
					throw new Error("Failed to open auth popup. Please allow popups for this site and try again.");
				}
				popupRef.current = popup;

				setPhase("pin");
				const pin = await generatePlexPin();
				const authUrl = getPlexAuthUrl(pin);

				setPhase("popup");
				popup.location.assign(authUrl);

				closedCheckerRef.current = window.setInterval(async () => {
					try {
						if (!popupRef.current || !popup.closed) {
							return;
						}
						cleanupPopup();
						setPhase("token");
						const token = await getPlexAuthToken(pin);
						if (!token) {
							cancel();
							return;
						}
						setToken(token);
						setPhase("servers");
						const servers = await getPlexResources(token);
						setServers(servers);
						setPhase("server");
					} catch (error) {
						handleError(error as Error);
					}
				}, 250);
			} catch (error) {
				handleError(error as Error);
			}
		},
		[cleanupPopup, cancel, handleError],
	);

	const confirmServer = useCallback(() => {
		let serverName: string;
		if (manualEntry) {
			serverName = server.trim();
		} else {
			serverName = servers.find((s) => s.clientIdentifier === server)?.name ?? "";
		}
		if (onAddRef.current) {
			onAddRef.current(token, serverName);
		}
		cleanupState();
		notifications.show({
			color: "green",
			title: "Plex authentication successful!",
			message: `Server "${serverName}" added`,
		});
	}, [server, manualEntry, servers, token, cleanupState]);

	const statusText = phase === "pin" ? "Generating PIN..." : phase === "popup" ? "Complete the sign-in process in the opened popup window. Alternatively, copy the popup window's URL, complete the sign-in process elsewhere, and close the popup window." : phase === "token" ? "Retrieving auth token..." : phase === "servers" ? "Fetching available servers..." : "";

	const modal = (
		<>
			{["pin", "popup", "token", "servers"].includes(phase) && <PlexAuthProgressModal onClose={cancel} showCancelButton={phase === "popup"} statusText={statusText} />}
			{phase === "server" && (
				<Modal autoFocus centered closeOnClickOutside={false} closeOnEscape={false} onClose={cancel} opened padding="lg" size="lg" trapFocus withCloseButton={false}>
					<Flex direction="column" gap="md">
						<Flex direction="column" gap={4}>
							<Text fw={500}>Add a Plex Media Server</Text>
							<Text c="dimmed" size="sm">
								Select a server from your account, or enter a server name manually.
							</Text>
						</Flex>
						<Divider />
						<Switch
							checked={manualEntry}
							label="Enter server name manually"
							onChange={(event) => {
								setManualEntry(event.currentTarget.checked);
								setServer("");
							}}
						/>
						{manualEntry ? (
							<>
								<Text c="dimmed" size="sm">
									The server name can be found/set in the Plex Web App under:
								</Text>
								<Kbd size="md">Settings &gt; Server Settings &gt; General &gt; Friendly Name</Kbd>
								<TextInput
									autoFocus
									data-autofocus
									label="Server Name / Friendly Name"
									onChange={(e) => {
										setServer(e.currentTarget.value);
									}}
									placeholder="e.g. XYZ-Desktop"
									value={server}
								/>
							</>
						) : (
							<Select
								autoFocus
								clearable
								data={servers.map((server) => ({ value: server.clientIdentifier, label: `${server.name} (${server.product} on ${server.device})` }))}
								data-autofocus
								label="Server"
								nothingFoundMessage="No servers found. Enter server name manually instead."
								onChange={(value) => {
									setServer(value ?? "");
								}}
								placeholder="Select a server"
								searchable
								value={server}
							/>
						)}
						<Button disabled={!server.trim()} leftSection={<IconPlus size={16} />} onClick={confirmServer}>
							Add
						</Button>
					</Flex>
				</Modal>
			)}
		</>
	);

	return (
		<PlexAuthContext.Provider value={auth}>
			{children}
			{modal}
		</PlexAuthContext.Provider>
	);
}

export function PlexAuthProgressModal({ onClose, statusText, showCancelButton = false }: { onClose?: () => void; statusText: string; showCancelButton?: boolean }) {
	const handleClose = onClose ?? (() => undefined);
	return (
		<Modal autoFocus centered closeOnClickOutside={false} closeOnEscape={false} onClose={handleClose} opened size="xs" trapFocus withCloseButton={false}>
			<Flex align="center" direction="column" gap="md" p="md" style={{ textAlign: "center" }}>
				<Loader type="bars" />
				<Text fw={500}>Plex authentication in progress...</Text>
				<Text c="dimmed" size="sm">
					{statusText}
				</Text>
				{showCancelButton && (
					<Button onClick={handleClose} variant="subtle">
						Cancel
					</Button>
				)}
			</Flex>
		</Modal>
	);
}
