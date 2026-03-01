import { apiBaseUrl } from "@/common/api";
import { Box, Divider, Flex, Indicator, Skeleton, Switch, Text, TextInput, Title, Tooltip } from "@mantine/core";
import { useLocalStorage } from "@mantine/hooks";
import { memo, useEffect, useMemo, useRef, useState } from "react";
import logEntryStyles from "./LogEntry.module.css";

const maxEntries = 1000;

interface Entry {
	id: number;
	timestamp: string;
	level: string;
	source: string;
	message: string;
}

export function LogStream() {
	const [entries, setEntries] = useState<Entry[]>([]);
	const [connected, setConnected] = useState(false);
	const [connectedOnce, setConnectedOnce] = useState(false);
	const [autoScroll, setAutoScroll] = useLocalStorage({ key: "logs-auto-scroll", defaultValue: true });
	const [wrapText, setWrapText] = useLocalStorage({ key: "logs-wrap-text", defaultValue: false });
	const [search, setSearch] = useState("");
	const [searchError, setSearchError] = useState("");
	const [searchRegExp, setSearchRegExp] = useState<RegExp | null>(null);
	const logsRef = useRef<HTMLDivElement>(null);

	const filteredEntries = useMemo(() => (searchRegExp ? entries.filter((entry) => searchRegExp.test(`${entry.timestamp} [${entry.level}] [${entry.source}] ${entry.message}`)) : entries), [searchRegExp, entries]);

	useEffect(() => {
		const eventSource = new EventSource(`${apiBaseUrl}/api/logs`);
		eventSource.onopen = () => {
			setConnected(true);
			setConnectedOnce(true);
			setEntries([]);
		};
		eventSource.onerror = () => {
			setConnected(false);
		};
		eventSource.onmessage = (event) => {
			const logEntry = JSON.parse(event.data as string) as Entry;
			setEntries((prevLogs) => [...prevLogs, logEntry].slice(-maxEntries));
		};
		return () => {
			eventSource.close();
		};
	}, []);

	useEffect(() => {
		if (autoScroll && logsRef.current) {
			logsRef.current.scrollTo({
				top: logsRef.current.scrollHeight,
				behavior: "instant",
			});
		}
	}, [autoScroll, filteredEntries, wrapText]);

	if (!connectedOnce) {
		return <Skeleton radius={0} w="50%" />;
	}

	return (
		<Flex direction="column" w="50%">
			<Flex align="center" gap="md" p="md">
				<Tooltip arrowSize={8} color={connected ? "green" : "red"} label={connected ? "Log stream connected" : "Log stream disconnected"} offset={16} position="top-start" withArrow>
					<Indicator color={connected ? "green" : "red"} ml="xs" processing={connected} />
				</Tooltip>
				<Title order={4}>Logs</Title>
				<Divider orientation="vertical" />
				<Switch
					checked={autoScroll}
					label="Auto-Scroll"
					onChange={(event) => {
						setAutoScroll(event.currentTarget.checked);
					}}
				/>
				<Switch
					checked={wrapText}
					label="Wrap Text"
					onChange={(event) => {
						setWrapText(event.currentTarget.checked);
					}}
				/>
				<Divider orientation="vertical" />
				<TextInput
					flex={1}
					onChange={(event) => {
						const search = event.currentTarget.value.trim();
						setSearch(search);
						setSearchError("");
						setSearchRegExp(null);
						if (!search) {
							return;
						}
						try {
							let regExp: RegExp;
							const lastSlashIndex = search.lastIndexOf("/");
							if (search.startsWith("/") && lastSlashIndex > 0) {
								const pattern = search.slice(1, lastSlashIndex);
								const flags = search.slice(lastSlashIndex + 1);
								regExp = new RegExp(pattern, flags);
							} else {
								regExp = new RegExp(search);
							}
							setSearchRegExp(regExp);
						} catch (error) {
							setSearchError((error as Error).message);
						}
					}}
					placeholder="Search logs (regex)"
					value={search}
				/>
				<Text c="dimmed" size="sm">
					{`${filteredEntries.length} log ${filteredEntries.length === 1 ? "entry" : "entries"}${entries.length !== filteredEntries.length ? ` (${entries.length} total)` : ""}`}
				</Text>
			</Flex>
			<Divider />
			{searchError && (
				<Text c="red" p="md" size="sm">
					[Search] {searchError}
				</Text>
			)}
			<Box ref={logsRef} style={{ overflow: "auto" }}>
				<Flex direction="column" style={{ minWidth: "fit-content", whiteSpace: wrapText ? "pre-wrap" : "pre" }}>
					{filteredEntries.map((entry) => (
						<LogEntry entry={entry} key={entry.id} />
					))}
				</Flex>
			</Box>
		</Flex>
	);
}

// eslint-disable-next-line prefer-arrow-callback
const LogEntry = memo(function LogEntry({ entry }: { entry: Entry }) {
	return (
		<div className={`${logEntryStyles.line} ${logEntryStyles[`line${entry.level}`]}`}>
			<span className={logEntryStyles.timestamp}>{entry.timestamp}</span> <span className={logEntryStyles[`level${entry.level}`]}>[{entry.level}]</span> <span className={logEntryStyles.source}>[{entry.source}]</span> {entry.message}
		</div>
	);
});
