import { Transaction, TransactionFrontend, PagedTransactionsResponse, SyncTransactionsResponse, toFrontendTransaction } from "@/types/transaction";
import { api, API_ENDPOINTS } from "@/lib/api/client";

/**
 * Get transactions for a wallet by chain type and address
 */
export async function getTransactionsByWallet(
  chainType: string, 
  address: string, 
  limit: number = 10, 
  offset: number = 0
): Promise<{ transactions: TransactionFrontend[], hasMore: boolean }> {
  const endpoint = `/wallets/${chainType}/${address}/transactions`;
  
  const { data, error } = await api.get<PagedTransactionsResponse>(
    endpoint,
    { limit, offset }
  );
  
  if (error) {
    throw new Error(error.message);
  }
  
  // Convert backend transactions to frontend format
  return {
    transactions: data.items.map(toFrontendTransaction),
    hasMore: data.has_more
  };
}

/**
 * Get a single transaction by hash
 */
export async function getTransaction(hash: string): Promise<TransactionFrontend> {
  const endpoint = `/transactions/${hash}`;
  const { data, error } = await api.get<Transaction>(endpoint);
  
  if (error) {
    throw new Error(error.message);
  }
  
  // Convert to frontend format
  return toFrontendTransaction(data);
}

/**
 * Sync transactions for a wallet
 */
export async function syncTransactions(chainType: string, address: string): Promise<number> {
  const endpoint = `/wallets/${chainType}/${address}/transactions/sync`;
  const { data, error } = await api.post<SyncTransactionsResponse>(endpoint);
  
  if (error) {
    throw new Error(error.message);
  }
  
  return data.count;
} 