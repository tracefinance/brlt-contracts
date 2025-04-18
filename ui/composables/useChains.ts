import { computed } from 'vue';
import { useNuxtApp, useAsyncData } from '#app';
import type { IChain } from '~/types';

/**
 * Composable for fetching supported blockchain chains reference data.
 */
export default function () {
  const { $api } = useNuxtApp();

  const { data: chainsData, status, error, refresh } = useAsyncData<IChain[]>(
    'chainsReference',
    async () => await $api.reference.listChains(),
    {
      default: () => [],
    }
  );

  const chains = computed(() => chainsData.value || []);
  const isLoading = computed(() => status.value === 'pending');
  const hasError = computed(() => status.value === 'error');

  return {
    chains,
    isLoading,
    hasError,
    error,
    refresh,
  };
} 