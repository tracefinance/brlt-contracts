import { ref } from 'vue'
import { useNuxtApp } from '#app'
import type { IAddTokenRequest, IToken, ChainType, IUpdateTokenRequest } from '~/types'
// Assuming getErrorMessage is used elsewhere or can be added later if needed
// import { getErrorMessage } from '~/lib/utils'

export default function useTokenMutations() {
  const { $api } = useNuxtApp()
  const isCreating = ref(false)
  const isDeleting = ref(false) // Add if delete functionality is needed later
  const isVerifying = ref(false) // Add if verify functionality is needed later
  const isUpdating = ref(false) // Added for update operation
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
      // Optionally re-throw or handle specific error types if needed
      return null
    } finally {
      isCreating.value = false
    }
  }

  /**
   * Delete a token.
   */
  const deleteToken = async (chainType: ChainType, address: string): Promise<boolean> => {
    isDeleting.value = true;
    error.value = null;
    try {
      await $api.token.deleteToken(chainType, address);
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

  // Expose reactive state and methods
  return {
    isCreating,
    isDeleting,
    isVerifying,
    isUpdating, // Expose isUpdating
    error,
    addToken,
    deleteToken,
    verifyToken,
    updateToken, // Expose updateToken
    // Expose specific mutation errors if needed, e.g., createError, deleteError
  }
} 