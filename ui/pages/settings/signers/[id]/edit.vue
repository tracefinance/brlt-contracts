<script setup lang="ts">
import { reactive, watch, computed, ref } from 'vue'
import { toast } from 'vue-sonner'
import type { IUpdateSignerRequest, IAddress } from '~/types'
import { getErrorMessage } from '~/lib/utils'

definePageMeta({
  layout: 'settings'
})

const route = useRoute()
const signerId = ref(route.params.id as string)

// Use the new composable to fetch signer details
const { signer, isLoading: isLoadingSigner, error: signerError, refresh: refreshSigner } = useSignerDetails(signerId)

// Fetch user data if userId exists
const userId = computed(() => signer.value?.userId)
const userIdRef = ref<string | undefined>(undefined)

// Update userIdRef when userId computed property changes
watch(userId, (newUserId) => {
  userIdRef.value = newUserId ? newUserId.toString() : undefined
}, { immediate: true })

// Use the composable for fetching user details
const { 
  user: associatedUser, 
  isLoading: isLoadingUser, 
  error: userError 
} = useUserDetails(userIdRef)

// Use signer mutations for name update only
const { 
  updateSigner,
  isUpdating,
  error: mutationError, // Keep for name update errors
} = useSignerMutations()

// Create reactive form data - only need name now
const formData = reactive<IUpdateSignerRequest>({
  name: '',
  // Initialize with a valid type value (still needed for the PUT request even if not editable)
  type: 'internal',
  userId: undefined
})

// State for dialogs
const isAddressDialogOpen = ref(false)
const isRemoveAddressDialogOpen = ref(false)
const addressToRemove = ref<IAddress | null>(null)

// Initialize form data when signer is loaded
watch(signer, (newSigner) => {
  if (newSigner) {
    formData.name = newSigner.name
    // Still need to set these values for a valid request object
    formData.type = newSigner.type
    formData.userId = newSigner.userId
  }
}, { immediate: true })

// Watch for name update mutation errors
watch(mutationError, (newError) => {
  if (newError) {
    // Check if the dialog is open to avoid showing unrelated errors from the dialog component
    if (!isAddressDialogOpen.value && !isRemoveAddressDialogOpen.value) {
      toast.error(getErrorMessage(newError, 'An unknown error occurred while updating the signer.'))
    }
  }
})

// Handle form submission (for name update)
const handleSubmit = async () => {
  mutationError.value = null

  // Basic validation
  if (!formData.name) {
    toast.error('Signer Name is required.')
    return
  }

  // Use the available updateSigner function
  const updatedSigner = await updateSigner(signerId.value, {
    // Only send the name to be updated
    name: formData.name,
    type: formData.type,
    userId: formData.userId
  })
  
  if (updatedSigner) {
    // Display success toast
    toast.success(`Signer "${updatedSigner.name}" updated successfully!`)
    // Refresh the signer data
    await refreshSigner()
  }
}

// Helper to display user info
const userDisplay = computed(() => {
  if (isLoadingUser.value) return 'Loading user info...'
  if (userError.value) return 'Error loading user'
  if (associatedUser.value) return `${associatedUser.value.email} (ID: ${associatedUser.value.id})`
  return 'None'
})

// Function called when the add address dialog emits 'addressAdded'
const onAddressAdded = async () => {
  await refreshSigner()
}

// Function called when the remove address dialog emits 'confirmRemove'
const onAddressRemoved = async () => {
  await refreshSigner()
  addressToRemove.value = null // Clear the address state after successful removal
}

// Function to open the remove confirmation dialog
const openRemoveAddressDialog = (address: IAddress) => {
  addressToRemove.value = address
  isRemoveAddressDialogOpen.value = true
}

</script>

