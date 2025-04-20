import {
  SyncTransactionsResponse,
  Transaction,
  fromJsonArray
} from '~/types';
import type {
  IPagedResponse,
  ISyncTransactionsResponse,
  ITransaction,
} from '~/types';
import type {
  ApiClient
} from './client';
import { API_ENDPOINTS } from './endpoints';

/**
 * Client for interacting with transaction-related API endpoints
 */
export class TransactionClient {
  private client: ApiClient;
  
  /**
   * Creates a new transaction client
   * @param client API client instance
   */
  constructor(client: ApiClient) {
    this.client = client;
  }
  
  /**
   * Lists transactions with token-based pagination and optional filtering
   * @param limit Maximum number of transactions to return (default: 10)
   * @param nextToken Token for retrieving the next page of results (default: undefined)
   * @returns Paginated list of transactions
   */
  async listTransactions(limit: number = 10, nextToken?: string): Promise<IPagedResponse<ITransaction>> {
    const params: Record<string, any> = { limit };
    
    if (nextToken) {
      params.next_token = nextToken;
    }
    
    const data = await this.client.get<any>(API_ENDPOINTS.TRANSACTIONS.BASE, params);
    return {
      items: fromJsonArray<ITransaction>(data.items || []),
      limit: data.limit,
      nextToken: data.nextToken
    };
  }
  
  /**
   * Gets a transaction by its ID
   * @param id Transaction ID
   * @returns Transaction details
   */
  async getTransaction(id: string): Promise<ITransaction> {
    const endpoint = API_ENDPOINTS.TRANSACTIONS.BY_ID(id);
    const data = await this.client.get<any>(endpoint);
    return Transaction.fromJson(data);
  }
  
  /**
   * Gets transactions for a specific wallet
   * @param address Wallet address
   * @param chainType Blockchain network type
   * @param tokenAddress Optional token address to filter transactions by
   * @param limit Maximum number of transactions to return (default: 10)
   * @param nextToken Token for retrieving the next page of results (default: undefined)
   * @returns Paginated list of transactions
   */
  async getWalletTransactions(
    address: string,
    chainType: string,
    tokenAddress?: string,
    limit: number = 10,
    nextToken?: string
  ): Promise<IPagedResponse<ITransaction>> {
    const endpoint = API_ENDPOINTS.TRANSACTIONS.BY_WALLET(chainType, address);
    
    const params: Record<string, any> = { limit };
    
    if (nextToken) {
      params.next_token = nextToken;
    }
    
    if (tokenAddress) {
      params.token_address = tokenAddress;
    }
    
    const data = await this.client.get<any>(endpoint, params);
    return {
      items: fromJsonArray<ITransaction>(data.items || []),
      limit: data.limit,
      nextToken: data.nextToken
    };
  }
  
  /**
   * Syncs transactions for a specific wallet
   * @param chainType Blockchain network type
   * @param address Wallet address
   * @returns Sync response containing count of synced transactions
   */
  async syncTransactions(chainType: string, address: string): Promise<ISyncTransactionsResponse> {
    const endpoint = `${API_ENDPOINTS.TRANSACTIONS.BY_WALLET(chainType, address)}/sync`;
    const data = await this.client.post<any>(endpoint);
    return SyncTransactionsResponse.fromJson(data);
  }
  
  /**
   * Filters transactions based on various criteria
   * @param options Filter options
   * @returns Paginated list of transactions
   */
  async filterTransactions(options: {
    chainType?: string;
    address?: string;
    tokenAddress?: string;
    status?: string;
    limit?: number;
    nextToken?: string;
  }): Promise<IPagedResponse<ITransaction>> {
    const {
      chainType,
      address,
      tokenAddress,
      status,
      limit = 10,
      nextToken
    } = options;
    
    const params: Record<string, any> = { limit };
    
    if (nextToken) {
      params.next_token = nextToken;
    }
    
    if (chainType) params.chain_type = chainType;
    if (address) params.address = address;
    if (tokenAddress) params.token_address = tokenAddress;
    if (status) params.status = status;
    
    const data = await this.client.get<any>(API_ENDPOINTS.TRANSACTIONS.BASE, params);
    return {
      items: fromJsonArray<ITransaction>(data.items || []),
      limit: data.limit,
      nextToken: data.nextToken
    };
  }
} 