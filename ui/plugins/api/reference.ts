import type { IChain } from '~/types';
import { Chain } from '~/types';
import { ApiClient } from './client';
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
} 