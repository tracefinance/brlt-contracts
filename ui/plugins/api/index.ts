import { defineNuxtPlugin } from '#app';
import { ApiClient } from './client';
import { WalletClient } from './wallet';
import { TokenClient } from './token';
import { TransactionClient } from './transaction';

/**
 * API service that provides access to all API clients
 */
export class ApiService {
  client: ApiClient;
  wallet: WalletClient;
  token: TokenClient;
  transaction: TransactionClient;

  constructor(baseUrl: string) {
    this.client = new ApiClient();
    this.client.setBaseUrl(baseUrl);
    this.wallet = new WalletClient(this.client);
    this.token = new TokenClient(this.client);
    this.transaction = new TransactionClient(this.client);
  }

  /**
   * Set the authentication token for all API requests
   */
  setToken(token: string) {
    this.client.setToken(token);
  }
}

export default defineNuxtPlugin((nuxtApp) => {
  // Create the API client
  const apiClient = new ApiClient();
  
  // Get the API base URL from runtime config
  const config = useRuntimeConfig();
  const apiBase = config.public.apiBase as string || 'http://localhost:8080/api/v1';
  apiClient.setBaseUrl(apiBase);
  
  // Create service clients
  const walletClient = new WalletClient(apiClient);
  const tokenClient = new TokenClient(apiClient);
  const transactionClient = new TransactionClient(apiClient);
  
  // Provide API services to the application
  return {
    provide: {
      api: {
        wallet: walletClient,
        token: tokenClient,
        transaction: transactionClient
      }
    }
  };
});