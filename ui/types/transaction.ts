import { fromJson } from './model';

/**
 * Interface representing a transaction
 */
export interface ITransaction {
  id: number;
  chainType: string;
  walletId?: number;
  hash: string;
  fromAddress: string;
  toAddress: string;
  value: string;
  gasPrice: string;
  gasLimit: number;
  nonce: number;
  type?: string;
  status: string;
  timestamp: number;
  data?: string;
  tokenSymbol?: string;
  tokenAddress?: string;
  createdAt: string;
  updatedAt: string;
}

/**
 * Factory functions for ITransaction
 */
export const Transaction = {
  /**
   * Converts a plain JSON object from the API to an ITransaction
   */
  fromJson(json: any): ITransaction {
    const transaction = fromJson<ITransaction>(json);
    return transaction;
  },

  /**
   * Converts an array of plain JSON objects from the API to ITransaction objects
   */
  fromJsonArray(jsonArray: any[]): ITransaction[] {
    return jsonArray.map(json => Transaction.fromJson(json));
  }
};

/**
 * Interface representing a sync transactions response
 */
export interface ISyncTransactionsResponse {
  count: number;
  status: string;
}

/**
 * Factory functions for ISyncTransactionsResponse
 */
export const SyncTransactionsResponse = {
  /**
   * Converts a plain JSON object from the API to an ISyncTransactionsResponse
   */
  fromJson(json: any): ISyncTransactionsResponse {
    return fromJson<ISyncTransactionsResponse>(json);
  }
}; 