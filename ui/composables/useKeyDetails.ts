import { computed } from 'vue'
import type { Ref } from 'vue'
import type { IKey } from '~/types'

/**
 * Composable for fetching key details by ID.
 *
 * @param keyId - Reactive ref for the target key ID.
 * @returns Reactive state including the key data, loading status, errors, and refresh function.
 */
export default function (keyId: Ref<string | undefined>) {
  const { $api } = useNuxtApp()

  const { 
    data: key, 
    status, 
    error, 
    refresh 
  } = useAsyncData<IKey | null>(
    `key-${keyId.value || 'none'}`,
    async () => {
      const id = keyId.value
      if (id) {
        return await $api.key.getKey(id)
      }
      return null
    },
    {
      watch: [keyId],
      default: () => null
    }
  )

  const isLoading = computed<boolean>(() => status.value === 'pending')

  return {
    key,
    isLoading,
    error,
    refresh,
  }
} 