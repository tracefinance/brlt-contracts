import { computed } from 'vue'
import type { Ref } from 'vue'
import type { IPagedResponse, ISigner } from '~/types'

/**
 * Composable for fetching a list of signers with pagination.
 *
 * @param limit - Reactive ref for the number of signers per page.
 * @param offset - Reactive ref for the starting offset.
 * @returns Reactive state including signers, loading status, errors, and refresh function.
 */
export default function (limit: Ref<number>, offset: Ref<number>) {
  const { $api } = useNuxtApp()

  const { 
    data: signersData, 
    status,
    error, 
    refresh 
  } = useAsyncData<IPagedResponse<ISigner>>(
    'signersList',
    () => $api.signer.listSigners(limit.value, offset.value), 
    {
      watch: [limit, offset],
      default: () => ({ items: [], limit: limit.value, offset: offset.value, hasMore: false })
    }
  )

  const signers = computed<ISigner[]>(() => signersData.value?.items || [])
  const hasMore = computed<boolean>(() => signersData.value?.hasMore || false)
  const isLoading = computed<boolean>(() => status.value === 'pending')

  return {
    signers,
    hasMore,
    isLoading,
    error,
    refresh,
  }
} 