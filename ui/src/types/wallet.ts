// Types based on backend DTO structures
export interface Wallet {
  id: number;
  key_id: string;
  chain_type: string;
  address: string;
  name: string;
  tags?: Record<string, string>;
  created_at: string;
  updated_at: string;
}

// Frontend-friendly versions with camelCase
export interface WalletFrontend {
  id: number;
  keyId: string;
  chainType: string;
  address: string;
  name: string;
  tags?: Record<string, string>;
  createdAt: string;
  updatedAt: string;
}

export interface CreateWalletRequest {
  chain_type: string;
  name: string;
  tags?: Record<string, string>;
}

export interface UpdateWalletRequest {
  name: string;
  tags?: Record<string, string>;
}

export interface PagedWalletsResponse {
  items: Wallet[];
  limit: number;
  offset: number;
  has_more: boolean;
}

// Helper function to convert backend format to frontend format
export function toFrontendWallet(wallet: Wallet): WalletFrontend {
  return {
    id: wallet.id,
    keyId: wallet.key_id,
    chainType: wallet.chain_type,
    address: wallet.address,
    name: wallet.name,
    tags: wallet.tags,
    createdAt: wallet.created_at,
    updatedAt: wallet.updated_at,
  };
} 