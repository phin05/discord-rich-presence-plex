export interface BaseSchema<T> {
	label: string;
	description?: string;
	link?: string;
	defaultValue?: T;
	hideDefaultValue?: boolean;
}

export interface StringSchema<T extends string> extends BaseSchema<T> {
	type: "string";
	masked?: boolean;
	template?: boolean;
}

export interface NumberSchema<T extends number> extends BaseSchema<T> {
	type: "number";
}

export interface BooleanSchema<T extends boolean> extends BaseSchema<T> {
	type: "boolean";
}

export interface AutocompleteSchema<T extends string> extends BaseSchema<T> {
	type: "autocomplete";
	options: string[];
	template?: boolean;
}

export interface ArraySchema<T extends unknown[]> extends BaseSchema<T> {
	type: "array";
	itemSchema: Schema<T[number]>;
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
type Object = Record<string, any>;

export interface ObjectSchema<T extends Object> extends BaseSchema<T> {
	type: "object";
	fields: Fields<T>;
}

export type Fields<T extends Object> = {
	[K in keyof T]: Schema<T[K]>;
};

export function eachField<T extends Object>(fields: Fields<T>) {
	return Object.entries(fields) as Array<[keyof T & string, Schema<T[keyof T]>]>;
}

export type Schema<T> = [T] extends [string] ? StringSchema<T> | AutocompleteSchema<T> : [T] extends [number] ? NumberSchema<T> : [T] extends [boolean] ? BooleanSchema<T> : [T] extends [unknown[]] ? ArraySchema<T> : [T] extends [Object] ? ObjectSchema<T> : never;

export function getDefaultValueForSchema<T>(schema: Schema<T>): T {
	if (schema.defaultValue !== undefined) {
		return schema.defaultValue;
	}
	switch (schema.type) {
		case "object": {
			const obj: Object = {};
			for (const [fieldName, fieldSchema] of eachField(schema.fields)) {
				obj[fieldName] = getDefaultValueForSchema(fieldSchema);
			}
			return obj as T;
		}
		case "array":
			return [] as T;
		case "boolean":
			return false as T;
		case "number":
			return 0 as T;
		case "string":
		case "autocomplete":
			return "" as T;
		default:
			throw new Error("Unsupported schema type");
	}
}
