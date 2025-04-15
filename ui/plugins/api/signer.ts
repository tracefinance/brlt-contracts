import type {
  IAddAddressRequest,
  ICreateSignerRequest,
  IPagedResponse, // Import IPagedResponse
  ISigner,
  IUpdateSignerRequest,
  IAddress,
} from '~/types';
import { Signer, Address, fromJsonArray } from '~/types';
import { ApiClient } from './client';
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
    await this.client.delete<void>(endpoint);
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
   * Lists signers with pagination
   * @param limit Maximum number of signers to return (default: 10)
   * @param offset Number of signers to skip for pagination (default: 0)
   * @returns Paginated list of signers
   */
  async listSigners(limit: number = 10, offset: number = 0): Promise<IPagedResponse<ISigner>> { // Use IPagedResponse<ISigner>
    const params = { limit, offset };
    const data = await this.client.get<any>(API_ENDPOINTS.SIGNERS.BASE, params);
    const items = data.items ? data.items.map((item: any) => Signer.fromJson(item)) : [];
    return {
      items: items,
      limit: data.limit,
      offset: data.offset,
      hasMore: data.hasMore,
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
    await this.client.delete<void>(endpoint);
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
