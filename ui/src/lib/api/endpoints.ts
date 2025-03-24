/**
 * API endpoint paths for the Vault0 backend
 */

export const API_ENDPOINTS = {
  // Wallet endpoints
  WALLETS: {
    BASE: '/wallets',
    BY_ADDRESS: (address: string, chainType: string) => `/wallets/${address}/${chainType}`,
    BALANCE: (address: string, chainType: string) => `/wallets/${address}/${chainType}/balance`,
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
    BY_WALLET: (address: string, chainType: string) => `/wallets/${address}/${chainType}/transactions`,
  },
  
  // Token endpoints
  TOKENS: {
    BASE: '/tokens',
    BY_ADDRESS: (address: string) => `/tokens/${address}`,
    VERIFY: (address: string) => `/tokens/${address}`,
  },
}; 