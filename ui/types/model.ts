import { snakeToCamelCase, camelToSnakeCase } from '~/lib/caseConversion';

/**
 * Converts a snake_case JSON object to a camelCase typed interface
 * @param json The plain JSON object
 * @returns A typed interface with camelCase keys
 */
export function fromJson<T>(json: any): T {
  return snakeToCamelCase<T>(json);
}

/**
 * Converts an array of snake_case JSON objects to camelCase typed interfaces
 * @param jsonArray The array of plain JSON objects
 * @returns An array of typed interfaces with camelCase keys
 */
export function fromJsonArray<T>(jsonArray: any[]): T[] {
  return jsonArray.map(json => fromJson<T>(json));
}

/**
 * Converts a camelCase typed interface to a snake_case JSON object
 * @param data The typed interface with camelCase keys
 * @returns A plain JSON object with snake_case keys
 */
export function toJson<T>(data: T): any {
  return camelToSnakeCase(data);
}

/**
 * Generic interface for paginated API responses
 */
export interface IPagedResponse<T> {
  items: T[];
  limit: number;
  offset: number;
  hasMore: boolean;
} 