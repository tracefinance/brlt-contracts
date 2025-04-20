import type {
  IAddAddressRequest,
  IAddress,
  ICreateSignerRequest,
  IPagedResponse, // Import IPagedResponse
  ISigner,
  IUpdateSignerRequest,
} from '~/types';
import { Address, Signer, fromJsonArray } from '~/types';
import type { ApiClient } from './client';
import { API_ENDPOINTS } from './endpoints';

/**
 * Client for interacting with signer-related API endpoints
 */
export class SignerClient {
  private client: ApiClient;

  /**
   * Creates a new signer client
   * @param client API client instance
   */
  constructor(client: ApiClient) {
    this.client = client;
  }

  /**
   * Creates a new signer
   * @param request Signer creation request
   * @returns Created signer
   */
  async createSigner(request: ICreateSignerRequest): Promise<ISigner> {
    const data = await this.client.post<any>(API_ENDPOINTS.SIGNERS.BASE, request);
    return Signer.fromJson(data);
  }

  /**
   * Updates a signer's properties
   * @param id Signer ID (as string, matching handler param)
   * @param request Signer update request
   * @returns Updated signer
   */
  async updateSigner(id: string, request: IUpdateSignerRequest): Promise<ISigner> {
    const endpoint = API_ENDPOINTS.SIGNERS.BY_ID(id);
    const data = await this.client.put<any>(endpoint, request);
    return Signer.fromJson(data);
  }

  /**
   * Deletes a signer
   * @param id Signer ID (as string)
   */
  async deleteSigner(id: string): Promise<void> {
    const endpoint = API_ENDPOINTS.SIGNERS.BY_ID(id);
    await this.client.delete(endpoint);
  }

  /**
   * Gets a signer by its ID
   * @param id Signer ID (as string)
   * @returns Signer details
   */
  async getSigner(id: string): Promise<ISigner> {
    const endpoint = API_ENDPOINTS.SIGNERS.BY_ID(id);
    const data = await this.client.get<any>(endpoint);
    return Signer.fromJson(data);
  }

  /**
   * Lists signers with token-based pagination
   * @param limit Maximum number of signers to return (default: 10)
   * @param nextToken Token for retrieving the next page of results (default: undefined)
   * @returns Paginated list of signers
   */
  async listSigners(limit: number = 10, nextToken?: string): Promise<IPagedResponse<ISigner>> {
    const params: Record<string, any> = { limit };
    if (nextToken) {
      params.next_token = nextToken;
    }
    
    const data = await this.client.get<any>(API_ENDPOINTS.SIGNERS.BASE, params);
    return {
      items: fromJsonArray<ISigner>(data.items || []),
      limit: data.limit,
      nextToken: data.nextToken
    };
  }

  /**
   * Gets signers associated with a specific user ID
   * @param userId User ID (as string)
   * @returns Array of signers
   */
  async getSignersByUser(userId: string): Promise<ISigner[]> {
    const endpoint = API_ENDPOINTS.SIGNERS.BY_USER_ID(userId);
    const data = await this.client.get<any[]>(endpoint);
    return data ? data.map((item: any) => Signer.fromJson(item)) : [];
  }

  /**
   * Adds an address to a signer
   * @param signerId Signer ID (as string)
   * @param request Address creation request
   * @returns Created address
   */
  async addAddress(signerId: string, request: IAddAddressRequest): Promise<IAddress> {
    const endpoint = API_ENDPOINTS.SIGNERS.ADDRESSES(signerId);
    const data = await this.client.post<any>(endpoint, request);
    return Address.fromJson(data);
  }

  /**
   * Deletes an address from a signer
   * @param signerId Signer ID (as string)
   * @param addressId Address ID (as string)
   */
  async deleteAddress(signerId: string, addressId: string): Promise<void> {
    const endpoint = API_ENDPOINTS.SIGNERS.ADDRESS_BY_ID(signerId, addressId);
    await this.client.delete(endpoint);
  }

  /**
   * Gets all addresses associated with a signer
   * @param signerId Signer ID (as string)
   * @returns Array of addresses
   */
  async getAddresses(signerId: string): Promise<IAddress[]> {
    const endpoint = API_ENDPOINTS.SIGNERS.ADDRESSES(signerId);
    const data = await this.client.get<any[]>(endpoint);
    return Address.fromJsonArray(data);
  }
}
