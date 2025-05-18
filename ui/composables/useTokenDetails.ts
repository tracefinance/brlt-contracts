import { computed } from 'vue'
import type { Ref } from 'vue'
import type { IToken } from '~/types'

/**
 * Composable for fetching details of a specific token.
 *
 * @param chainType - Reactive ref for the target chain type.
 * @param tokenAddress - Reactive ref for the target token address.
 * @returns Reactive state including the token details, loading status, errors, and refresh function.
 */
export default function (tokenAddress: Ref<string | undefined>) {
  const { $api } = useNuxtApp()

  const { 
    data: currentToken, 
    status, 
    error, 
    refresh 
  } = useAsyncData<IToken | null>(
    'currentToken',
    async () => {
      const tokenAddressValue = tokenAddress.value
      if (tokenAddressValue) {
        return await $api.token.getToken(tokenAddressValue)
      }
      return null
    },
    {
      watch: [tokenAddress],
      default: () => null
    }
  )

  const isLoading = computed<boolean>(() => status.value === 'pending')

  return {
    currentToken,
    isLoading,
    error,
    refresh,
  }
} 