import { apiBaseUrl } from "@/common/api";
import { Code, Modal, Tabs, Text } from "@mantine/core";
import { useQuery } from "@tanstack/react-query";

const files = [
	{ label: "OSS Attribution", filename: "NOTICE.md" },
	{ label: "Readme", filename: "README.md" },
	{ label: "License", filename: "LICENSE" },
	{ label: "Copyright", filename: "COPYRIGHT" },
];

export function InfoModal({ opened, onClose }: { opened: boolean; onClose: () => void }) {
	return (
		<Modal onClose={onClose} opened={opened} size="75%" title={<Text fw={500}>Info</Text>}>
			<Tabs defaultValue={files[0].filename}>
				<Tabs.List>
					{files.map((file) => (
						<Tabs.Tab key={file.filename} value={file.filename}>
							{file.label}
						</Tabs.Tab>
					))}
				</Tabs.List>
				{files.map((file) => (
					<Tabs.Panel key={file.filename} value={file.filename}>
						<FileContents filename={file.filename} />
					</Tabs.Panel>
				))}
			</Tabs>
		</Modal>
	);
}

async function fetchFile(filename: string) {
	const response = await fetch(`${apiBaseUrl}/static/${filename}`);
	return response.text();
}

function FileContents({ filename }: { filename: string }) {
	const { data, isLoading, isError, error } = useQuery({
		queryKey: ["static", filename],
		queryFn: () => fetchFile(filename),
		staleTime: Infinity,
	});

	if (isLoading) {
		return <Code block>Loading...</Code>;
	}

	if (isError) {
		return <Code block>Error: {error.message}</Code>;
	}

	return (
		<Code block style={{ whiteSpace: "pre-wrap", wordBreak: "break-word" }}>
			{data}
		</Code>
	);
}
