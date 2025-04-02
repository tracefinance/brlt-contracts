import type { Ref } from 'vue';
import { 
  Transaction, 
  PagedTransactions
} from '~/types/transaction';

/**
 * Composable for transaction-related functionality
 */
export function useTransactions() {
  // Get the API service from the Nuxt plugin
  const { $api } = useNuxtApp();
  
  // Reactive state
  const pagedData = ref<PagedTransactions>(new PagedTransactions({
    items: [],
    limit: 10,
    offset: 0,
    hasMore: false
  }));
  const currentTransaction: Ref<Transaction | null> = ref(null);
  const isLoading: Ref<boolean> = ref(false);
  const error: Ref<string | null> = ref(null);
  
  // Computed properties for easier access
  const transactions = computed(() => pagedData.value.items);
  const hasMoreTransactions = computed(() => pagedData.value.hasMore);
  
  /**
   * Loads all transactions with pagination
   */
  async function loadTransactions(limit = 10, offset = 0) {
    isLoading.value = true;
    error.value = null;
    
    try {
      const result: PagedTransactions = await $api.transaction.listTransactions(limit, offset);      
      pagedData.value = result;    
      return result;
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to load transactions';
      resetPagedData();
      return null;
    } finally {
      isLoading.value = false;
    }
  }
  
  /**
   * Reset paged data to initial state
   */
  function resetPagedData() {
    pagedData.value = new PagedTransactions({
      items: [],
      limit: 10,
      offset: 0,
      hasMore: false
    });
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
      
      pagedData.value = result;
      
      return result;
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to load wallet transactions';
      resetPagedData();
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
      if (pagedData.value.items.length > 0) {
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
    const { limit = 10, offset = 0 } = options;
    isLoading.value = true;
    error.value = null;
    
    try {
      const result = await $api.transaction.filterTransactions(options);
      
      pagedData.value = result;
      
      return result;
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to filter transactions';
      resetPagedData();
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
    hasMoreTransactions,
    
    // Methods
    loadTransactions,
    getTransaction,
    getTransactionsByAddress: getWalletTransactions,
    syncTransactions,
    filterTransactions
  };
} 