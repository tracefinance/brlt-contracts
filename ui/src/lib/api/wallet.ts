import { Wallet, WalletFrontend, CreateWalletRequest, UpdateWalletRequest, PagedWalletsResponse, toFrontendWallet } from "@/types/wallet";
import { api, API_ENDPOINTS } from "@/lib/api/client";

/**
 * Get a paginated list of wallets
 */
export async function getWallets(limit: number = 10, offset: number = 0): Promise<{ wallets: WalletFrontend[], hasMore: boolean }> {
  // Use params parameter instead of building query string manually
  const { data, error } = await api.get<PagedWalletsResponse>(
    API_ENDPOINTS.wallets.base,
    { limit, offset }
  );
  
  if (error) {
    throw new Error(error.message);
  }
  
  // Convert backend wallets to frontend format
  return {
    wallets: data.items.map(toFrontendWallet),
    hasMore: data.has_more
  };
}

/**
 * Get a single wallet by chain type and address
 */
export async function getWallet(chainType: string, address: string): Promise<WalletFrontend> {
  const endpoint = API_ENDPOINTS.wallets.byId(chainType, address);
  const { data, error } = await api.get<Wallet>(endpoint);
  
  if (error) {
    throw new Error(error.message);
  }
  
  // Convert to frontend format
  return toFrontendWallet(data);
}

/**
 * Create a new wallet
 */
export async function createWallet(data: CreateWalletRequest): Promise<WalletFrontend> {
  const { data: createdWallet, error } = await api.post<Wallet>(
    API_ENDPOINTS.wallets.base,
    data
  );
  
  if (error) {
    throw new Error(error.message);
  }
  
  // Convert to frontend format
  return toFrontendWallet(createdWallet);
}

/**
 * Update an existing wallet
 */
export async function updateWallet(chainType: string, address: string, data: UpdateWalletRequest): Promise<WalletFrontend> {
  const endpoint = API_ENDPOINTS.wallets.byId(chainType, address);
  const { data: updatedWallet, error } = await api.put<Wallet>(endpoint, data);
  
  if (error) {
    throw new Error(error.message);
  }
  
  // Convert to frontend format
  return toFrontendWallet(updatedWallet);
}

/**
 * Delete a wallet
 */
export async function deleteWallet(chainType: string, address: string): Promise<void> {
  const endpoint = API_ENDPOINTS.wallets.byId(chainType, address);
  const { error } = await api.delete<void>(endpoint);
  
  if (error) {
    throw new Error(error.message);
  }
} 