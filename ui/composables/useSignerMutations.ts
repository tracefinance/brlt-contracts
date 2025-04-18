import { ref } from 'vue'
import { useNuxtApp } from '#app'
import type { ICreateSignerRequest, IUpdateSignerRequest, ISigner, IAddAddressRequest, IAddress } from '~/types'

/**
 * Composable for handling signer mutations (create, update, delete, add/remove addresses).
 * 
 * Manages loading states and errors, returns API result or indication of success/failure.
 */
export default function () {
  const { $api } = useNuxtApp()

  // --- State References ---
  const isCreating = ref(false)
  const isUpdating = ref(false)
  const isDeleting = ref(false)
  const isAddingAddress = ref(false)
  const isDeletingAddress = ref(false)
  const error = ref<Error | null>(null)

  // --- Create Signer --- 
  const createSigner = async (payload: ICreateSignerRequest): Promise<ISigner | null> => {
    isCreating.value = true
    error.value = null
    try {
      const newSigner = await $api.signer.createSigner(payload)
      return newSigner
    } catch (err) {
      console.error('Error creating signer:', err)
      error.value = err as Error
      return null
    } finally {
      isCreating.value = false
    }
  }

  // --- Update Signer ---
  const updateSigner = async (id: string, payload: IUpdateSignerRequest): Promise<ISigner | null> => {
    isUpdating.value = true
    error.value = null
    try {      
      const updatedSigner = await $api.signer.updateSigner(id, payload)
      return updatedSigner
    } catch (err) {
      console.error('Error updating signer:', err)
      error.value = err as Error
      return null
    } finally {
      isUpdating.value = false
    }
  }

  // --- Delete Signer ---
  const deleteSigner = async (id: string): Promise<boolean> => {
    isDeleting.value = true
    error.value = null
    try {
      await $api.signer.deleteSigner(id)
      return true
    } catch (err) {
      console.error('Error deleting signer:', err)
      error.value = err as Error
      return false
    } finally {
      isDeleting.value = false
    }
  }

  // --- Add Address ---
  const addAddress = async (signerId: string, payload: IAddAddressRequest): Promise<IAddress | null> => {
    isAddingAddress.value = true
    error.value = null
    try {
      const address = await $api.signer.addAddress(signerId, payload)
      return address
    } catch (err) {
      console.error('Error adding address:', err)
      error.value = err as Error
      return null
    } finally {
      isAddingAddress.value = false
    }
  }

  // --- Delete Address ---
  const deleteAddress = async (signerId: string, addressId: string): Promise<boolean> => {
    isDeletingAddress.value = true
    error.value = null
    try {
      await $api.signer.deleteAddress(signerId, addressId)
      return true
    } catch (err) {
      console.error('Error deleting address:', err)
      error.value = err as Error
      return false
    } finally {
      isDeletingAddress.value = false
    }
  }

  return {
    isCreating,
    isUpdating,
    isDeleting,
    isAddingAddress,
    isDeletingAddress,
    error,
    createSigner,
    updateSigner,
    deleteSigner,
    addAddress,
    deleteAddress
  }
} 