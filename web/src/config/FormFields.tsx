import { type ArraySchema, eachField, getDefaultValueForSchema, type Schema } from "@/common/schema";
import { configSchema } from "@/config/schema";
import { useTemplateModal } from "@/config/TemplateModalContext";
import { type Config } from "@/config/types";
import { Accordion, ActionIcon, Autocomplete, Box, Button, Flex, Input, NumberInput, PasswordInput, Switch, Text, TextInput, Tooltip } from "@mantine/core";
import { IconBraces, IconExternalLink, IconPlus, IconTrash } from "@tabler/icons-react";
import { memo, type ReactNode } from "react";
import { type Control, Controller, type FieldArray, type FieldArrayPath, type FieldPath, type FieldValues, useFieldArray } from "react-hook-form";

// eslint-disable-next-line prefer-arrow-callback
export const FormFields = memo(function FormFields({ control }: { control: Control<Config> }) {
	return (
		<Flex direction="column" gap="md" p="md" style={{ overflowY: "auto" }}>
			{eachField(configSchema.fields).map(([name, schema]) => (
				<FormField control={control} key={name} name={name} schema={schema} />
			))}
		</Flex>
	);
});

function FormField<T extends FieldValues, S>({ name, control, schema, label }: { name: FieldPath<T>; control: Control<T>; schema: Schema<S>; label?: ReactNode }) {
	const openTemplateModal = useTemplateModal();

	const fieldLabel = (
		<Flex align="center" gap={4}>
			{label ?? schema.label}
			{(schema.type === "string" || schema.type === "autocomplete") && schema.template && (
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
			)}
			{schema.link && (
				<Tooltip label={schema.link} withArrow>
					<ActionIcon
						component="a"
						href={schema.link}
						onClick={(e) => {
							e.stopPropagation();
						}}
						rel="noopener noreferrer"
						size="sm"
						target="_blank"
						variant="subtle"
					>
						<IconExternalLink size={16} />
					</ActionIcon>
				</Tooltip>
			)}
		</Flex>
	);

	const showDefault = schema.defaultValue !== undefined && !schema.hideDefaultValue;
	const defaultText = showDefault ? `[default: ${schema.type === "array" || schema.type === "object" ? JSON.stringify(schema.defaultValue) : String(schema.defaultValue)}]` : "";
	const fieldDescription = [schema.description, defaultText].filter(Boolean).join(" ");

	if (schema.type === "object") {
		return (
			<Accordion defaultValue="item" variant="separated">
				<Accordion.Item value="item">
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
							{eachField(schema.fields).map(([fieldName, fieldSchema]) => (
								<FormField control={control} key={fieldName} name={`${name}.${fieldName}` as FieldPath<T>} schema={fieldSchema} />
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
					return <ArrayField control={control} description={fieldDescription} error={error} label={fieldLabel} name={name as FieldArrayPath<T>} schema={schema} />;
				}

				if (schema.type === "boolean") {
					return (
						<Switch
							checked={field.value}
							description={fieldDescription}
							error={error}
							label={fieldLabel}
							onChange={(event) => {
								field.onChange(event.currentTarget.checked);
							}}
						/>
					);
				}

				if (schema.type === "autocomplete") {
					return <Autocomplete data={schema.options} description={fieldDescription} error={error} filter={({ options }) => options} label={fieldLabel} onChange={field.onChange} value={field.value} />;
				}

				if (schema.type === "number") {
					return <NumberInput description={fieldDescription} error={error} label={fieldLabel} onChange={field.onChange} value={field.value} />;
				}

				const TextInputComponent = schema.masked ? PasswordInput : TextInput;
				return <TextInputComponent description={fieldDescription} error={error} label={fieldLabel} onChange={field.onChange} value={field.value} />;
			}}
		/>
	);
}

function ArrayField<T extends FieldValues, S extends unknown[]>({ name, control, schema, label, error, description }: { name: FieldArrayPath<T>; control: Control<T>; schema: ArraySchema<S>; label: ReactNode; error?: string; description?: string }) {
	const { fields, append, remove } = useFieldArray({
		control,
		name,
	});

	// TODO: Add button to initiate the server selection part of the auth flow for adding additional servers to an existing user

	return (
		<Flex bd={error ? "1px solid red" : undefined} direction="column" gap="sm">
			<Flex align="center" gap="md" justify="space-between">
				<Box>
					<Input.Label>{label}</Input.Label>
					<Input.Description>{description}</Input.Description>
					{error && <Input.Error mt="xs">{error}</Input.Error>}
				</Box>
				<Flex align="center" gap="md">
					{name === "plex.users" && (
						<Button
							leftSection={<IconPlus size={16} />}
							onClick={() => {
								document.getElementById("start-plex-auth")?.click(); // TODO: This is a bit hacky
							}}
							variant="outline"
						>
							Add (Plex Auth Flow)
						</Button>
					)}
					<Button
						leftSection={<IconPlus size={16} />}
						onClick={() => {
							append(getDefaultValueForSchema(schema.itemSchema) as FieldArray<T>);
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
										onClick={(e) => {
											e.stopPropagation();
											remove(index);
										}}
										size="sm"
										variant="outline"
									>
										<IconTrash size={16} />
									</ActionIcon>
								</Flex>
							}
							name={`${name}.${index}` as FieldPath<T>}
							schema={schema.itemSchema}
						/>
					))}
				</Flex>
			)}
		</Flex>
	);
}
