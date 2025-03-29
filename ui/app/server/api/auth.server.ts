import { ApiClient } from './client';
import { API_ENDPOINTS } from './endpoints';
import { User } from './user.server';

/**
 * Login request payload
 */
export interface LoginRequest {
  email: string;
  password: string;
}

/**
 * Login response data
 */
export interface LoginResponse {
  token: string;
  user: User;
}

/**
 * Client for interacting with authentication-related API endpoints
 */
export class AuthClient {
  private client: ApiClient;
  
  /**
   * Creates a new auth client
   * @param token Optional authentication token
   */
  constructor(token?: string) {
    this.client = new ApiClient(token);
  }
  
  /**
   * Authenticates a user with email and password
   * @param email User email
   * @param password User password
   * @returns Login response with token and user information
   */
  async login(email: string, password: string): Promise<LoginResponse> {
    const data = await this.client.post<any>(
      API_ENDPOINTS.AUTH.LOGIN,
      { email, password }
    );
    
    return {
      token: data.token,
      user: data.user as User,
    };
  }
  
  /**
   * Logs out the current user
   */
  async logout(): Promise<void> {
    await this.client.post<void>(API_ENDPOINTS.AUTH.LOGOUT);
  }
  
  /**
   * Gets the current authenticated user's information
   * @returns User information
   */
  async getCurrentUser(): Promise<User> {
    const data = await this.client.get<User>(API_ENDPOINTS.AUTH.ME);
    return data;
  }
}

// For backward compatibility - these will be deprecated in future versions

/**
 * Authenticates a user with email and password
 * @param email User email
 * @param password User password
 * @returns Login response with token and user information
 */
export async function login(
  email: string,
  password: string
): Promise<LoginResponse> {
  const client = new AuthClient();
  return client.login(email, password);
}

/**
 * Logs out the current user
 * @param token Authentication token
 */
export async function logout(token: string): Promise<void> {
  const client = new AuthClient(token);
  return client.logout();
}

/**
 * Gets the current authenticated user's information
 * @param token Authentication token
 * @returns User information
 */
export async function getCurrentUser(token: string): Promise<User> {
  const client = new AuthClient(token);
  return client.getCurrentUser();
} 