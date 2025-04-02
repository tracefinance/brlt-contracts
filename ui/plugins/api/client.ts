/**
 * Error response from the API
 */
export interface ApiError {
  code: string;
  message: string;
  details?: Record<string, any>;
}

/**
 * HTTP methods supported by the API client
 */
export type HttpMethod = 'GET' | 'POST' | 'PATCH' | 'PUT' | 'DELETE';

/**
 * Options for making API requests
 */
export interface ApiRequestOptions {
  method?: HttpMethod;
  body?: any;
  params?: Record<string, string | number | boolean>;
  headers?: HeadersInit;
}

import { camelToSnakeCase, snakeToCamelCase } from '~/lib/caseConversion';

/**
 * Base API client class that handles API communication
 */
export class ApiClient {
  private baseUrl: string;
  private token: string | undefined;

  /**
   * Creates a new API client
   * @param token Optional authentication token
   * @param baseUrl Optional custom base URL (defaults to environment variable or runtime config)
   */
  constructor(token?: string, baseUrl?: string) {
    // In Nuxt, we'll use runtime config for the API URL
    this.baseUrl = baseUrl || '';
    this.token = token;
  }

  /**
   * Sets the base URL for API requests
   */
  setBaseUrl(baseUrl: string) {
    this.baseUrl = baseUrl;
  }

  /**
   * Sets the authentication token for API requests
   */
  setToken(token: string) {
    this.token = token;
  }

  /**
   * Makes a request to the API
   * @param path API endpoint path
   * @param options Request options
   * @returns Response data
   */
  async request<T>(path: string, options: ApiRequestOptions = {}): Promise<T> {
    const { method = 'GET', body, params, headers = {} } = options;
    
    // Prepare headers
    const requestHeaders: Record<string, string> = { 
      ...headers as Record<string, string>,
      'Content-Type': 'application/json'
    };
    
    if (this.token) {
      requestHeaders['Authorization'] = `Bearer ${this.token}`;
    }
    
    // Convert request body from camelCase to snake_case if necessary
    const processedBody = method !== 'GET' && body ? camelToSnakeCase<Record<string, any>>(body) : undefined;
    
    try {
      // Make the request using $fetch
      const response = await $fetch(path, {
        baseURL: this.baseUrl,
        method,
        body: processedBody,
        params: params ? camelToSnakeCase(params) : undefined,
        headers: requestHeaders
      });
      
      // Convert response from snake_case to camelCase
      return snakeToCamelCase<T>(response);
    } catch (error: any) {
      // Handle FetchError type from Nuxt
      if (error.name === 'FetchError') {
        const apiError = error.data as ApiError;
        throw new Error(apiError?.message || error.message || `API Error: ${error.status}`);
      }
      
      // Fallback for other error types
      if (error instanceof Error) {
        throw error;
      }
      
      throw new Error('Unknown API error');
    }
  }

  /**
   * Makes a GET request to the API
   * @param path API endpoint path
   * @param params Optional query parameters
   * @returns Response data
   */
  get<T>(path: string, params?: Record<string, string | number | boolean>): Promise<T> {
    return this.request<T>(path, { method: 'GET', params });
  }

  /**
   * Makes a POST request to the API
   * @param path API endpoint path
   * @param body Request body
   * @returns Response data
   */
  post<T>(path: string, body?: any): Promise<T> {
    return this.request<T>(path, { method: 'POST', body });
  }

  /**
   * Makes a PUT request to the API
   * @param path API endpoint path
   * @param body Request body
   * @returns Response data
   */
  put<T>(path: string, body?: any): Promise<T> {
    return this.request<T>(path, { method: 'PUT', body });
  }

  /**
   * Makes a PATCH request to the API
   * @param path API endpoint path
   * @param body Request body
   * @returns Response data
   */
  patch<T>(path: string, body?: any): Promise<T> {
    return this.request<T>(path, { method: 'PATCH', body });
  }

  /**
   * Makes a DELETE request to the API
   * @param path API endpoint path
   * @returns Response data
   */
  delete<T>(path: string): Promise<T> {
    return this.request<T>(path, { method: 'DELETE' });
  }
}

// Export a function to create a new API client
export function createApiClient(token?: string, baseUrl?: string): ApiClient {
  return new ApiClient(token, baseUrl);
} 