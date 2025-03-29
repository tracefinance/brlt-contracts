/**
 * Base API URL from environment variable or default to localhost
 */
const API_URL = process.env.API_URL || 'http://localhost:8080/api/v1';

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

/**
 * Makes a request to the API
 * @param path API endpoint path
 * @param options Request options
 * @param token Optional authentication token
 * @returns Response data
 */
export async function apiRequest<T>(
  path: string,
  options: ApiRequestOptions = {},
  token?: string
): Promise<T> {
  const { method = 'GET', body, params, headers = {} } = options;
  
  // Build URL with query parameters
  let url = `${API_URL}${path}`;
  if (params && Object.keys(params).length > 0) {
    const searchParams = new URLSearchParams();
    Object.entries(params).forEach(([key, value]) => {
      if (value !== undefined && value !== null) {
        searchParams.append(key, String(value));
      }
    });
    url = `${url}?${searchParams.toString()}`;
  }
  
  // Prepare headers
  const requestHeaders = new Headers(headers);
  requestHeaders.set('Content-Type', 'application/json');
  
  if (token) {
    requestHeaders.set('Authorization', `Bearer ${token}`);
  }
  
  // Prepare request options
  const requestOptions: RequestInit = {
    method,
    headers: requestHeaders,
  };
  
  // Add body for non-GET requests
  if (body && method !== 'GET') {
    requestOptions.body = JSON.stringify(body);
  }
  
  try {
    // Make the request
    const response = await fetch(url, requestOptions);
    
    // Handle non-JSON responses
    const contentType = response.headers.get('content-type');
    if (!contentType || !contentType.includes('application/json')) {
      if (!response.ok) {
        throw new Response(
          JSON.stringify({ message: `API Error: ${response.statusText}` }),
          { 
            status: response.status,
            headers: { 'Content-Type': 'application/json' }
          }
        );
      }
      return {} as T;
    }
    
    // Parse JSON response
    const data = await response.json();
    
    // Handle error responses
    if (!response.ok) {
      throw new Response(
        JSON.stringify(data),
        { 
          status: response.status,
          headers: { 'Content-Type': 'application/json' }
        }
      );
    }
    
    return data as T;
  } catch (error) {
    // Handle fetch errors (network issues, etc.)
    if (error instanceof Error && !(error instanceof Response)) {
      throw new Response(
        JSON.stringify({ message: `Network error: ${error.message}` }),
        { 
          status: 500,
          headers: { 'Content-Type': 'application/json' }
        }
      );
    }
    
    throw error;
  }
}

/**
 * Helper function to make a GET request
 */
export function apiGet<T>(path: string, params?: Record<string, string | number | boolean>, token?: string): Promise<T> {
  return apiRequest<T>(path, { method: 'GET', params }, token);
}

/**
 * Helper function to make a POST request
 */
export function apiPost<T>(path: string, body?: any, token?: string): Promise<T> {
  return apiRequest<T>(path, { method: 'POST', body }, token);
}

/**
 * Helper function to make a PUT request
 */
export function apiPut<T>(path: string, body?: any, token?: string): Promise<T> {
  return apiRequest<T>(path, { method: 'PUT', body }, token);
}

/**
 * Helper function to make a DELETE request
 */
export function apiDelete<T>(path: string, token?: string): Promise<T> {
  return apiRequest<T>(path, { method: 'DELETE' }, token);
} 