/* eslint-disable @typescript-eslint/no-explicit-any */
import { fromJson } from './model';

/**
 * Types of transactions
 */
export enum TransactionType {
  NATIVE = 'native',
  ERC20 = 'erc20',
  DEPLOY = 'deploy'
}

/**
 * Interface representing a transaction
 */
export interface ITransaction {
  chainType: string;
  hash: string;
  fromAddress: string;
  toAddress: string;
  value: string;
  data?: string;
  nonce: number;
  gasPrice?: string;
  gasLimit?: number;
  type: string;
  tokenAddress?: string;
  tokenSymbol?: string;
  tokenDecimals?: number;
  status: string;
  timestamp: number;

  // MultiSig specific fields from DTO (ensure camelCase matches fromJson logic)
  withdrawalNonce?: number;
  requestId?: string;
  proposalId?: string;
  targetTokenAddress?: string;
  newRecoveryAddress?: string;
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