import type { IChain, IToken } from '~/types';
import { Chain, Token } from '~/types';
import type { ApiClient } from './client';
import { API_ENDPOINTS } from './endpoints';

/**
 * Client for interacting with reference data API endpoints
 */
export class ReferenceClient {
  private client: ApiClient;

  /**
   * Creates a new reference client
   * @param client API client instance
   */
  constructor(client: ApiClient) {
    this.client = client;
  }

  /**
   * Lists supported blockchain chains
   * @returns Array of chain references
   */
  async listChains(): Promise<IChain[]> {
    const data = await this.client.get<any[]>(API_ENDPOINTS.REFERENCES.CHAINS);
    return Chain.fromJsonArray(data);
  }

  /**
   * Lists native tokens for all supported blockchains
   * @returns Array of native tokens
   */
  async listNativeTokens(): Promise<IToken[]> {
    const data = await this.client.get<any[]>(API_ENDPOINTS.REFERENCES.NATIVE_TOKENS);
    return Token.fromJsonArray(data);
  }
} 