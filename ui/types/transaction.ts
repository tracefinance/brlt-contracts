import { fromJson, fromJsonArray } from './model';
import { Token } from './token';
import type { IToken } from './token';
import { Wallet } from './wallet';
import type { IWallet } from './wallet';

/**
 * Interface representing a transaction
 */
export interface ITransaction {
  id: string;
  chainType: string;
  walletAddress: string;
  hash: string;
  blockHash: string;
  blockNumber: number;
  from: string;
  to: string;
  value: string;
  gasPrice: string;
  gasLimit: string;
  gasUsed: string;
  nonce: number;
  status: string;
  timestamp: number;
  data?: string;
  token?: IToken;
  wallet?: IWallet;
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
    
    // Handle nested objects
    if (json.token) {
      transaction.token = Token.fromJson(json.token);
    }
    
    if (json.wallet) {
      transaction.wallet = Wallet.fromJson(json.wallet);
    }
    
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
 * Interface representing a paginated response containing Transactions
 */
export interface IPagedTransactions {
  items: ITransaction[];
  limit: number;
  offset: number;
  hasMore: boolean;
}

/**
 * Factory functions for IPagedTransactions
 */
export const PagedTransactions = {
  /**
   * Converts a plain JSON paged response to IPagedTransactions
   */
  fromJson(json: any): IPagedTransactions {
    const response = fromJson<IPagedTransactions>(json);
    
    // Convert each item in the items array
    if (json.items && Array.isArray(json.items)) {
      response.items = Transaction.fromJsonArray(json.items);
    } else {
      response.items = [];
    }

    return response;
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