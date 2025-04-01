import type { Ref } from 'vue';
import type { 
  Token, 
  TokenListResponse
} from '~/types/token';
import { AddTokenRequest } from '~/types/token';

/**
 * Composable for token-related functionality
 */
export function useTokens() {
  // Get the API service from the Nuxt plugin
  const { $api } = useNuxtApp();
  
  // Reactive state
  const tokens: Ref<Token[]> = ref([]);
  const currentToken: Ref<Token | null> = ref(null);
  const isLoading: Ref<boolean> = ref(false);
  const error: Ref<string | null> = ref(null);
  
  /**
   * Loads tokens with pagination and optional filtering
   */
  async function loadTokens(
    chainType?: string, 
    tokenType?: string, 
    limit = 10, 
    offset = 0
  ) {
    isLoading.value = true;
    error.value = null;
    
    try {
      const result: TokenListResponse = await $api.token.listTokens(
        chainType,
        tokenType,
        limit, 
        offset
      );
      tokens.value = result.items;
      return result;
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to load tokens';
      return null;
    } finally {
      isLoading.value = false;
    }
  }
  
  /**
   * Load a specific token by chain type and address
   */
  async function loadToken(chainType: string, address: string) {
    isLoading.value = true;
    error.value = null;
    
    try {
      currentToken.value = await $api.token.getToken(chainType, address);
      return currentToken.value;
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to load token';
      return null;
    } finally {
      isLoading.value = false;
    }
  }
  
  /**
   * Load a specific token by ID
   */
  async function loadTokenById(id: string) {
    isLoading.value = true;
    error.value = null;
    
    try {
      currentToken.value = await $api.token.getTokenById(id);
      return currentToken.value;
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to load token';
      return null;
    } finally {
      isLoading.value = false;
    }
  }
  
  /**
   * Verify a token by its address and chain
   */
  async function verifyToken(address: string, chainType: string) {
    isLoading.value = true;
    error.value = null;
    
    try {
      const verifiedToken = await $api.token.verifyToken(address, chainType);
      return verifiedToken;
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to verify token';
      return null;
    } finally {
      isLoading.value = false;
    }
  }
  
  /**
   * Add a new token
   */
  async function addToken(
    address: string,
    chainType: string,
    tokenType: string,
    name: string,
    symbol: string,
    decimals: number,
    logo?: string
  ) {
    isLoading.value = true;
    error.value = null;
    
    try {
      const request = new AddTokenRequest(
        address,
        chainType,
        tokenType,
        name,
        symbol,
        decimals,
        logo
      );
      
      const newToken = await $api.token.addToken(request);
      
      // Add to local token list if exists
      if (tokens.value.length > 0) {
        tokens.value = [...tokens.value, newToken];
      }
      
      return newToken;
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to add token';
      return null;
    } finally {
      isLoading.value = false;
    }
  }
  
  /**
   * Delete a token
   */
  async function deleteToken(chainType: string, address: string) {
    isLoading.value = true;
    error.value = null;
    
    try {
      await $api.token.deleteToken(chainType, address);
      
      // Remove from list if present
      if (tokens.value.length > 0) {
        tokens.value = tokens.value.filter(t => 
          !(t.address === address && t.chainType === chainType)
        );
      }
      
      // Clear current token if it's the same one
      if (currentToken.value?.address === address && currentToken.value?.chainType === chainType) {
        currentToken.value = null;
      }
      
      return true;
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to delete token';
      return false;
    } finally {
      isLoading.value = false;
    }
  }
  
  return {
    // State
    tokens,
    currentToken,
    isLoading,
    error,
    
    // Methods
    loadTokens,
    loadToken,
    loadTokenById,
    verifyToken,
    addToken,
    deleteToken
  };
} 