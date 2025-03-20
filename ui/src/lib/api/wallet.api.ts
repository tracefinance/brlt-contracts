import { BaseApi } from './base.api';
import { 
  Wallet, 
  PagedWallets, 
  CreateWalletRequest, 
  UpdateWalletRequest 
} from '@/types/models/wallet.model';

/**
 * API service for wallet-related operations
 */
export class WalletApi extends BaseApi {
  private static readonly WALLETS_PATH = '/wallets';

  /**
   * Fetches a list of wallets
   */
  static async getWallets(
    limit: number = 10, 
    offset: number = 0
  ): Promise<PagedWallets> {
    const queryParams = { limit, offset };
    const data = await this.get<any>(this.WALLETS_PATH, queryParams);
    return PagedWallets.fromJson(data);
  }

  /**
   * Fetches a specific wallet by chainType and address
   */
  static async getWallet(
    chainType: string, 
    address: string
  ): Promise<Wallet> {
    const path = `${this.WALLETS_PATH}/${chainType}/${address}`;
    const data = await this.get<any>(path);
    return Wallet.fromJson(data);
  }

  /**
   * Creates a new wallet
   */
  static async createWallet(request: CreateWalletRequest): Promise<Wallet> {
    const data = await this.post<any>(this.WALLETS_PATH, request.toJson());
    return Wallet.fromJson(data);
  }

  /**
   * Updates a wallet
   */
  static async updateWallet(
    chainType: string,
    address: string,
    request: UpdateWalletRequest
  ): Promise<Wallet> {
    const path = `${this.WALLETS_PATH}/${chainType}/${address}`;
    const data = await this.put<any>(path, request.toJson());
    return Wallet.fromJson(data);
  }

  /**
   * Deletes a wallet
   */
  static async deleteWallet(
    chainType: string,
    address: string
  ): Promise<void> {
    const path = `${this.WALLETS_PATH}/${chainType}/${address}`;
    await this.delete<any>(path);
  }
} 