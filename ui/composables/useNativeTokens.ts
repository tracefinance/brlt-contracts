import { computed } from 'vue';
import { useNuxtApp, useAsyncData } from '#app';
import type { IToken } from '~/types';

/**
 * Composable for fetching native tokens reference data for all supported blockchains.
 */
export default function () {
  const { $api } = useNuxtApp();

  const { data: tokensData, status, error, refresh } = useAsyncData<IToken[]>(
    'nativeTokensReference',
    async () => await $api.reference.listNativeTokens(),
    {
      default: () => [],
    }
  );

  const nativeTokens = computed(() => tokensData.value || []);
  const isLoading = computed(() => status.value === 'pending');
  const hasError = computed(() => status.value === 'error');

  return {
    nativeTokens,
    isLoading,
    hasError,
    error,
    refresh,
  };
} 