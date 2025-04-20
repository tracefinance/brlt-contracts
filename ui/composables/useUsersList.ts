import { computed } from 'vue'
import type { Ref } from 'vue'
import type { IPagedResponse, IUser } from '~/types'

export default function (limit: Ref<number>, nextToken: Ref<string | undefined>) {
  const { $api } = useNuxtApp()

  const { 
    data: usersData, 
    status,
    error, 
    refresh 
  } = useAsyncData<IPagedResponse<IUser>>(
    'usersList',
    () => $api.user.listUsers(limit.value, nextToken.value), 
    {
      watch: [limit, nextToken],
      default: () => ({ items: [], limit: limit.value, nextToken: undefined })
    }
  )

  const users = computed<IUser[]>(() => usersData.value?.items || [])
  const nextPageToken = computed<string | undefined>(() => usersData.value?.nextToken)
  const isLoading = computed<boolean>(() => status.value === 'pending')

  return {
    users,
    nextPageToken,
    isLoading,
    error,
    refresh,
  }
} 