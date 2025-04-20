import { computed } from 'vue'
import type { Ref } from 'vue'
import type { IPagedResponse, ISigner } from '~/types'

/**
 * Composable for fetching a list of signers with token-based pagination.
 *
 * @param limit - Reactive ref for the number of signers per page.
 * @param nextToken - Reactive ref for the pagination token.
 * @returns Reactive state including signers, loading status, errors, refresh function, and next token.
 */
export default function (limit: Ref<number>, nextToken: Ref<string | undefined>) {
  const { $api } = useNuxtApp()

  const { 
    data: signersData, 
    status,
    error, 
    refresh 
  } = useAsyncData<IPagedResponse<ISigner>>(
    'signersList',
    () => $api.signer.listSigners(limit.value, nextToken.value), 
    {
      watch: [limit, nextToken],
      default: () => ({ items: [], limit: limit.value, nextToken: undefined })
    }
  )

  const signers = computed<ISigner[]>(() => signersData.value?.items || [])
  const nextPageToken = computed<string | undefined>(() => signersData.value?.nextToken)
  const isLoading = computed<boolean>(() => status.value === 'pending')

  return {
    signers,
    nextPageToken,
    isLoading,
    error,
    refresh,
  }
} 