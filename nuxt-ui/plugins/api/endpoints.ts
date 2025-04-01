/**
 * API endpoint paths for the Vault0 backend
 */

export const API_ENDPOINTS = {
  // Wallet endpoints
  WALLETS: {
    BASE: '/wallets',
    BY_ADDRESS: (chainType: string, address: string) => `/wallets/${chainType}/${address}`,
    BALANCE: (chainType: string, address: string) => `/wallets/${chainType}/${address}/balance`,
  },
  
  // Auth endpoints
  AUTH: {
    LOGIN: '/auth/login',
    LOGOUT: '/auth/logout',
    ME: '/auth/me',
  },
  
  // Transaction endpoints
  TRANSACTIONS: {
    BASE: '/transactions',
    BY_ID: (id: string) => `/transactions/${id}`,
    BY_WALLET: (chainType: string, address: string) => `/wallets/${chainType}/${address}/transactions`,
  },
  
  // Token endpoints
  TOKENS: {
    BASE: '/tokens',
    BY_ADDRESS: (chainType: string, address: string) => `/tokens/${chainType}/${address}`,
    DELETE: (address: string) => `/tokens/${address}`,
    VERIFY: (address: string) => `/tokens/verify/${address}`,
  },
  
  // Signer endpoints
  SIGNERS: {
    BASE: '/signers',
    BY_ID: (id: string) => `/signers/${id}`,
  },
  
  // User endpoints
  USERS: {
    BASE: '/users',
    BY_ID: (id: string) => `/users/${id}`,
    PROFILE: '/users/profile',
  },
};