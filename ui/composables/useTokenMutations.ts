import { ref } from 'vue'
import { useNuxtApp } from '#app'
import type { IAddTokenRequest, IToken, IUpdateTokenRequest } from '~/types'

export default function useTokenMutations() {
  const { $api } = useNuxtApp()
  const isCreating = ref(false)
  const isDeleting = ref(false)
  const isVerifying = ref(false)
  const isUpdating = ref(false)
  const error = ref<Error | null>(null)

  /**
   * Add a new token.
   */
  const addToken = async (payload: IAddTokenRequest): Promise<IToken | null> => {
    isCreating.value = true
    error.value = null
    try {
      const newToken = await $api.token.addToken(payload)
      return newToken
    } catch (err) {
      console.error('Error adding token:', err)
      error.value = err as Error
      return null
    } finally {
      isCreating.value = false
    }
  }

  /**
   * Delete a token.
   * @param address - The address of the token to delete.
   * @returns Whether the operation was successful.
   */
  const deleteToken = async (address: string): Promise<boolean> => {
    isDeleting.value = true;
    error.value = null;
    try {
      await $api.token.deleteToken(address);
      return true;
    } catch (err) {
      console.error('Error deleting token:', err);
      error.value = err as Error;
      return false;
    } finally {
      isDeleting.value = false;
    }
  }
  
  /**
   * Verify a token.
   */
   const verifyToken = async (address: string): Promise<IToken | null> => {
    isVerifying.value = true;
    error.value = null;
    try {
      const verifiedToken = await $api.token.verifyToken(address);
      return verifiedToken;
    } catch (err) {
      console.error('Error verifying token:', err);
      error.value = err as Error;
      return null;
    } finally {
      isVerifying.value = false;
    }
  }

  /**
   * Update an existing token.
   * Updates a token's symbol, type, and decimals by address.
   */
  const updateToken = async (address: string, payload: IUpdateTokenRequest): Promise<IToken | null> => {
    isUpdating.value = true;
    error.value = null;
    try {
      const updatedToken = await $api.token.updateToken(address, payload);
      return updatedToken;
    } catch (err) {
      console.error(`Error updating token ${address}:`, err);
      error.value = err as Error;
      return null;
    } finally {
      isUpdating.value = false;
    }
  }

  return {
    isCreating,
    isDeleting,
    isVerifying,
    isUpdating,
    error,
    addToken,
    deleteToken,
    verifyToken,
    updateToken,
  }
} 