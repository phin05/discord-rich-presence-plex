import { InfoModal } from "@/InfoModal";
import { apiRequest } from "@/common/api";
import { ConfigEditor } from "@/config/ConfigEditor";
import { LogStream } from "@/logs/LogStream";
import { ActionIcon, Button, Divider, Flex, Text, Title, Tooltip, useComputedColorScheme, useMantineColorScheme } from "@mantine/core";
import { useDisclosure } from "@mantine/hooks";
import { IconBrandGithub, IconExternalLink, IconInfoCircle, IconMoon, IconSun } from "@tabler/icons-react";
import { useQuery } from "@tanstack/react-query";

// TODO: Responsive design for smaller screens

export function App() {
	return (
		<Flex direction="column" h="100vh">
			<Header />
			<Divider />
			<Flex flex="1" style={{ overflowY: "hidden" }}>
				<ConfigEditor />
				<Divider orientation="vertical" />
				<LogStream />
			</Flex>
		</Flex>
	);
}

async function getLatestRelease() {
	return await apiRequest<{ tag_name: string }>("GET", "https://api.github.com/repos/phin05/discord-rich-presence-plex/releases/latest");
}

function Header() {
	const { setColorScheme } = useMantineColorScheme();
	const currentColorScheme = useComputedColorScheme();
	const [infoOpened, { open: openInfo, close: closeInfo }] = useDisclosure(false);

	const latestRelease = useQuery({
		queryKey: ["latestRelease"],
		meta: { errorTitle: "Failed to fetch latest release info from GitHub" },
		queryFn: getLatestRelease,
		enabled: false,
	});

	return (
		<Flex align="center" gap="md" justify="space-between" p="md">
			<Flex align="center" gap="md">
				<Title order={3}>Discord Rich Presence for Plex (DRPP)</Title>
				<>
					<Divider orientation="vertical" />
					<Text>
						Version: v{APP_VERSION}
						{latestRelease.data && ` (Latest: ${latestRelease.data.tag_name})`}
					</Text>
					{latestRelease.data ? (
						`v${APP_VERSION}` !== latestRelease.data.tag_name && (
							<Button color="blue" component="a" href="https://github.com/phin05/discord-rich-presence-plex/releases/latest" rel="noopener noreferrer" rightSection={<IconExternalLink size={16} />} target="_blank" variant="light">
								Update Available
							</Button>
						)
					) : (
						<Button
							color="blue"
							loading={latestRelease.isFetching}
							onClick={() => {
								void latestRelease.refetch();
							}}
							variant="light"
						>
							Check for Updates
						</Button>
					)}
				</>
			</Flex>
			<Flex gap="md">
				<Tooltip label={`${currentColorScheme === "light" ? "Dark" : "Light"} Mode`} position="bottom">
					<ActionIcon
						onClick={() => {
							setColorScheme(currentColorScheme === "light" ? "dark" : "light");
						}}
						size="xl"
						variant="default"
					>
						{currentColorScheme === "light" ? <IconMoon size={24} /> : <IconSun size={24} />}
					</ActionIcon>
				</Tooltip>
				<Tooltip label="GitHub" position="bottom">
					<ActionIcon component="a" href="https://github.com/phin05/discord-rich-presence-plex" rel="noopener noreferrer" size="xl" target="_blank" variant="default">
						<IconBrandGithub size={24} />
					</ActionIcon>
				</Tooltip>
				<Tooltip label="Info" position="bottom">
					<ActionIcon onClick={openInfo} size="xl" variant="default">
						<IconInfoCircle size={24} />
					</ActionIcon>
				</Tooltip>
				<InfoModal onClose={closeInfo} opened={infoOpened} />
			</Flex>
		</Flex>
	);
}
