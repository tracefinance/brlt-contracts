/**
 * API endpoint paths for the Vault0 backend
 */

export const API_ENDPOINTS = {
  // Wallet endpoints
  WALLETS: {
    BASE: '/wallets',
    BY_ADDRESS: (chainType: string, address: string) => `/wallets/${chainType}/${address}`,
    BALANCE: (chainType: string, address: string) => `/wallets/${chainType}/${address}/balance`,
    ACTIVATE_TOKEN: (chainType: string, address: string) => `/wallets/${chainType}/${address}/activate-token`,
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
    UPDATE: (address: string) => `/tokens/${address}`,
  },
  
  // Signer endpoints
  SIGNERS: {
    BASE: '/signers',
    BY_ID: (id: string) => `/signers/${id}`,
    BY_USER_ID: (userId: string) => `/signers/user/${userId}`,
    ADDRESSES: (signerId: string) => `/signers/${signerId}/addresses`,
    ADDRESS_BY_ID: (signerId: string, addressId: string) => `/signers/${signerId}/addresses/${addressId}`,
  },
  
  // User endpoints
  USERS: {
    BASE: '/users',
    BY_ID: (id: string) => `/users/${id}`,
    PROFILE: '/users/profile',
  },

  // Reference data endpoints
  REFERENCES: {
    CHAINS: '/references/chains',
  },
  
  // Keystore endpoints
  KEYS: {
    BASE: '/keys',
    BY_ID: (id: string) => `/keys/${id}`,
    IMPORT: '/keys/import',
    SIGN: (id: string) => `/keys/${id}/sign`,
  },
};
