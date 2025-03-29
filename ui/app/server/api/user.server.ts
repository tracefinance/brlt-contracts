import { ApiClient } from './client';
import { API_ENDPOINTS } from './endpoints';

// Define basic User model for use until the actual model is available
export interface User {
  id: string;
  email: string;
  name?: string;
  createdAt: string;
  updatedAt: string;
}

/**
 * Update user request payload
 */
export interface UpdateUserRequest {
  name?: string;
  email?: string;
  password?: string;
}

/**
 * Client for interacting with user-related API endpoints
 */
export class UserClient {
  private client: ApiClient;
  
  /**
   * Creates a new user client
   * @param token Authentication token
   */
  constructor(token: string) {
    this.client = new ApiClient(token);
  }
  
  /**
   * Gets the current user's profile
   * @returns User profile information
   */
  async getProfile(): Promise<User> {
    const data = await this.client.get<User>(API_ENDPOINTS.USERS.PROFILE);
    return data;
  }
  
  /**
   * Gets a user by ID
   * @param id User ID
   * @returns User information
   */
  async getUser(id: string): Promise<User> {
    const endpoint = API_ENDPOINTS.USERS.BY_ID(id);
    const data = await this.client.get<User>(endpoint);
    return data;
  }
  
  /**
   * Updates the current user's profile
   * @param request Update request data
   * @returns Updated user information
   */
  async updateProfile(request: UpdateUserRequest): Promise<User> {
    const data = await this.client.put<User>(API_ENDPOINTS.USERS.PROFILE, request);
    return data;
  }
} 