import { computed } from 'vue'
import type { Ref } from 'vue'
import type { ITokenBalanceResponse } from '~/types'

/**
 * Composable for fetching token balances for a specific wallet.
 *
 * @param chainType - Reactive ref for the target chain type.
 * @param address - Reactive ref for the target wallet address.
 * @returns Reactive state including balances, loading status, errors, and refresh function.
 */
export default function (chainType: Ref<string | undefined>, address: Ref<string | undefined>) {
  const { $api } = useNuxtApp()

  const { 
    data: balancesData, 
    status, 
    error, 
    refresh 
  } = useAsyncData<ITokenBalanceResponse[]>(
    'walletBalances',
    async () => {
      const chainTypeValue = chainType.value
      const addressValue = address.value
      if (chainTypeValue && addressValue) {
        return await $api.wallet.getWalletBalance(chainTypeValue, addressValue)
      }
      return []
    },
    {
      watch: [chainType, address],
      default: () => []
    }
  )

  // Make sure balances is always treated as an array
  const balances = computed<ITokenBalanceResponse[]>(() => balancesData.value || [])
  const isLoading = computed<boolean>(() => status.value === 'pending')

  return {
    balances,
    isLoading,
    error,
    refresh,
  }
} 