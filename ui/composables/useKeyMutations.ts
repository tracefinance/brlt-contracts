import { ref } from 'vue'
import type {
  IKey,
  ICreateKeyRequest,
  IUpdateKeyRequest,
  IImportKeyRequest,
  ISignDataRequest,
  ISignDataResponse,
} from '~/types'

export default function () {
  const { $api } = useNuxtApp()

  const isCreating = ref(false)
  const isUpdating = ref(false)
  const isDeleting = ref(false)
  const isImporting = ref(false)
  const isSigning = ref(false)
  
  const createError = ref<Error | null>(null)
  const updateError = ref<Error | null>(null)
  const deleteError = ref<Error | null>(null)
  const importError = ref<Error | null>(null)
  const signError = ref<Error | null>(null)

  async function createKey(request: ICreateKeyRequest): Promise<IKey | null> {
    isCreating.value = true
    createError.value = null
    
    try {
      const key = await $api.key.createKey(request)
      return key
    } catch (error) {
      createError.value = error as Error
      return null
    } finally {
      isCreating.value = false
    }
  }

  async function updateKey(id: string, request: IUpdateKeyRequest): Promise<IKey | null> {
    isUpdating.value = true
    updateError.value = null
    
    try {
      const key = await $api.key.updateKey(id, request)
      return key
    } catch (error) {
      updateError.value = error as Error
      return null
    } finally {
      isUpdating.value = false
    }
  }

  async function deleteKey(id: string): Promise<boolean> {
    isDeleting.value = true
    deleteError.value = null
    
    try {
      await $api.key.deleteKey(id)
      return true
    } catch (error) {
      deleteError.value = error as Error
      return false
    } finally {
      isDeleting.value = false
    }
  }

  async function importKey(request: IImportKeyRequest): Promise<IKey | null> {
    isImporting.value = true
    importError.value = null
    
    try {
      const key = await $api.key.importKey(request)
      return key
    } catch (error) {
      importError.value = error as Error
      return null
    } finally {
      isImporting.value = false
    }
  }

  async function signData(id: string, request: ISignDataRequest): Promise<ISignDataResponse | null> {
    isSigning.value = true
    signError.value = null
    
    try {
      const response = await $api.key.signData(id, request)
      return response
    } catch (error) {
      signError.value = error as Error
      return null
    } finally {
      isSigning.value = false
    }
  }

  return {
    // State
    isCreating,
    isUpdating,
    isDeleting,
    isImporting,
    isSigning,
    
    // Errors
    createError,
    updateError,
    deleteError,
    importError,
    signError,
    
    // Methods
    createKey,
    updateKey,
    deleteKey,
    importKey,
    signData
  }
} 