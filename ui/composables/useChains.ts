import { ref, computed } from 'vue';
import { useNuxtApp, useAsyncData } from '#app';
import type { IChain } from '~/types';

/**
 * Composable for fetching supported blockchain chains reference data.
 */
export default function () {
  const { $api } = useNuxtApp();

  // Use useAsyncData for efficient data fetching
  const { data: chainsData, status, error, refresh } = useAsyncData<IChain[]>(
    'chainsReference', // Unique key for this data fetch
    () => $api.reference.listChains(), // API call function
    {
      default: () => [], // Provide an empty array as default
      // Consider adding `server: false` if this data is purely client-side
      // server: false, 
    }
  );

  // Computed properties for easier access in components
  const chains = computed(() => chainsData.value || []);
  const isLoading = computed(() => status.value === 'pending');
  const hasError = computed(() => status.value === 'error');

  return {
    chains,      // The reactive list of chains
    isLoading,   // Reactive boolean indicating if data is loading
    hasError,    // Reactive boolean indicating if an error occurred
    error,       // The error object if an error occurred
    refresh,     // Function to manually trigger a refetch
  };
} 