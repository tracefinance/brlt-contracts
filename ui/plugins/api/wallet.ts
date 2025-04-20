import type {
  ICreateWalletRequest,
  IPagedResponse,
  ITokenBalanceResponse,
  IUpdateWalletRequest,
  IWallet
} from '~/types';
import {
  TokenBalanceResponse,
  Wallet,
  fromJsonArray
} from '~/types';
import type {
  ApiClient
} from './client';
import { API_ENDPOINTS } from './endpoints';

/**
 * Client for interacting with wallet-related API endpoints
 */
export class WalletClient {
  private client: ApiClient;
  
  /**
   * Creates a new wallet client
   * @param client API client instance
   */
  constructor(client: ApiClient) {
    this.client = client;
  }
  
  /**
   * Creates a new wallet
   * @param request Wallet creation request
   * @returns Created wallet
   */
  async createWallet(request: ICreateWalletRequest): Promise<IWallet> {
    const data = await this.client.post<any>(API_ENDPOINTS.WALLETS.BASE, request);
    return Wallet.fromJson(data);
  }
  
  /**
   * Gets a wallet by its chain type and address
   * @param chainType Blockchain network type (e.g., ethereum, bitcoin)
   * @param address Wallet address
   * @returns Wallet details
   */
  async getWallet(chainType: string, address: string): Promise<IWallet> {
    const endpoint = API_ENDPOINTS.WALLETS.BY_ADDRESS(chainType, address);
    const data = await this.client.get<any>(endpoint);
    return Wallet.fromJson(data);
  }
  
  /**
   * Updates a wallet's properties
   * @param chainType Blockchain network type
   * @param address Wallet address
   * @param request Wallet update request
   * @returns Updated wallet
   */
  async updateWallet(
    chainType: string,
    address: string,
    request: IUpdateWalletRequest
  ): Promise<IWallet> {
    const endpoint = API_ENDPOINTS.WALLETS.BY_ADDRESS(chainType, address);
    const data = await this.client.put<any>(endpoint, request);
    return Wallet.fromJson(data);
  }
  
  /**
   * Deletes a wallet
   * @param chainType Blockchain network type
   * @param address Wallet address
   */
  async deleteWallet(chainType: string, address: string): Promise<void> {
    const endpoint = API_ENDPOINTS.WALLETS.BY_ADDRESS(chainType, address);
    await this.client.delete(endpoint);
  }
  
  /**
   * Lists wallets with token-based pagination
   * @param limit Maximum number of wallets to return (default: 10)
   * @param nextToken Token for retrieving the next page of results (default: undefined)
   * @returns Paginated list of wallets
   */
  async listWallets(limit: number = 10, nextToken?: string): Promise<IPagedResponse<IWallet>> {
    const params: Record<string, any> = { limit };
    if (nextToken) {
      params.next_token = nextToken;
    }
    
    const data = await this.client.get<any>(API_ENDPOINTS.WALLETS.BASE, params);
    return {
      items: fromJsonArray<IWallet>(data.items || []),
      limit: data.limit,
      nextToken: data.nextToken
    };
  }
  
  /**
   * Gets the balance of a wallet
   * @param chainType Blockchain network type
   * @param address Wallet address
   * @returns Array of token balances
   */
  async getWalletBalance(chainType: string, address: string): Promise<ITokenBalanceResponse[]> {
    const endpoint = API_ENDPOINTS.WALLETS.BALANCE(chainType, address);
    const data = await this.client.get<any[]>(endpoint);
    return data.map(item => TokenBalanceResponse.fromJson(item));
  }
  
  /**
   * Activates a token for a wallet (creates a token balance entry)
   * @param chainType Blockchain network type
   * @param address Wallet address
   * @param tokenAddress Token contract address
   */
  async activateToken(chainType: string, address: string, tokenAddress: string): Promise<void> {
    const endpoint = API_ENDPOINTS.WALLETS.ACTIVATE_TOKEN(chainType, address);
    await this.client.post(endpoint, { tokenAddress });
  }
} 