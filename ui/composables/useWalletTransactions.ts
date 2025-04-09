import { computed } from 'vue'
import type { Ref } from 'vue'
import type { IPagedResponse, ITransaction } from '~/types'

/**
 * Composable for fetching transactions for a specific wallet and token, with pagination.
 *
 * @param address - Reactive ref for the wallet address.
 * @param chainType - Reactive ref for the chain type.
 * @param limit - Reactive ref for the number of transactions per page.
 * @param offset - Reactive ref for the starting offset.
 * @param tokenAddress - Reactive ref for the token address (optional).
 * @returns Reactive state including transactions, pagination info, loading status, errors, and refresh function.
 */
export default function (
  chainType: Ref<string | undefined>,
  address: Ref<string | undefined>,  
  tokenAddress: Ref<string | undefined>,
  limit: Ref<number>,
  offset: Ref<number>,  
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
          offset.value
        )
      }
      return null
    },
    {
      watch: [address, chainType, tokenAddress, limit, offset],
      default: () => null
    }
  )

  const transactions = computed<ITransaction[]>(() => transactionsData.value?.items || [])
  const hasMore = computed<boolean>(() => transactionsData.value?.hasMore || false)
  const isLoading = computed<boolean>(() => status.value === 'pending')

  return {
    transactions,
    hasMore,
    isLoading,
    error,
    refresh,
  }
} 