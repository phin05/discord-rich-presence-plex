import { type ArraySchema, getDefaultValueForSchema, type Schema } from "@/common/schema";
import { configSchema } from "@/config/schema";
import { useTemplateModal } from "@/config/TemplateModalContext";
import { type Config } from "@/config/types";
import { Accordion, ActionIcon, Box, Button, Flex, Input, NumberInput, PasswordInput, Select, Switch, Text, TextInput, Tooltip } from "@mantine/core";
import { IconBraces, IconPlus, IconTrash } from "@tabler/icons-react";
import { memo, type ReactNode } from "react";
import { type Control, Controller, type FieldArray, type FieldArrayPath, type FieldPath, useFieldArray } from "react-hook-form";

// eslint-disable-next-line prefer-arrow-callback
export const FormFields = memo(function FormFields({ control }: { control: Control<Config> }) {
	return (
		<Flex direction="column" gap="md" p="md" style={{ overflowY: "auto" }}>
			{Object.entries(configSchema.fields).map(([key, schema]) => (
				<FormField control={control} key={key} name={key as FieldPath<Config>} schema={schema} />
			))}
		</Flex>
	);
});

function getFieldDescription(schema: Schema) {
	const showDefault = schema.defaultValue !== undefined && !schema.hideDefaultValue;
	const defaultText = showDefault ? `[default: ${schema.type === "object" ? JSON.stringify(schema.defaultValue) : String(schema.defaultValue)}]` : "";
	const description = [schema.description, defaultText].filter(Boolean).join(" ");
	return description;
}

function FormField({ name, control, schema, label }: { name: FieldPath<Config>; control: Control<Config>; schema: Schema; label?: ReactNode }) {
	const openTemplateModal = useTemplateModal();

	const fieldLabel =
		schema.type === "string" && schema.template ? (
			<Flex align="center" gap={4}>
				{label ?? schema.label}
				<Tooltip label="Template String" withArrow>
					<ActionIcon
						onClick={() => {
							openTemplateModal(name);
						}}
						size="sm"
						variant="subtle"
					>
						<IconBraces size={16} />
					</ActionIcon>
				</Tooltip>
			</Flex>
		) : (
			(label ?? schema.label)
		);

	const fieldDescription = getFieldDescription(schema);

	if (schema.type === "object") {
		return (
			<Accordion defaultValue={name} variant="separated">
				<Accordion.Item value={name}>
					<Accordion.Control>
						{fieldLabel}
						{fieldDescription && (
							<Text c="dimmed" size="xs">
								{fieldDescription}
							</Text>
						)}
					</Accordion.Control>
					<Accordion.Panel>
						<Flex direction="column" gap="sm">
							{Object.entries(schema.fields).map(([key, propertySchema]) => (
								<FormField control={control} key={key} name={`${name}.${key}` as FieldPath<Config>} schema={propertySchema} />
							))}
						</Flex>
					</Accordion.Panel>
				</Accordion.Item>
			</Accordion>
		);
	}

	return (
		<Controller
			control={control}
			name={name}
			render={({ field, fieldState }) => {
				const error = fieldState.error?.message;

				if (schema.type === "array") {
					return <ArrayField control={control} error={error} label={fieldLabel} name={name as FieldArrayPath<Config>} schema={schema} />;
				}

				if (schema.type === "boolean") {
					return (
						<Switch
							checked={field.value as boolean}
							description={fieldDescription}
							error={error}
							label={fieldLabel}
							onChange={(event) => {
								field.onChange(event.currentTarget.checked);
							}}
						/>
					);
				}

				if (schema.type === "select") {
					return <Select allowDeselect={false} data={schema.options} description={fieldDescription} error={error} label={fieldLabel} onChange={field.onChange} value={field.value as string} />;
				}

				if (schema.type === "number") {
					return <NumberInput description={fieldDescription} error={error} label={fieldLabel} onChange={field.onChange} value={field.value as number} />;
				}

				const TextInputComponent = schema.masked ? PasswordInput : TextInput;
				return <TextInputComponent description={fieldDescription} error={error} label={fieldLabel} onChange={field.onChange} value={field.value as string} />;
			}}
		/>
	);
}

function ArrayField({ name, control, schema, label, error }: { name: FieldArrayPath<Config>; control: Control<Config>; schema: ArraySchema; label: ReactNode; error?: string }) {
	const { fields, append, remove } = useFieldArray({
		control,
		name,
	});

	// TODO: Add button to initiate the server selection part of the auth flow for adding additional servers to an existing user

	const fieldDescription = getFieldDescription(schema);

	return (
		<Flex bd={error ? "1px solid red" : undefined} direction="column" gap="sm">
			<Flex align="center" gap="md" justify="space-between">
				<Box>
					<Input.Label>{label}</Input.Label>
					<Input.Description>{fieldDescription}</Input.Description>
					<Input.Error>{error}</Input.Error>
				</Box>
				<Flex align="center" gap="md">
					{name === "plex.users" && (
						<Button
							leftSection={<IconPlus size={16} />}
							onClick={() => {
								document.getElementById("start-plex-auth")?.click();
							}}
							variant="outline"
						>
							Add (Plex Auth Flow)
						</Button>
					)}
					<Button
						leftSection={<IconPlus size={16} />}
						onClick={() => {
							append(getDefaultValueForSchema(schema.itemSchema) as FieldArray<Config>);
						}}
						variant="outline"
					>
						Add{name === "plex.users" ? " (Manual)" : ""}
					</Button>
				</Flex>
			</Flex>
			{fields.length > 0 && (
				<Flex direction="column" gap="sm" ml={schema.itemSchema.type !== "object" ? "md" : undefined}>
					{fields.map((field, index) => (
						<FormField
							control={control}
							key={field.id}
							label={
								<Flex align="center" gap="sm">
									{`${schema.itemSchema.label} ${index + 1}`}
									<ActionIcon
										color="red"
										component="div"
										onClick={() => {
											remove(index);
										}}
										size="sm"
										variant="outline"
									>
										<IconTrash size={16} />
									</ActionIcon>
								</Flex>
							}
							name={`${name}.${index}`}
							schema={schema.itemSchema}
						/>
					))}
				</Flex>
			)}
		</Flex>
	);
}
