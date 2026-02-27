export type Fields = Record<string, Schema>;

export interface BaseSchema {
	label: string;
	description?: string;
	hideDefaultValue?: boolean;
}

export interface StringSchema extends BaseSchema {
	type: "string";
	masked?: boolean;
	template?: boolean;
	defaultValue?: string;
}

export interface NumberSchema extends BaseSchema {
	type: "number";
	defaultValue?: number;
}

export interface BooleanSchema extends BaseSchema {
	type: "boolean";
	defaultValue?: boolean;
}

export interface SelectSchema extends BaseSchema {
	type: "select";
	options: Array<{ label: string; value: string }>;
	defaultValue?: string;
}

export interface ObjectSchema extends BaseSchema {
	type: "object";
	fields: Fields;
	defaultValue?: Record<string, unknown>;
}

export interface ArraySchema extends BaseSchema {
	type: "array";
	itemSchema: Schema;
	defaultValue?: unknown[];
}

export type Schema = StringSchema | NumberSchema | BooleanSchema | SelectSchema | ObjectSchema | ArraySchema;

export function getDefaultValueForSchema(schema: Schema): unknown {
	if (schema.defaultValue !== undefined) {
		return schema.defaultValue;
	}
	if (schema.type === "object") {
		const obj: Record<string, unknown> = {};
		for (const [key, prop] of Object.entries(schema.fields)) {
			obj[key] = getDefaultValueForSchema(prop);
		}
		return obj;
	}
	if (schema.type === "array") {
		return [];
	}
	if (schema.type === "boolean") {
		return false;
	}
	if (schema.type === "number") {
		return 0;
	}
	return "";
}
