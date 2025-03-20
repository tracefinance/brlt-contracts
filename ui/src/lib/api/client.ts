/**
 * Core API client for handling requests to the backend
 */

// Base configuration
const API_CONFIG = {
  baseUrl: '/api/v1',
  defaultHeaders: {
    'Content-Type': 'application/json',
  }
};

// Standard error response structure
export interface ApiError {
  message: string;
  status: number;
  details?: any;
}

// Generic API response with proper typing
export type ApiResponse<T> = {
  data: T;
  error?: ApiError;
};

/**
 * Base API request method with error handling
 */
export async function apiRequest<T>(
  endpoint: string,
  options: RequestInit = {}
): Promise<ApiResponse<T>> {
  const url = `${API_CONFIG.baseUrl}${endpoint}`;
  
  // Merge default headers with provided headers
  const headers = {
    ...API_CONFIG.defaultHeaders,
    ...options.headers,
  };

  try {
    const response = await fetch(url, {
      ...options,
      headers,
    });

    // Log request details in development
    if (process.env.NODE_ENV === 'development') {
      console.log(`API ${options.method || 'GET'} ${url}`, { 
        status: response.status 
      });
    }

    // Parse the JSON response
    let data: any;
    const contentType = response.headers.get('content-type');
    if (contentType && contentType.includes('application/json')) {
      data = await response.json();
    } else {
      data = await response.text();
    }

    // Handle HTTP error responses
    if (!response.ok) {
      const error: ApiError = {
        message: data.message || 'An error occurred',
        status: response.status,
        details: data.details || data
      };
      
      return { data: null as any, error };
    }

    return { data, error: undefined };
  } catch (error) {
    console.error('API request failed:', error);
    
    // Handle network/fetch errors
    return {
      data: null as any,
      error: {
        message: error instanceof Error ? error.message : 'Network request failed',
        status: 0,
        details: error
      }
    };
  }
}

/**
 * HTTP method wrappers with URL parameter support
 */
export const api = {
  get: <T>(endpoint: string, params?: Record<string, any>, options?: RequestInit) => {
    const queryString = params ? buildQueryParams(params) : '';
    return apiRequest<T>(`${endpoint}${queryString}`, { method: 'GET', ...options });
  },
  
  post: <T>(endpoint: string, data?: any, params?: Record<string, any>, options?: RequestInit) => {
    const queryString = params ? buildQueryParams(params) : '';
    return apiRequest<T>(`${endpoint}${queryString}`, { 
      method: 'POST', 
      body: data ? JSON.stringify(data) : undefined,
      ...options 
    });
  },
  
  put: <T>(endpoint: string, data?: any, params?: Record<string, any>, options?: RequestInit) => {
    const queryString = params ? buildQueryParams(params) : '';
    return apiRequest<T>(`${endpoint}${queryString}`, { 
      method: 'PUT', 
      body: data ? JSON.stringify(data) : undefined,
      ...options 
    });
  },
  
  delete: <T>(endpoint: string, params?: Record<string, any>, options?: RequestInit) => {
    const queryString = params ? buildQueryParams(params) : '';
    return apiRequest<T>(`${endpoint}${queryString}`, { method: 'DELETE', ...options });
  },
  
  patch: <T>(endpoint: string, data?: any, params?: Record<string, any>, options?: RequestInit) => {
    const queryString = params ? buildQueryParams(params) : '';
    return apiRequest<T>(`${endpoint}${queryString}`, { 
      method: 'PATCH', 
      body: data ? JSON.stringify(data) : undefined,
      ...options 
    });
  },
};

/**
 * API endpoints map
 */
export const API_ENDPOINTS = {
  wallets: {
    base: '/wallets',
    byId: (chainType: string, address: string) => `/wallets/${chainType}/${address}`,
  },
  // Add other API endpoints here
};

/**
 * Utility to build query params
 */
export function buildQueryParams(params: Record<string, any>): string {
  const queryParams = new URLSearchParams();
  
  for (const [key, value] of Object.entries(params)) {
    if (value !== undefined && value !== null) {
      // Handle arrays
      if (Array.isArray(value)) {
        value.forEach(item => {
          if (item !== undefined && item !== null) {
            queryParams.append(`${key}[]`, item.toString());
          }
        });
      }
      // Handle objects (convert to JSON string)
      else if (typeof value === 'object' && value !== null) {
        queryParams.append(key, JSON.stringify(value));
      }
      // Handle primitive values
      else {
        queryParams.append(key, value.toString());
      }
    }
  }
  
  const queryString = queryParams.toString();
  return queryString ? `?${queryString}` : '';
} 