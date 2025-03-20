// Types based on backend DTO structures
export interface Transaction {
  id: number;
  wallet_id: number;
  chain_type: string;
  hash: string;
  from_address: string;
  to_address: string;
  value: string;
  data?: string;
  nonce: number;
  gas_price?: string;
  gas_limit?: number;
  type: string;
  token_address?: string;
  token_symbol?: string;
  status: string;
  timestamp: number;
  created_at: string;
  updated_at: string;
}

// Frontend-friendly versions with camelCase
export interface TransactionFrontend {
  id: number;
  walletId: number;
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
  status: string;
  timestamp: number;
  createdAt: string;
  updatedAt: string;
}

export interface PagedTransactionsResponse {
  items: Transaction[];
  limit: number;
  offset: number;
  has_more: boolean;
}

export interface SyncTransactionsResponse {
  count: number;
}

// Helper function to convert backend format to frontend format
export function toFrontendTransaction(tx: Transaction): TransactionFrontend {
  return {
    id: tx.id,
    walletId: tx.wallet_id,
    chainType: tx.chain_type,
    hash: tx.hash,
    fromAddress: tx.from_address,
    toAddress: tx.to_address,
    value: tx.value,
    data: tx.data,
    nonce: tx.nonce,
    gasPrice: tx.gas_price,
    gasLimit: tx.gas_limit,
    type: tx.type,
    tokenAddress: tx.token_address,
    tokenSymbol: tx.token_symbol,
    status: tx.status,
    timestamp: tx.timestamp,
    createdAt: tx.created_at,
    updatedAt: tx.updated_at,
  };
} 