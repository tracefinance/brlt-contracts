/**
 * API module index file
 * Centralizes exports from all API-related modules
 */

// Re-export the core API client
export * from './client';

// Re-export domain-specific API services 
export * as walletApi from './wallet';

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
 * import { walletApi } from '@/lib/api';
 * 
 * // Then use the wallet API functions
 * const wallets = await walletApi.getWallets();
 */ 