import { computed } from 'vue'
import type { IUser } from '~/types'

export default function (userId: string) {
  const { $api } = useNuxtApp()

  const { 
    data: userData, 
    status,
    error, 
    refresh 
  } = useAsyncData<IUser>(
    `user-${userId}`,
    () => $api.user.getUser(userId), 
    {
      default: () => null
    }
  )

  const user = computed<IUser | null>(() => userData.value)
  const isLoading = computed<boolean>(() => status.value === 'pending')

  return {
    user,
    isLoading,
    error,
    refresh,
  }
} 