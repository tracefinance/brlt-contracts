import { defineNuxtPlugin } from '#app';
import { ApiClient } from './client';
import { WalletClient } from './wallet';
import { TokenClient } from './token';
import { TransactionClient } from './transaction';
import { ReferenceClient } from './reference';
import { SignerClient } from './signer';
import { UserClient } from './user';
import { KeyClient } from './key';

/**
 * API service that provides access to all API clients
 */
export class ApiService {
  client: ApiClient;
  wallet: WalletClient;
  token: TokenClient;
  transaction: TransactionClient;
  reference: ReferenceClient;
  signer: SignerClient;
  user: UserClient;
  key: KeyClient;

  constructor(baseUrl: string) {
    this.client = new ApiClient();
    this.client.setBaseUrl(baseUrl);
    this.wallet = new WalletClient(this.client);
    this.token = new TokenClient(this.client);
    this.transaction = new TransactionClient(this.client);
    this.reference = new ReferenceClient(this.client);
    this.signer = new SignerClient(this.client);
    this.user = new UserClient(this.client);
    this.key = new KeyClient(this.client);
  }

  /**
   * Set the authentication token for all API requests
   */
  setToken(token: string) {
    this.client.setToken(token);
  }
}

/**
 * Plugin that provides API client services to components and composables
 */
export default defineNuxtPlugin(() => {
  const apiClient = new ApiClient();
  
  const config = useRuntimeConfig();
  const apiBase = config.public.apiBase as string || 'http://localhost:8080/api/v1';
  apiClient.setBaseUrl(apiBase);
  
  const walletClient = new WalletClient(apiClient);
  const tokenClient = new TokenClient(apiClient);
  const transactionClient = new TransactionClient(apiClient);
  const referenceClient = new ReferenceClient(apiClient);
  const signerClient = new SignerClient(apiClient);
  const userClient = new UserClient(apiClient);
  const keyClient = new KeyClient(apiClient);
  
  return {
    provide: {
      api: {
        wallet: walletClient,
        token: tokenClient,
        transaction: transactionClient,
        user: userClient,
        signer: signerClient,
        reference: referenceClient,
        key: keyClient
      }
    }
  };
});
