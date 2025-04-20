import { computed } from 'vue'
import type { Ref } from 'vue'
import type { IPagedResponse, IKey } from '~/types'

export default function (limit: Ref<number>, nextToken: Ref<string | undefined>) {
  const { $api } = useNuxtApp()

  const { 
    data: keysData, 
    status,
    error, 
    refresh 
  } = useAsyncData<IPagedResponse<IKey>>(
    'keysList',
    () => $api.key.listKeys(limit.value, nextToken.value), 
    {
      watch: [limit, nextToken],
      default: () => ({ items: [], limit: limit.value, nextToken: undefined })
    }
  )

  const keys = computed<IKey[]>(() => keysData.value?.items || [])
  const nextPageToken = computed<string | undefined>(() => keysData.value?.nextToken)
  const isLoading = computed<boolean>(() => status.value === 'pending')

  return {
    keys,
    nextPageToken,
    isLoading,
    error,
    refresh,
  }
} 