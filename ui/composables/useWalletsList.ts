import { computed } from 'vue'
import type { Ref } from 'vue'
import type { IPagedResponse, IWallet } from '~/types'

/**
 * Composable for fetching a list of wallets with token-based pagination.
 *
 * @param limit - Reactive ref for the number of wallets per page.
 * @param nextToken - Reactive ref for the pagination token.
 * @returns Reactive state including wallets, loading status, errors, and refresh function.
 */
export default function (limit: Ref<number>, nextToken: Ref<string | undefined>) {
  const { $api } = useNuxtApp()

  const { 
    data: walletsData, 
    status,
    error, 
    refresh 
  } = useAsyncData<IPagedResponse<IWallet>>(
    'walletsList',
    () => $api.wallet.listWallets(limit.value, nextToken.value), 
    {
      watch: [limit, nextToken],
      default: () => ({ items: [], limit: limit.value, nextToken: undefined })
    }
  )

  const wallets = computed<IWallet[]>(() => walletsData.value?.items || [])
  const isLoading = computed<boolean>(() => status.value === 'pending')
  const nextPageToken = computed<string | undefined>(() => walletsData.value?.nextToken)

  return {
    wallets,
    nextPageToken,
    isLoading,
    error,
    refresh
  }
} 