import type { Ref } from 'vue';
import type { 
  Wallet, 
  PagedWallets,
  TokenBalanceResponse 
} from '~/types/wallet';
import { 
  CreateWalletRequest, 
  UpdateWalletRequest 
} from '~/types/wallet';

/**
 * Composable for wallet-related functionality
 */
export function useWallets() {
  // Get the API service from the Nuxt plugin
  const { $api } = useNuxtApp();
  
  // Reactive state
  const wallets: Ref<Wallet[]> = ref([]);
  const currentWallet: Ref<Wallet | null> = ref(null);
  const balances: Ref<TokenBalanceResponse[]> = ref([]);
  const isLoading: Ref<boolean> = ref(false);
  const error: Ref<string | null> = ref(null);
  
  /**
   * Loads wallets with pagination
   */
  async function loadWallets(limit = 10, offset = 0) {
    isLoading.value = true;
    error.value = null;
    
    try {
      const result: PagedWallets = await $api.wallet.listWallets(limit, offset);
      wallets.value = result.items;
      return result;
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to load wallets';
      return null;
    } finally {
      isLoading.value = false;
    }
  }
  
  /**
   * Load a specific wallet by chain type and address
   */
  async function loadWallet(chainType: string, address: string) {
    isLoading.value = true;
    error.value = null;
    
    try {
      currentWallet.value = await $api.wallet.getWallet(chainType, address);
      return currentWallet.value;
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to load wallet';
      return null;
    } finally {
      isLoading.value = false;
    }
  }
  
  /**
   * Create a new wallet
   */
  async function createWallet(chainType: string, name: string, tags?: Record<string, string>) {
    isLoading.value = true;
    error.value = null;
    
    try {
      const request = new CreateWalletRequest(chainType, name, tags);
      const newWallet = await $api.wallet.createWallet(request);
      
      // Add to local wallet list if exists
      if (wallets.value.length > 0) {
        wallets.value = [...wallets.value, newWallet];
      }
      
      return newWallet;
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to create wallet';
      return null;
    } finally {
      isLoading.value = false;
    }
  }
  
  /**
   * Update an existing wallet
   */
  async function updateWallet(chainType: string, address: string, name: string, tags?: Record<string, string>) {
    isLoading.value = true;
    error.value = null;
    
    try {
      const request = new UpdateWalletRequest(name, tags);
      const updatedWallet = await $api.wallet.updateWallet(chainType, address, request);
      
      // Update in the list if present
      if (wallets.value.length > 0) {
        wallets.value = wallets.value.map(w => 
          (w.address === address && w.chainType === chainType) ? updatedWallet : w
        );
      }
      
      // Update current wallet if it's the same one
      if (currentWallet.value?.address === address && currentWallet.value?.chainType === chainType) {
        currentWallet.value = updatedWallet;
      }
      
      return updatedWallet;
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to update wallet';
      return null;
    } finally {
      isLoading.value = false;
    }
  }
  
  /**
   * Delete a wallet
   */
  async function deleteWallet(chainType: string, address: string) {
    isLoading.value = true;
    error.value = null;
    
    try {
      await $api.wallet.deleteWallet(chainType, address);
      
      // Remove from list if present
      if (wallets.value.length > 0) {
        wallets.value = wallets.value.filter(w => 
          !(w.address === address && w.chainType === chainType)
        );
      }
      
      // Clear current wallet if it's the same one
      if (currentWallet.value?.address === address && currentWallet.value?.chainType === chainType) {
        currentWallet.value = null;
      }
      
      return true;
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to delete wallet';
      return false;
    } finally {
      isLoading.value = false;
    }
  }
  
  /**
   * Load wallet balances
   */
  async function loadWalletBalances(chainType: string, address: string) {
    isLoading.value = true;
    error.value = null;
    
    try {
      balances.value = await $api.wallet.getWalletBalance(chainType, address);
      return balances.value;
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to load wallet balances';
      balances.value = [];
      return null;
    } finally {
      isLoading.value = false;
    }
  }
  
  return {
    // State
    wallets,
    currentWallet,
    balances,
    isLoading,
    error,
    
    // Methods
    loadWallets,
    loadWallet,
    createWallet,
    updateWallet,
    deleteWallet,
    loadWalletBalances
  };
} 