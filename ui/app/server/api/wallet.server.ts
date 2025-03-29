import { 
  apiGet, 
  apiPost, 
  apiPut, 
  apiDelete 
} from './client';
import { 
  Wallet, 
  CreateWalletRequest, 
  UpdateWalletRequest, 
  PagedWallets 
} from '~/models/wallet';
import { 
  Token 
} from '~/models/token';
import { API_ENDPOINTS } from './endpoints';

/**
 * Response type for token balance endpoints
 */
export interface TokenBalanceResponse {
  token: Token;
  balance: string;
  updatedAt: string;
}

/**
 * Creates a new wallet
 * @param request Wallet creation request data
 * @param token Authentication token
 * @returns Created wallet
 */
export async function createWallet(
  request: CreateWalletRequest,
  token: string
): Promise<Wallet> {
  const data = await apiPost<any>(API_ENDPOINTS.WALLETS.BASE, request.toJson(), token);
  return Wallet.fromJson(data);
}

/**
 * Gets a wallet by its chain type and address
 * @param chainType Blockchain network type (e.g., ethereum, bitcoin)
 * @param address Wallet address
 * @param token Authentication token
 * @returns Wallet details
 */
export async function getWallet(
  chainType: string,
  address: string,
  token: string
): Promise<Wallet> {
  const endpoint = API_ENDPOINTS.WALLETS.BY_ADDRESS(address, chainType);
  const data = await apiGet<any>(endpoint, undefined, token);
  return Wallet.fromJson(data);
}

/**
 * Updates a wallet's properties
 * @param chainType Blockchain network type
 * @param address Wallet address
 * @param request Wallet update request data
 * @param token Authentication token
 * @returns Updated wallet
 */
export async function updateWallet(
  chainType: string,
  address: string,
  request: UpdateWalletRequest,
  token: string
): Promise<Wallet> {
  const endpoint = API_ENDPOINTS.WALLETS.BY_ADDRESS(address, chainType);
  const data = await apiPut<any>(endpoint, request.toJson(), token);
  return Wallet.fromJson(data);
}

/**
 * Deletes a wallet
 * @param chainType Blockchain network type
 * @param address Wallet address
 * @param token Authentication token
 */
export async function deleteWallet(
  chainType: string,
  address: string,
  token: string
): Promise<void> {
  const endpoint = API_ENDPOINTS.WALLETS.BY_ADDRESS(address, chainType);
  await apiDelete<void>(endpoint, token);
}

/**
 * Lists wallets with pagination
 * @param limit Maximum number of wallets to return (default: 10)
 * @param offset Number of wallets to skip for pagination (default: 0)
 * @param token Authentication token
 * @returns Paginated list of wallets
 */
export async function listWallets(
  limit: number = 10,
  offset: number = 0,
  token: string
): Promise<PagedWallets> {
  const params = {
    limit,
    offset
  };
  
  const data = await apiGet<any>(API_ENDPOINTS.WALLETS.BASE, params, token);
  return PagedWallets.fromJson(data);
}

/**
 * Gets the balance of a wallet
 * @param chainType Blockchain network type
 * @param address Wallet address
 * @param token Authentication token
 * @returns Array of token balances
 */
export async function getWalletBalance(
  chainType: string,
  address: string,
  token: string
): Promise<TokenBalanceResponse[]> {
  const endpoint = API_ENDPOINTS.WALLETS.BALANCE(address, chainType);
  const data = await apiGet<any[]>(endpoint, undefined, token);
  
  return data.map(item => ({
    token: Token.fromJson(item.token),
    balance: item.balance,
    updatedAt: item.updated_at
  }));
} 