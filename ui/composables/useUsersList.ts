import { computed } from 'vue'
import type { Ref } from 'vue'
import type { IPagedResponse, IUser } from '~/types'

export default function (limit: Ref<number>, offset: Ref<number>) {
  const { $api } = useNuxtApp()

  const { 
    data: usersData, 
    status,
    error, 
    refresh 
  } = useAsyncData<IPagedResponse<IUser>>(
    'usersList',
    () => $api.user.listUsers(limit.value, offset.value), 
    {
      watch: [limit, offset],
      default: () => ({ items: [], limit: limit.value, offset: offset.value, hasMore: false })
    }
  )

  const users = computed<IUser[]>(() => usersData.value?.items || [])
  const hasMore = computed<boolean>(() => usersData.value?.hasMore || false)
  const isLoading = computed<boolean>(() => status.value === 'pending')

  return {
    users,
    hasMore,
    isLoading,
    error,
    refresh,
  }
} 