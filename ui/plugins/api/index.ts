import { defineNuxtPlugin } from '#app';
import { ApiClient } from './client';
import { WalletClient } from './wallet';
import { TokenClient } from './token';
import { TransactionClient } from './transaction';
import { ReferenceClient } from './reference';
import { SignerClient } from './signer';
import { UserClient } from './user';
import { KeyClient } from './key';
import { VaultClient } from './vault';

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
  const vaultClient = new VaultClient(apiClient);
  
  return {
    provide: {
      api: {
        wallet: walletClient,
        token: tokenClient,
        transaction: transactionClient,
        user: userClient,
        signer: signerClient,
        reference: referenceClient,
        key: keyClient,
        vault: vaultClient
      }
    }
  };
});
