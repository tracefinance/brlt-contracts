import type { IPagedResponse } from '~/types';
import type { 
  IVault, 
  ICreateVaultRequest, 
  IUpdateVaultRequest, 
  ITokenRequest, 
  ITokenActionResponse, 
  IRecoveryResponse,
  IVaultFilter
} from '~/types/vault';
import { Vault } from '~/types/vault';
import type { ApiClient } from './client';
import { API_ENDPOINTS } from './endpoints';

/**
 * Client for interacting with vault-related API endpoints
 */
export class VaultClient {
  private client: ApiClient;

  /**
   * Creates a new vault client
   * @param client API client instance
   */
  constructor(client: ApiClient) {
    this.client = client;
  }

  /**
   * Creates a new vault
   * @param request Vault creation parameters
   * @returns The created vault
   */
  async createVault(request: ICreateVaultRequest): Promise<IVault> {
    const data = await this.client.post<any>(API_ENDPOINTS.VAULTS.BASE, request);
    return Vault.fromJson(data);
  }

  /**
   * Lists vaults with pagination and optional filtering
   * @param limit Maximum number of items to return
   * @param nextToken Pagination token for fetching the next page
   * @param filter Optional filter parameters
   * @returns Paginated list of vaults
   */
  async listVaults(
    limit: number = 10, 
    nextToken?: string, 
    filter?: IVaultFilter
  ): Promise<IPagedResponse<IVault>> {
    const params: Record<string, any> = { limit };
    
    if (nextToken) {
      params.next_token = nextToken;
    }
    
    if (filter?.status) {
      params.status = filter.status;
    }
    
    if (filter?.address) {
      params.address = filter.address;
    }
    
    const data = await this.client.get<any>(API_ENDPOINTS.VAULTS.BASE, params);
    return {
      items: Vault.fromJsonArray(data.items || []),
      limit: data.limit,
      nextToken: data.next_token
    };
  }

  /**
   * Gets a specific vault by ID
   * @param id Vault ID
   * @returns Vault details
   */
  async getVault(id: number): Promise<IVault> {
    const endpoint = API_ENDPOINTS.VAULTS.BY_ID(id);
    const data = await this.client.get<any>(endpoint);
    return Vault.fromJson(data);
  }

  /**
   * Updates a vault's name
   * @param id Vault ID
   * @param request Update parameters
   * @returns Updated vault
   */
  async updateVault(id: number, request: IUpdateVaultRequest): Promise<IVault> {
    const endpoint = API_ENDPOINTS.VAULTS.BY_ID(id);
    const data = await this.client.put<any>(endpoint, request);
    return Vault.fromJson(data);
  }

  /**
   * Adds a token to a vault's whitelist
   * @param vaultId Vault ID
   * @param request Token address
   * @returns Token action response
   */
  async addToken(vaultId: number, request: ITokenRequest): Promise<ITokenActionResponse> {
    const endpoint = API_ENDPOINTS.VAULTS.TOKENS(vaultId);
    return await this.client.post<ITokenActionResponse>(endpoint, request);
  }

  /**
   * Removes a token from a vault's whitelist
   * @param vaultId Vault ID
   * @param tokenAddress Token address
   * @returns Token action response
   */
  async removeToken(vaultId: number, tokenAddress: string): Promise<ITokenActionResponse> {
    const endpoint = API_ENDPOINTS.VAULTS.TOKEN_BY_ADDRESS(vaultId, tokenAddress);
    return await this.client.delete<ITokenActionResponse>(endpoint);
  }

  /**
   * Starts the recovery process for a vault
   * @param vaultId Vault ID
   * @returns Recovery response
   */
  async startRecovery(vaultId: number): Promise<IRecoveryResponse> {
    const endpoint = API_ENDPOINTS.VAULTS.RECOVERY_START(vaultId);
    return await this.client.post<IRecoveryResponse>(endpoint);
  }

  /**
   * Cancels an in-progress recovery process
   * @param vaultId Vault ID
   * @returns Recovery response
   */
  async cancelRecovery(vaultId: number): Promise<IRecoveryResponse> {
    const endpoint = API_ENDPOINTS.VAULTS.RECOVERY_CANCEL(vaultId);
    return await this.client.post<IRecoveryResponse>(endpoint);
  }

  /**
   * Executes a recovery process after the waiting period
   * @param vaultId Vault ID
   * @returns Recovery response
   */
  async executeRecovery(vaultId: number): Promise<IRecoveryResponse> {
    const endpoint = API_ENDPOINTS.VAULTS.RECOVERY_EXECUTE(vaultId);
    return await this.client.post<IRecoveryResponse>(endpoint);
  }
} 