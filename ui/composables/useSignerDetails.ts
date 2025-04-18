import { computed } from 'vue'
import type { Ref } from 'vue'
import type { ISigner } from '~/types'

/**
 * Composable for fetching signer details by ID.
 *
 * @param signerId - Reactive ref for the target signer ID.
 * @returns Reactive state including the signer data, loading status, errors, and refresh function.
 */
export default function (signerId: Ref<string | undefined>) {
  const { $api } = useNuxtApp()

  const { 
    data: signer, 
    status, 
    error, 
    refresh 
  } = useAsyncData<ISigner | null>(
    `signer-${signerId.value}`,
    async () => {
      const id = signerId.value
      if (id) {
        return await $api.signer.getSigner(id)
      }
      return null
    },
    {
      watch: [signerId],
      default: () => null
    }
  )

  const isLoading = computed<boolean>(() => status.value === 'pending')

  return {
    signer,
    isLoading,
    error,
    refresh,
  }
} 