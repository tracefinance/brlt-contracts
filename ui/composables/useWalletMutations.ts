import { ref } from 'vue'
import { useNuxtApp } from '#app'
import type { ICreateWalletRequest, IUpdateWalletRequest, IWallet } from '~/types'

/**
 * Composable for handling wallet mutations (create, update, delete).
 * 
 * Manages loading states and errors, returns API result or indication of success/failure.
 */
export default function () {
  const { $api } = useNuxtApp()

  // --- State References ---
  const isCreating = ref(false)
  const isUpdating = ref(false)
  const isDeleting = ref(false)
  const isActivating = ref(false)
  // Single error ref for simplicity. The consuming component can inspect this.
  const error = ref<Error | null>(null)

  // --- Create Wallet --- 
  const createWallet = async (payload: ICreateWalletRequest): Promise<IWallet | null> => {
    isCreating.value = true
    error.value = null
    try {
      // Assuming $api.wallet.createWallet returns the created wallet object
      const newWallet = await $api.wallet.createWallet(payload)
      return newWallet
    } catch (err) {
      console.error('Error creating wallet:', err)
      error.value = err as Error
      return null
    } finally {
      isCreating.value = false
    }
  }

  // --- Update Wallet ---
  const updateWallet = async (chainType: string, address: string, payload: IUpdateWalletRequest): Promise<IWallet | null> => {
    isUpdating.value = true
    error.value = null
    try {      
      const updatedWallet = await $api.wallet.updateWallet(chainType, address, payload)
      return updatedWallet
    } catch (err) {
      console.error('Error updating wallet:', err)
      error.value = err as Error
      return null
    } finally {
      isUpdating.value = false
    }
  }

  // --- Delete Wallet ---
  const deleteWallet = async (chainType: string, address: string): Promise<boolean> => {
    isDeleting.value = true
    error.value = null
    try {
      await $api.wallet.deleteWallet(chainType, address)
      return true
    } catch (err) {
      console.error('Error deleting wallet:', err)
      error.value = err as Error 
      return false
    } finally {
      isDeleting.value = false
    }
  }

  // --- Activate Token ---
  const activateToken = async (chainType: string, address: string, tokenAddress: string): Promise<boolean> => {
    isActivating.value = true
    error.value = null
    try {
      await $api.wallet.activateToken(chainType, address, tokenAddress)
      return true
    } catch (err) {
      console.error('Error activating token:', err)
      error.value = err as Error
      return false
    } finally {
      isActivating.value = false
    }
  }

  return {
    isCreating,
    isUpdating,
    isDeleting,
    isActivating,
    error,
    createWallet,
    updateWallet,
    deleteWallet,
    activateToken,
  }
} 