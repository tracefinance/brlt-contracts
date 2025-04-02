/**
 * Converts a snake_case string to camelCase
 */
export function snakeToCamel(str: string): string {
  return str.replace(/_([a-z])/g, (_, letter) => letter.toUpperCase());
}

/**
 * Converts a camelCase string to snake_case
 */
export function camelToSnake(str: string): string {
  return str.replace(/([A-Z])/g, (_, letter) => `_${letter.toLowerCase()}`);
}

/**
 * Transforms all keys in an object (including nested objects and arrays)
 * using the provided transformer function
 */
export function transformKeys<T>(obj: Record<string, any>, transformer: (key: string) => string): T {
  if (!obj || typeof obj !== 'object') return obj as unknown as T;
  
  if (Array.isArray(obj)) {
    return obj.map(item => transformKeys(item, transformer)) as unknown as T;
  }
  
  return Object.keys(obj).reduce((result, key) => {
    const value = obj[key];
    const transformedKey = transformer(key);
    
    // Handle nested objects recursively
    result[transformedKey] = value && typeof value === 'object' 
      ? transformKeys(value, transformer) 
      : value;
      
    return result;
  }, {} as Record<string, any>) as T;
}

/**
 * Transforms all keys in an object from snake_case to camelCase
 */
export function snakeToCamelCase<T>(data: any): T {
  return transformKeys<T>(data, snakeToCamel);
}

/**
 * Transforms all keys in an object from camelCase to snake_case
 */
export function camelToSnakeCase<T>(data: any): T {
  return transformKeys<T>(data, camelToSnake);
} 