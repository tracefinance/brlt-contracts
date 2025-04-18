import { computed } from 'vue'
import type { Ref } from 'vue'
import type { IPagedResponse, IWallet } from '~/types'

/**
 * Composable for fetching a list of wallets with pagination.
 *
 * @param limit - Reactive ref for the number of wallets per page.
 * @param offset - Reactive ref for the starting offset.
 * @returns Reactive state including wallets, loading status, errors, and refresh function.
 */
export default function (limit: Ref<number>, offset: Ref<number>) {
  const { $api } = useNuxtApp()

  const { 
    data: walletsData, 
    status,
    error, 
    refresh 
  } = useAsyncData<IPagedResponse<IWallet>>(
    'walletsList',
    () => $api.wallet.listWallets(limit.value, offset.value), 
    {
      watch: [limit, offset],
      default: () => ({ items: [], limit: limit.value, offset: offset.value, hasMore: false })
    }
  )

  const wallets = computed<IWallet[]>(() => walletsData.value?.items || [])
  const hasMore = computed<boolean>(() => walletsData.value?.hasMore || false)
  const isLoading = computed<boolean>(() => status.value === 'pending')

  return {
    wallets,
    hasMore,
    isLoading,
    error,
    refresh,
  }
} 