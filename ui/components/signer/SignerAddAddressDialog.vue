<script setup lang="ts">
import { ref, watch } from 'vue'
import { toast } from 'vue-sonner'
import { getErrorMessage } from '~/lib/utils'

// Props
const props = defineProps<{
  signerId: string
  open: boolean
}>()

// Emits
const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'addressAdded'): void
}>()

// Internal state for the dialog
const internalOpen = ref(props.open)
const newAddress = ref('')
const newAddressChain = ref<string>('') // Initialize with empty string for ChainSelect

// Watch for external changes to the open prop
watch(() => props.open, (newVal) => {
  internalOpen.value = newVal
  if (!newVal) {
    // Reset form when dialog closes externally
    resetForm()
  }
})

// Watch for internal changes to notify parent
watch(internalOpen, (newVal) => {
  if (newVal !== props.open) {
    emit('update:open', newVal)
  }
  if (!newVal) {
    // Reset form when dialog closes internally (e.g., Cancel button)
    resetForm()
  }
})

// Use signer mutations specifically for adding an address
const { 
  addAddress,
  isAddingAddress,
  error: addAddressError 
} = useSignerMutations()

// Reset form helper
const resetForm = () => {
  newAddress.value = ''
  newAddressChain.value = '' // Reset chain selection
  addAddressError.value = null // Clear previous errors
}

// Add new address function
const handleAddAddress = async () => {
  if (!newAddressChain.value) {
    toast.error('Please select a blockchain.')
    return
  }
  if (!newAddress.value.trim()) {
    toast.error('Please enter a valid address.')
    return
  }
  
  try {
    await addAddress(props.signerId, {
      address: newAddress.value.trim(),
      chainType: newAddressChain.value
    })
    
    toast.success('Address added successfully!')
    emit('addressAdded') // Notify parent
    internalOpen.value = false // Close dialog on success
  } catch (err) {
    toast.error(getErrorMessage(err, 'Failed to add address'))
    // Error state is handled by watching addAddressError
  }
}

// Watch for add address errors
watch(addAddressError, (newError) => {
  if (newError) {
    toast.error(getErrorMessage(newError, 'Failed to add address'))
  }
})
</script>

<template>
  <Dialog :open="internalOpen" @update:open="internalOpen = $event">
    <DialogContent class="sm:max-w-md">
      <DialogHeader>
        <DialogTitle>Add New Address</DialogTitle>
        <DialogDescription>
          Add a blockchain address to this signer.
        </DialogDescription>
      </DialogHeader>
      
      <div class="space-y-6 py-4">
        <div class="space-y-2">
          <Label for="newAddressChain" class="text-base">Blockchain</Label>
          <ChainSelect 
            id="newAddressChain" 
            v-model="newAddressChain" 
            :required="true" 
          />
        </div>
        
        <div class="space-y-2">
          <Label for="newAddress" class="text-base">Address</Label>
          <Input
            id="newAddress"
            v-model="newAddress"
            placeholder="Enter blockchain address"
            class="font-mono bg-background"
            required
          />
        </div>
        <div v-if="addAddressError" class="text-destructive text-sm">
          {{ getErrorMessage(addAddressError, 'Failed to add address') }}
        </div>
      </div>
      
      <DialogFooter>
        <Button 
          type="button" 
          variant="outline" 
          :disabled="isAddingAddress"
          @click="internalOpen = false"
        >
          Cancel
        </Button>
        <Button 
          type="button" 
          variant="default"
          :disabled="isAddingAddress || !newAddress || !newAddressChain"
          @click="handleAddAddress"
        >
          <Icon v-if="isAddingAddress" name="svg-spinners:3-dots-fade" class="w-4 h-4 mr-1" />
          {{ isAddingAddress ? 'Adding...' : 'Add' }}
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template> 