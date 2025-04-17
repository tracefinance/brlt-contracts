import { ref } from 'vue'
import type { ICreateUserRequest, IUpdateUserRequest, IUser } from '~/types'

export default function () {
  const { $api } = useNuxtApp()

  const isCreating = ref(false)
  const isUpdating = ref(false)
  const isDeleting = ref(false)
  const error = ref<Error | null>(null)

  const createUser = async (payload: ICreateUserRequest): Promise<IUser | null> => {
    isCreating.value = true
    error.value = null
    try {
      const newUser = await $api.user.createUser(payload)
      return newUser
    } catch (err: unknown) {
      console.error('Error creating user:', err)
      error.value = err instanceof Error ? err : new Error(String(err))
      return null
    } finally {
      isCreating.value = false
    }
  }

  const updateUser = async (id: string, payload: IUpdateUserRequest): Promise<IUser | null> => {
    isUpdating.value = true
    error.value = null
    try {      
      const updatedUser = await $api.user.updateUser(id, payload)
      return updatedUser
    } catch (err: unknown) {
      console.error('Error updating user:', err)
      error.value = err instanceof Error ? err : new Error(String(err))
      return null
    } finally {
      isUpdating.value = false
    }
  }

  const deleteUser = async (id: string): Promise<boolean> => {
    isDeleting.value = true
    error.value = null
    try {
      await $api.user.deleteUser(id)
      return true
    } catch (err: unknown) {
      console.error('Error deleting user:', err)
      error.value = err instanceof Error ? err : new Error(String(err))
      return false
    } finally {
      isDeleting.value = false
    }
  }

  return {
    isCreating,
    isUpdating,
    isDeleting,
    error,
    createUser,
    updateUser,
    deleteUser
  }
} 