<template>
  <div>
    <div v-if="isLoadingSigner" class="flex justify-center p-6">
      <Spinner class="h-6 w-6" />
    </div>

    <div v-else-if="signerError || !signer">
      <Alert variant="destructive">
        <Icon name="lucide:alert-triangle" class="w-4 h-4" />
        <AlertTitle>Error Loading Signer</AlertTitle>
        <AlertDescription>
          {{ getErrorMessage(signerError, 'Signer not found') }}
        </AlertDescription>
      </Alert>
    </div>

    <div v-else class="flex flex-col justify-center space-y-6">
      <!-- Signer Information Card -->
      <Card class="w-full max-w-2xl mx-auto">
        <CardHeader>
          <CardTitle>Edit Signer</CardTitle>
          <CardDescription>ID: {{ signer.id }}</CardDescription>
        </CardHeader>
        <CardContent>
          <form class="space-y-6" @submit.prevent="handleSubmit">
            <div class="space-y-2">
              <Label for="name">Name</Label>
              <Input 
                id="name" 
                v-model="formData.name"
                placeholder="Enter signer name"
                required
              />
            </div>

            <div class="space-y-2">
              <Label>Signer Type</Label>
              <div>
                <SignerTypeBadge :type="signer.type" /> 
              </div>
            </div>

            <div class="space-y-2">
              <Label>Associated User</Label>
              <p class="text-sm text-muted-foreground rounded-md p-2 bg-muted">
                {{ userDisplay }} 
              </p>
            </div>

            <div class="flex justify-end gap-2 pt-4 border-t">
              <NuxtLink to="/settings/signers">
                <Button type="button" variant="outline">Cancel</Button> 
              </NuxtLink>
              <Button type="submit" :disabled="isUpdating">
                <Icon v-if="isUpdating" name="svg-spinners:3-dots-fade" class="w-4 h-4 mr-2" />
                {{ isUpdating ? 'Saving...' : 'Save Changes' }} 
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
      
      <!-- Addresses Card -->
      <Card class="w-full max-w-2xl mx-auto">
        <CardHeader class="flex flex-row items-center justify-between">
          <div>
            <CardTitle>Addresses</CardTitle>
            <CardDescription>Manage blockchain addresses for this signer</CardDescription>
          </div>
          <Button 
            type="button" 
            variant="outline" 
            @click="isAddressDialogOpen = true"
          >
            <Icon name="lucide:plus" class="h-4 w-4 mr-1" />
            Add Address
          </Button>
        </CardHeader>
        <CardContent class="space-y-4">
          <!-- Address List -->
          <div v-if="signer.addresses && signer.addresses.length > 0" class="space-y-3">
            <div 
              v-for="address in signer.addresses" 
              :key="address.id"
              class="flex items-center justify-between p-4 bg-card border rounded-lg"
            >
              <div class="space-y-2">
                <p class="flex items-center capitalize font-medium text-sm text-muted-foreground">
                  <Web3Icon :symbol="address.chainType" variant="branded" class="size-5 mr-1"/> 
                  <span>{{ address.chainType }}</span>
                </p>
                <p class="font-mono text-sm break-all">{{ address.address }}</p>
              </div>
              <Button 
                variant="ghost" 
                size="icon"
                class="text-muted-foreground hover:text-destructive"
                @click="openRemoveAddressDialog(address)"
              >
                <Icon name="lucide:trash-2" class="h-5 w-5" />
              </Button>
            </div>
          </div>
          
          <!-- No Addresses Message -->
          <div v-else-if="!isAddressDialogOpen && !isRemoveAddressDialogOpen">
            <Alert>
              <Icon name="lucide:notebook" class="w-4 h-4" />
              <AlertTitle>No Addresses Found</AlertTitle>
              <AlertDescription>
                No blockchain addresses have been added to this signer yet.
              </AlertDescription>
            </Alert>
          </div>
        </CardContent>
      </Card>
    </div>

    <!-- Add Address Dialog Component -->
    <SignerAddAddressDialog 
      v-if="signer" 
      v-model:open="isAddressDialogOpen" 
      :signer-id="signer.id.toString()" 
      @address-added="onAddressAdded" 
    />

    <!-- Remove Address Confirmation Dialog Component -->
    <SignerRemoveAddressDialog
      v-model:open="isRemoveAddressDialogOpen"
      :address="addressToRemove"
      @confirm-remove="onAddressRemoved"
    />

  </div>
</template> 