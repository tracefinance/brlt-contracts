import { ApiClient, createApiClient } from './client';
import { WalletClient } from './wallet';
import { TransactionClient } from './transaction';

/**
 * API service that provides access to all API clients
 */
export class ApiService {
  client: ApiClient;
  wallet: WalletClient;
  transaction: TransactionClient;

  constructor(baseUrl: string) {
    this.client = createApiClient(undefined, baseUrl);
    this.wallet = new WalletClient(this.client);
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
  // Get API URL from runtime config
  const config = useRuntimeConfig();
  const baseUrl = (config.public?.apiUrl as string) || 'http://localhost:8080/api/v1';
  
  // Create API service
  const apiService = new ApiService(baseUrl);
  
  // Provide the API service to the app
  return {
    provide: {
      api: apiService
    }
  };
});