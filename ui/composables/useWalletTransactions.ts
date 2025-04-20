import { computed, ref, watch } from 'vue'
import type { Ref } from 'vue'
import type { IPagedResponse, ITransaction } from '~/types'

/**
 * Composable for fetching transactions for a specific wallet and token, with token-based pagination.
 *
 * @param address - Reactive ref for the wallet address.
 * @param chainType - Reactive ref for the chain type.
 * @param limit - Reactive ref for the number of transactions per page.
 * @param nextToken - Reactive ref for the pagination token.
 * @param tokenAddress - Reactive ref for the token address (optional).
 * @returns Reactive state including transactions, pagination info, loading status, errors, refresh function, and initial load status.
 */
export default function (
  chainType: Ref<string | undefined>,
  address: Ref<string | undefined>,  
  tokenAddress: Ref<string | undefined>,
  limit: Ref<number>,
  nextToken: Ref<string | undefined>,  
) {
  const { $api } = useNuxtApp()

  const { 
    data: transactionsData, 
    status, 
    error, 
    refresh 
  } = useAsyncData<IPagedResponse<ITransaction> | null>(
    'walletTransactions',
    async () => {
      const addressValue = address.value
      const chainTypeValue = chainType.value
      const tokenAddressValue = tokenAddress.value
      
      if (addressValue && chainTypeValue && tokenAddressValue) {
        return await $api.transaction.getWalletTransactions(
          addressValue,
          chainTypeValue,
          tokenAddressValue,
          limit.value,
          nextToken.value
        )
      }
      return null
    },
    {
      watch: [address, chainType, tokenAddress, limit, nextToken],
      default: () => null
    }
  )

  const transactions = computed<ITransaction[]>(() => transactionsData.value?.items || [])
  const nextPageToken = computed<string | undefined>(() => transactionsData.value?.nextToken)
  const isLoading = computed<boolean>(() => status.value === 'pending')

  // --- Track Initial Load State ---
  const hasInitiallyLoaded = ref(false)
  watch(status, (currentStatus, prevStatus) => {
    // Check if the status transitioned from pending to something else for the first time
    if (prevStatus === 'pending' && currentStatus !== 'pending' && !hasInitiallyLoaded.value) {
      hasInitiallyLoaded.value = true
    }
    // Consider resetting if inputs change causing a new load? For now, keep it simple.
  }, { immediate: true }) // Use immediate to check initial status
  // --- End Track Initial Load State ---

  return {
    transactions,
    nextPageToken,
    isLoading,
    error,
    refresh,
    hasInitiallyLoaded,
  }
} 