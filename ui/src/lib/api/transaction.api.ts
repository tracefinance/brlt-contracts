import { BaseApi } from './base.api';
import { Transaction, PagedTransactions, SyncTransactionsResponse } from '@/types/models/transaction.model';

/**
 * API service for transaction-related operations
 */
export class TransactionApi extends BaseApi {
  private static readonly WALLETS_PATH = '/wallets';
  private static readonly TRANSACTIONS_PATH = '/transactions';

  /**
   * Fetches a list of transactions for a given wallet
   */
  static async getTransactions(
    chainType: string, 
    address: string, 
    limit: number = 10, 
    offset: number = 0
  ): Promise<PagedTransactions> {
    const path = `${this.WALLETS_PATH}/${chainType}/${address}${this.TRANSACTIONS_PATH}`;
    const queryParams = { limit, offset };
    
    const data = await this.get<any>(path, queryParams);
    return PagedTransactions.fromJson(data);
  }

  /**
   * Fetches a specific transaction by ID
   */
  static async getTransaction(id: number): Promise<Transaction> {
    const path = `${this.TRANSACTIONS_PATH}/${id}`;
    
    const data = await this.get<any>(path);
    return Transaction.fromJson(data);
  }

  /**
   * Syncs transactions for a wallet
   */
  static async syncTransactions(
    chainType: string,
    address: string
  ): Promise<SyncTransactionsResponse> {
    const path = `${this.WALLETS_PATH}/${chainType}/${address}${this.TRANSACTIONS_PATH}:sync`;
    
    const data = await this.post<any>(path);
    return SyncTransactionsResponse.fromJson(data);
  }
} 