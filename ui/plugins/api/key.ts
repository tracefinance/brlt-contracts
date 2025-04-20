import type {
  ICreateKeyRequest,
  IKey,
  IPagedResponse,
  IUpdateKeyRequest,
  IImportKeyRequest,
  ISignDataRequest,
  ISignDataResponse,
} from '~/types';
import { Key } from '~/types';
import type { ApiClient } from './client';
import { API_ENDPOINTS } from './endpoints';

/**
 * Client for interacting with key management API endpoints
 */
export class KeyClient {
  private client: ApiClient;

  /**
   * Creates a new key client
   * @param client API client instance
   */
  constructor(client: ApiClient) {
    this.client = client;
  }

  /**
   * Creates a new cryptographic key
   * @param request Key creation request
   * @returns Created key
   */
  async createKey(request: ICreateKeyRequest): Promise<IKey> {
    const data = await this.client.post<any>(API_ENDPOINTS.KEYS.BASE, request);
    return Key.fromJson(data);
  }

  /**
   * Deletes a key
   * @param id Key ID
   */
  async deleteKey(id: string): Promise<void> {
    const endpoint = API_ENDPOINTS.KEYS.BY_ID(id);
    await this.client.delete(endpoint);
  }

  /**
   * Gets a key by ID
   * @param id Key ID
   * @returns Key details
   */
  async getKey(id: string): Promise<IKey> {
    const endpoint = API_ENDPOINTS.KEYS.BY_ID(id);
    const data = await this.client.get<any>(endpoint);
    return Key.fromJson(data);
  }

  /**
   * Lists keys with token-based pagination
   * @param limit Maximum number of keys to return (default: 10)
   * @param nextToken Token for retrieving the next page of results (default: undefined)
   * @returns Paginated list of keys
   */
  async listKeys(limit: number = 10, nextToken?: string): Promise<IPagedResponse<IKey>> {
    const params: Record<string, any> = { limit };
    if (nextToken) {
      params.next_token = nextToken;
    }
    
    const data = await this.client.get<any>(API_ENDPOINTS.KEYS.BASE, params);
    return {
      items: Key.fromJsonArray(data.items || []),
      limit: data.limit,
      nextToken: data.nextToken
    };
  }

  /**
   * Updates key metadata
   * @param id Key ID
   * @param request Update request
   * @returns Updated key
   */
  async updateKey(id: string, request: IUpdateKeyRequest): Promise<IKey> {
    const endpoint = API_ENDPOINTS.KEYS.BY_ID(id);
    const data = await this.client.put<any>(endpoint, request);
    return Key.fromJson(data);
  }

  /**
   * Imports an existing key
   * @param request Import request
   * @returns Imported key
   */
  async importKey(request: IImportKeyRequest): Promise<IKey> {
    const data = await this.client.post<any>(API_ENDPOINTS.KEYS.IMPORT, request);
    return Key.fromJson(data);
  }

  /**
   * Signs data using a specific key
   * @param id Key ID
   * @param request Sign data request
   * @returns Signature response
   */
  async signData(id: string, request: ISignDataRequest): Promise<ISignDataResponse> {
    const endpoint = API_ENDPOINTS.KEYS.SIGN(id);
    // Assuming the backend returns ISignDataResponse directly
    return await this.client.post<ISignDataResponse>(endpoint, request);
  }

  /* // Commented out - No backend route
  async exportPublicKey(id: string, format: 'pem' | 'der' | 'jwk' = 'pem'): Promise<IPublicKeyExport> {
    const endpoint = API_ENDPOINTS.KEYS.EXPORT(id);
    return await this.client.get<IPublicKeyExport>(endpoint, { format });
  }
  */

  /* // Commented out - No backend route
  async verifyKey(id: string): Promise<boolean> {
    const endpoint = API_ENDPOINTS.KEYS.VERIFY(id);
    const data = await this.client.post<any>(endpoint, {});
    return data.verified === true;
  }
  */

  /* // Commented out - No backend route
  async getKeyAuditLogs(id: string, limit: number = 20): Promise<any[]> {
    const endpoint = API_ENDPOINTS.KEYS.AUDIT(id);
    const data = await this.client.get<any>(endpoint, { limit });
    return data.logs || [];
  }
  */
} 