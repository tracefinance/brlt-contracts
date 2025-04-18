<script setup lang="ts">
import { ref, watch } from 'vue'
import { toast } from 'vue-sonner'
import type { IAddress } from '~/types' // Import IAddress
import { getErrorMessage } from '~/lib/utils'

// Props
const props = defineProps<{
  open: boolean
  address: IAddress | null // Accept the full address object or null
}>()

// Emits
const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'confirmRemove', addressId: string): void // Emit the ID on confirmation
}>()

// Internal state for the dialog
const internalOpen = ref(props.open)

// Watch for external changes to the open prop
watch(() => props.open, (newVal) => {
  internalOpen.value = newVal
  if (!newVal) {
    // Reset any component-specific state if needed when closing externally
    removeAddressError.value = null
  }
})

// Watch for internal changes to notify parent
watch(internalOpen, (newVal) => {
  if (newVal !== props.open) {
    emit('update:open', newVal)
  }
  if (!newVal) {
    // Reset when closing internally (e.g., Cancel button)
    removeAddressError.value = null
  }
})

// Use signer mutations specifically for deleting an address
const { 
  deleteAddress,
  isDeletingAddress,
  error: removeAddressError 
} = useSignerMutations()

// Handle confirmation of address removal
const handleConfirm = async () => {
  if (!props.address) return // Guard against null address

  try {
    // Call the deleteAddress mutation using the ID from the prop
    await deleteAddress(props.address.signerId.toString(), props.address.id.toString())
    
    toast.success(`Address ${props.address.address} removed successfully!`)
    emit('confirmRemove', props.address.id.toString()) // Notify parent of success
    internalOpen.value = false // Close dialog on success
  } catch {
    // Error is handled by the error watcher below
    // toast.error(getErrorMessage(err, 'Failed to remove address')) // Can show toast here too if preferred
  }
}

// Watch for remove address errors
watch(removeAddressError, (newError) => {
  if (newError) {
    toast.error(getErrorMessage(newError, 'Failed to remove address'))
  }
})

</script>

<template>
  <AlertDialog :open="internalOpen" @update:open="internalOpen = $event">
    <AlertDialogContent>
      <AlertDialogHeader>
        <AlertDialogTitle>Are you absolutely sure?</AlertDialogTitle>
        <AlertDialogDescription>
          This action cannot be undone. This will permanently remove the address:
          <div v-if="address" class="mt-2 flex items-center gap-2 p-2 bg-muted rounded-md">
            <Web3Icon :symbol="address.chainType" variant="branded" class="size-5"/> 
            <span class="font-mono text-sm break-all text-primary">{{ address.address }}</span>
          </div>
          <div v-else class="mt-2 text-muted-foreground">
            No address selected.
          </div>
        </AlertDialogDescription>
      </AlertDialogHeader>
      <AlertDialogFooter>
        <AlertDialogCancel 
          :disabled="isDeletingAddress" 
          @click="internalOpen = false"
        >
          Cancel
        </AlertDialogCancel>
        <AlertDialogAction 
          variant="destructive" 
          :disabled="isDeletingAddress || !address" 
          @click="handleConfirm"
        >
          <Icon v-if="isDeletingAddress" name="svg-spinners:3-dots-fade" class="w-4 h-4 mr-1" />
          <span v-if="isDeletingAddress">Removing...</span>
          <span v-else>Remove Address</span>
        </AlertDialogAction>
      </AlertDialogFooter>
    </AlertDialogContent>
  </AlertDialog>
</template> 