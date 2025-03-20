/**
 * API module index file
 * Centralizes exports from all API-related modules
 */

// Export the base API class for extending
export { BaseApi } from './base.api';

// Export API services
export { WalletApi } from './wallet.api';
export { TransactionApi } from './transaction.api';

// Additional API services can be exported here as the application grows

// Add additional API services as they are created
// export * as userApi from './user';
// export * as authApi from './auth';
// etc...

/**
 * This file serves as the main entry point for all API-related functionality.
 * By importing from '@/lib/api' instead of individual files, we maintain a cleaner
 * dependency structure and make it easier to refactor the API layer in the future.
 * 
 * Example usage:
 * 
 * import { WalletApi, TransactionApi } from '@/lib/api';
 * 
 * // Then use the wallet API functions
 * const wallets = await WalletApi.getWallets();
 * 
 * // Or transaction API functions
 * const transactions = await TransactionApi.getTransactions('ethereum', '0x123...');
 */ 