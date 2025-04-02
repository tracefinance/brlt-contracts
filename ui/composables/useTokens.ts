import type { Ref } from 'vue';
import type { 
  IToken,
  IPagedTokens,
  IAddTokenRequest
} from '~/types';

/**
 * Composable for token-related functionality
 */
export function useTokens() {
  // Get the API service from the Nuxt plugin
  const { $api } = useNuxtApp();
  
  // Reactive state
  const pagedData = ref<IPagedTokens>({
    items: [],
    limit: 10,
    offset: 0,
    hasMore: false
  });
  const currentToken: Ref<IToken | null> = ref(null);
  const isLoading: Ref<boolean> = ref(false);
  const error: Ref<string | null> = ref(null);
  
  // Computed properties for easier access
  const tokens = computed(() => pagedData.value.items);
  const hasMoreTokens = computed(() => pagedData.value.hasMore);
  
  /**
   * Reset paged data to initial state
   */
  function resetPagedData() {
    pagedData.value = {
      items: [],
      limit: 10,
      offset: 0,
      hasMore: false
    };
  }
  
  /**
   * Loads all tokens with pagination
   */
  async function loadTokens(chainType?: string, tokenType?: string, limit = 10, offset = 0) {
    isLoading.value = true;
    error.value = null;
    
    try {
      const result: IPagedTokens = await $api.token.listTokens(
        chainType,
        tokenType,
        limit, 
        offset
      );
      
      pagedData.value = result;
      
      return result;
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to load tokens';
      resetPagedData();
      return null;
    } finally {
      isLoading.value = false;
    }
  }
  
  /**
   * Load a specific token by address and chain type
   */
  async function loadToken(chainType: string, address: string) {
    isLoading.value = true;
    error.value = null;
    
    try {
      const token = await $api.token.getToken(chainType, address);
      currentToken.value = token;
      return token;
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to load token';
      currentToken.value = null;
      return null;
    } finally {
      isLoading.value = false;
    }
  }
  
  /**
   * Verify a token by its address
   */
  async function verifyToken(address: string) {
    isLoading.value = true;
    error.value = null;
    
    try {
      return await $api.token.verifyToken(address);
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
  async function addToken(request: IAddTokenRequest) {
    isLoading.value = true;
    error.value = null;
    
    try {
      const newToken = await $api.token.addToken(request);
      
      // Refresh tokens list if we're viewing tokens
      if (pagedData.value.items.length > 0) {
        await loadTokens();
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
  async function deleteToken(address: string) {
    isLoading.value = true;
    error.value = null;
    
    try {
      await $api.token.deleteToken(address);
      
      // Remove from list if in memory
      if (pagedData.value.items.length > 0) {
        pagedData.value = {
          ...pagedData.value,
          items: pagedData.value.items.filter(t => t.address !== address)
        };
      }
      
      return true;
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to delete token';
      return false;
    } finally {
      isLoading.value = false;
    }
  }
  
  /**
   * Get tokens for a specific chain type
   */
  async function getTokensByChainType(chainType: string, limit = 10, offset = 0) {
    // Reuse listTokens with chainType filter
    return loadTokens(chainType, undefined, limit, offset);
  }
  
  return {
    // State
    tokens,
    currentToken,
    isLoading,
    error,
    hasMoreTokens,
    
    // Methods
    loadTokens,
    loadToken,
    verifyToken,
    addToken,
    deleteToken,
    getTokensByChainType
  };
} 