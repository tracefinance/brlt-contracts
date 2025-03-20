/**
 * Base API class to abstract common API operations
 */
export abstract class BaseApi {
  protected static readonly BASE_PATH = '/api/v1';

  /**
   * Builds a complete API URL
   */
  protected static buildUrl(path: string, queryParams?: Record<string, any>): string {
    let url = `${this.BASE_PATH}${path}`;
    
    if (queryParams && Object.keys(queryParams).length > 0) {
      const params = new URLSearchParams();
      Object.entries(queryParams).forEach(([key, value]) => {
        if (value !== undefined && value !== null) {
          params.append(key, value.toString());
        }
      });
      url += `?${params.toString()}`;
    }
    
    return url;
  }

  /**
   * Performs a GET request
   */
  protected static async get<T>(path: string, queryParams?: Record<string, any>): Promise<T> {
    const url = this.buildUrl(path, queryParams);
    const response = await fetch(url);
    return this.handleResponse<T>(response);
  }

  /**
   * Performs a POST request
   */
  protected static async post<T>(path: string, body?: any, queryParams?: Record<string, any>): Promise<T> {
    const url = this.buildUrl(path, queryParams);
    const response = await fetch(url, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: body ? JSON.stringify(body) : undefined
    });
    return this.handleResponse<T>(response);
  }

  /**
   * Performs a PUT request
   */
  protected static async put<T>(path: string, body?: any, queryParams?: Record<string, any>): Promise<T> {
    const url = this.buildUrl(path, queryParams);
    const response = await fetch(url, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json'
      },
      body: body ? JSON.stringify(body) : undefined
    });
    return this.handleResponse<T>(response);
  }

  /**
   * Performs a DELETE request
   */
  protected static async delete<T>(path: string, queryParams?: Record<string, any>): Promise<T> {
    const url = this.buildUrl(path, queryParams);
    const response = await fetch(url, {
      method: 'DELETE'
    });
    return this.handleResponse<T>(response);
  }

  /**
   * Handles API responses and errors
   */
  private static async handleResponse<T>(response: Response): Promise<T> {
    if (!response.ok) {
      throw new Error(`API Error: ${response.status} ${response.statusText}`);
    }
    
    // Handle empty responses
    if (response.status === 204 || response.headers.get('content-length') === '0') {
      return {} as T;
    }
    
    return await response.json();
  }
} 