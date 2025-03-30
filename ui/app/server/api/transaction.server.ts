import {
  PagedTransactions,
  SyncTransactionsResponse,
  Transaction
} from '~/models/transaction';
import {
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
   * @param token Authentication token
   */
  constructor(token: string) {
    this.client = new ApiClient(token);
  }
  
  /**
   * Gets a transaction by its hash
   * @param hash Transaction hash
   * @returns Transaction details
   */
  async getTransaction(hash: string): Promise<Transaction> {
    const endpoint = API_ENDPOINTS.TRANSACTIONS.BY_ID(hash);
    const data = await this.client.get<any>(endpoint);
    return Transaction.fromJson(data);
  }
  
  /**
   * Gets transactions for a specific wallet
   * @param chainType Blockchain network type
   * @param address Wallet address
   * @param limit Maximum number of transactions to return (default: 10)
   * @param offset Number of transactions to skip for pagination (default: 0)
   * @param tokenAddress Optional token address to filter by
   * @returns Paginated list of transactions
   */
  async getTransactionsByAddress(
    chainType: string,
    address: string,
    limit: number = 10,
    offset: number = 0,
    tokenAddress?: string
  ): Promise<PagedTransactions> {
    const endpoint = API_ENDPOINTS.TRANSACTIONS.BY_WALLET(address, chainType);
    const params: Record<string, any> = { limit, offset };
    
    if (tokenAddress) {
      params.token_address = tokenAddress;
    }
    
    const data = await this.client.get<any>(endpoint, params);
    return PagedTransactions.fromJson(data);
  }
  
  /**
   * Syncs transactions for a specific wallet
   * @param chainType Blockchain network type
   * @param address Wallet address
   * @returns Sync response containing count of synced transactions
   */
  async syncTransactions(chainType: string, address: string): Promise<SyncTransactionsResponse> {
    const endpoint = `${API_ENDPOINTS.TRANSACTIONS.BY_WALLET(address, chainType)}/sync`;
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
    offset?: number;
  }): Promise<PagedTransactions> {
    const {
      chainType,
      address,
      tokenAddress,
      status,
      limit = 10,
      offset = 0
    } = options;
    
    const params: Record<string, any> = { limit, offset };
    
    if (chainType) params.chain_type = chainType;
    if (address) params.address = address;
    if (tokenAddress) params.token_address = tokenAddress;
    if (status) params.status = status;
    
    const data = await this.client.get<any>(API_ENDPOINTS.TRANSACTIONS.BASE, params);
    return PagedTransactions.fromJson(data);
  }
} 