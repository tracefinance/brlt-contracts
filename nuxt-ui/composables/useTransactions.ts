import type { Ref } from 'vue';
import type { 
  Transaction, 
  TransactionListResponse
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
   * Loads all transactions with pagination
   */
  async function loadTransactions(limit = 10, offset = 0) {
    isLoading.value = true;
    error.value = null;
    
    try {
      const result: TransactionListResponse = await $api.transaction.listTransactions(limit, offset);
      transactions.value = result.items;
      return result;
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to load transactions';
      return null;
    } finally {
      isLoading.value = false;
    }
  }
  
  /**
   * Load a specific transaction by ID
   */
  async function getTransaction(id: string) {
    isLoading.value = true;
    error.value = null;
    
    try {
      currentTransaction.value = await $api.transaction.getTransaction(id);
      return currentTransaction.value;
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to load transaction';
      return null;
    } finally {
      isLoading.value = false;
    }
  }
  
  /**
   * Get transactions for a specific wallet
   */
  async function getWalletTransactions(
    chainType: string,
    address: string,
    limit = 10,
    offset = 0,
    tokenAddress?: string
  ) {
    isLoading.value = true;
    error.value = null;
    
    try {
      const result = await $api.transaction.getWalletTransactions(
        address,
        chainType,
        limit,
        offset,
        tokenAddress
      );
      
      transactions.value = result.items;
      return result;
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to load wallet transactions';
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
        await getWalletTransactions(chainType, address);
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
    loadTransactions,
    getTransaction,
    getTransactionsByAddress: getWalletTransactions,
    syncTransactions,
    filterTransactions
  };
} 