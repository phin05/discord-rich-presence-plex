import { commonTemplateVariables, templateFunctions, type TemplateVariable, templateVariableGroups } from "@/config/template";
import { TemplateModalContext } from "@/config/TemplateModalContext";
import { ActionIcon, Anchor, Code, CopyButton, Flex, Modal, Table, Tabs, Text, Tooltip } from "@mantine/core";
import { useDisclosure } from "@mantine/hooks";
import { IconBraces, IconCheck, IconCopy } from "@tabler/icons-react";
import { type PropsWithChildren, useCallback, useState } from "react";

export function TemplateModalProvider({ children }: PropsWithChildren) {
	const [opened, { open, close }] = useDisclosure(false);
	const [mediaType, setMediaType] = useState("");

	const openFunc = useCallback(
		(fieldName: string) => {
			const fnSegments = fieldName.split(".");
			setMediaType(fnSegments.length >= 3 ? fnSegments[2] : templateVariableGroups[0].mediaType);
			open();
		},
		[open],
	);

	return (
		<TemplateModalContext.Provider value={openFunc}>
			{children}
			<TemplateModal close={close} mediaType={mediaType} opened={opened} />
		</TemplateModalContext.Provider>
	);
}

function TemplateModal({ opened, close, mediaType }: { opened: boolean; close: () => void; mediaType: string }) {
	return (
		<Modal onClose={close} opened={opened} size="xl" title={<Text fw={500}>Template Variables and Functions</Text>}>
			<Text c="dimmed" size="sm">
				Templating System:{" "}
				<Anchor href="https://pkg.go.dev/text/template" target="_blank" underline="never">
					https://pkg.go.dev/text/template
				</Anchor>
			</Text>
			<Text c="dimmed" size="sm">
				The below template variables and functions can be used in fields marked with the <IconBraces color="#228be6" size={16} style={{ verticalAlign: "-3px" }} /> icon.
			</Text>
			<Tabs defaultValue={mediaType} mt="md">
				<Tabs.List>
					{templateVariableGroups.map((group) => (
						<Tabs.Tab key={group.mediaType} value={group.mediaType}>
							{group.label}
						</Tabs.Tab>
					))}
					<Tabs.Tab value="functions">Functions</Tabs.Tab>
				</Tabs.List>
				{templateVariableGroups.map((group) => (
					<Tabs.Panel key={group.mediaType} value={group.mediaType}>
						<VariablesTable variables={group.variables} />
						<VariablesTable variables={commonTemplateVariables} />
					</Tabs.Panel>
				))}
				<Tabs.Panel mt="md" value="functions">
					<Table withColumnBorders withTableBorder>
						<Table.Thead>
							<Table.Tr>
								<Table.Th>Name / Signature</Table.Th>
								<Table.Th>Example Usage</Table.Th>
								<Table.Th>Description</Table.Th>
							</Table.Tr>
						</Table.Thead>
						<Table.Tbody>
							{templateFunctions.map((func) => (
								<Table.Tr key={func.name}>
									<Table.Td>
										<Code style={{ display: "block" }}>{func.signature}</Code>
									</Table.Td>
									<Table.Td>
										<Code style={{ display: "block" }}>{func.example}</Code>
									</Table.Td>
									<Table.Td>{func.description}</Table.Td>
								</Table.Tr>
							))}
						</Table.Tbody>
					</Table>
				</Tabs.Panel>
			</Tabs>
		</Modal>
	);
}

function VariablesTable({ variables }: { variables: TemplateVariable[] }) {
	return (
		<Table mt="md" withColumnBorders withTableBorder>
			<Table.Thead>
				<Table.Tr>
					<Table.Th>Variable</Table.Th>
					<Table.Th>Description</Table.Th>
				</Table.Tr>
			</Table.Thead>
			<Table.Tbody>
				{variables.map((variable) => {
					const template = `{{ .${variable.name} }}`;
					return (
						<Table.Tr key={variable.name}>
							<Table.Td>
								<Flex align="center" gap="xs">
									<Code>{template}</Code>
									<CopyButton value={template}>
										{({ copied, copy }) => (
											<Tooltip label={copied ? "Copied!" : "Copy"} withArrow>
												<ActionIcon color={copied ? "green" : "gray"} onClick={copy} size="xs" variant="subtle">
													{copied ? <IconCheck size={16} /> : <IconCopy size={16} />}
												</ActionIcon>
											</Tooltip>
										)}
									</CopyButton>
								</Flex>
							</Table.Td>
							<Table.Td>{variable.description}</Table.Td>
						</Table.Tr>
					);
				})}
			</Table.Tbody>
		</Table>
	);
}
