import { computed } from 'vue'
import type { Ref } from 'vue'
import type { IUser } from '~/types'

/**
 * Composable for fetching user details by ID.
 *
 * @param userId - Reactive ref for the target user ID.
 * @returns Reactive state including the user data, loading status, errors, and refresh function.
 */
export default function (userId: Ref<string | undefined>) {
  const { $api } = useNuxtApp()

  const { 
    data: user, 
    status, 
    error, 
    refresh 
  } = useAsyncData<IUser | null>(
    `user-${userId.value || 'none'}`,
    async () => {
      const id = userId.value
      if (id) {
        return await $api.user.getUser(id)
      }
      return null
    },
    {
      watch: [userId],
      default: () => null
    }
  )

  const isLoading = computed<boolean>(() => status.value === 'pending')

  return {
    user,
    isLoading,
    error,
    refresh,
  }
} 