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
 * Base API client class that handles API communication
 */
export class ApiClient {
  private baseUrl: string;
  private token: string | undefined;

  /**
   * Creates a new API client
   * @param token Optional authentication token
   * @param baseUrl Optional custom base URL (defaults to environment variable or localhost)
   */
  constructor(token?: string, baseUrl: string = API_URL) {
    this.baseUrl = baseUrl;
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
    
    // Build URL with query parameters
    let url = `${this.baseUrl}${path}`;
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
    
    if (this.token) {
      requestHeaders.set('Authorization', `Bearer ${this.token}`);
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

// Export a function to create a new API client for backwards compatibility
export function createApiClient(token?: string): ApiClient {
  return new ApiClient(token);
}

// For backward compatibility - these will be deprecated in future versions
export async function apiRequest<T>(
  path: string,
  options: ApiRequestOptions = {},
  token?: string
): Promise<T> {
  const client = new ApiClient(token);
  return client.request<T>(path, options);
}

export function apiGet<T>(path: string, params?: Record<string, string | number | boolean>, token?: string): Promise<T> {
  const client = new ApiClient(token);
  return client.get<T>(path, params);
}

export function apiPost<T>(path: string, body?: any, token?: string): Promise<T> {
  const client = new ApiClient(token);
  return client.post<T>(path, body);
}

export function apiPut<T>(path: string, body?: any, token?: string): Promise<T> {
  const client = new ApiClient(token);
  return client.put<T>(path, body);
}

export function apiDelete<T>(path: string, token?: string): Promise<T> {
  const client = new ApiClient(token);
  return client.delete<T>(path);
} 