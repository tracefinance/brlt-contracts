import type { Ref } from 'vue';
import type { 
  Transaction, 
  PagedTransactions,
  SyncTransactionsResponse
} from '~/types/transaction';

/**
 * Composable for transaction-related functionality
 */
export function useTransactions() {
  // Get the API service from the Nuxt plugin
  const { $api } = useNuxtApp();
  
  // Reactive state
  const transactions: Ref<Transaction[]> = ref([]);
  const currentTransaction: Ref<Transaction | null> = ref(null);
  const isLoading: Ref<boolean> = ref(false);
  const error: Ref<string | null> = ref(null);
  
  /**
   * Gets a transaction by its hash
   */
  async function getTransaction(hash: string) {
    isLoading.value = true;
    error.value = null;
    
    try {
      currentTransaction.value = await $api.transaction.getTransaction(hash);
      return currentTransaction.value;
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to get transaction';
      return null;
    } finally {
      isLoading.value = false;
    }
  }
  
  /**
   * Gets transactions for a specific wallet
   */
  async function getTransactionsByAddress(
    chainType: string,
    address: string,
    limit: number = 10,
    offset: number = 0,
    tokenAddress?: string
  ) {
    isLoading.value = true;
    error.value = null;
    
    try {
      const result = await $api.transaction.getTransactionsByAddress(
        chainType, 
        address, 
        limit, 
        offset, 
        tokenAddress
      );
      
      transactions.value = result.items;
      return result;
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to get transactions';
      return null;
    } finally {
      isLoading.value = false;
    }
  }
  
  /**
   * Syncs transactions for a specific wallet
   */
  async function syncTransactions(chainType: string, address: string) {
    isLoading.value = true;
    error.value = null;
    
    try {
      const result = await $api.transaction.syncTransactions(chainType, address);
      
      // Refresh the transactions list if we have transactions for this wallet
      if (transactions.value.length > 0) {
        await getTransactionsByAddress(chainType, address);
      }
      
      return result;
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to sync transactions';
      return null;
    } finally {
      isLoading.value = false;
    }
  }
  
  /**
   * Filters transactions based on various criteria
   */
  async function filterTransactions(options: {
    chainType?: string;
    address?: string;
    tokenAddress?: string;
    status?: string;
    limit?: number;
    offset?: number;
  }) {
    isLoading.value = true;
    error.value = null;
    
    try {
      const result = await $api.transaction.filterTransactions(options);
      transactions.value = result.items;
      return result;
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to filter transactions';
      return null;
    } finally {
      isLoading.value = false;
    }
  }
  
  return {
    // State
    transactions,
    currentTransaction,
    isLoading,
    error,
    
    // Methods
    getTransaction,
    getTransactionsByAddress,
    syncTransactions,
    filterTransactions
  };
} 