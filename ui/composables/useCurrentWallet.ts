import { computed } from 'vue'
import type { Ref } from 'vue'
import type { IWallet } from '~/types'

/**
 * Composable for fetching the currently selected wallet details.
 *
 * @param chainType - Reactive ref for the target chain type.
 * @param address - Reactive ref for the target wallet address.
 * @returns Reactive state including the current wallet, loading status, errors, and refresh function.
 */
export default function (chainType: Ref<string | undefined>, address: Ref<string | undefined>) {
  const { $api } = useNuxtApp()

  const { 
    data: currentWallet, 
    status, 
    error, 
    refresh 
  } = useAsyncData<IWallet | null>(
    'currentWallet',
    async () => {
      const chainTypeValue = chainType.value
      const addressValue = address.value
      if (chainTypeValue && addressValue) {
        return await $api.wallet.getWallet(chainTypeValue, addressValue)
      }
      return null
    },
    {
      watch: [chainType, address],
      default: () => null
    }
  )

  const isLoading = computed<boolean>(() => status.value === 'pending')

  return {
    currentWallet,
    isLoading,
    error,
    refresh,
  }
} 