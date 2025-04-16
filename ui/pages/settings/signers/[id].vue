<script setup lang="ts">
import { ref } from 'vue'
import { toast } from 'vue-sonner'
import type { IAddAddressRequest, IAddress, ISigner } from '~/types'

// Define page metadata
definePageMeta({
  layout: 'settings'
})

// Get route params
const route = useRoute()
const signerId = route.params.id as string

// Use Nuxt data fetching
const { $api } = useNuxtApp()
const { data: signer, refresh: refreshSigner, error: signerError, pending: isLoading } = 
  useAsyncData<ISigner>(`signer-${signerId}`, () => $api.signer.getSigner(signerId))

// Get mutations
const { 
  updateSigner, isUpdating,
  addAddress, isAddingAddress,
  deleteAddress, isDeletingAddress,
  error: mutationError 
} = useSignerMutations()

// State for address form
const showAddAddressForm = ref(false)
const newAddress = ref<IAddAddressRequest>({
  chainType: '',
  address: ''
})

// Handle adding a new address
const handleAddAddress = async () => {
  if (!newAddress.value.chainType || !newAddress.value.address) {
    toast.error('Please enter both chain type and address.')
    return
  }

  const result = await addAddress(signerId, newAddress.value)
  if (result) {
    // Reset form and refresh data
    toast.success('Address added successfully.')
    newAddress.value = { chainType: '', address: '' }
    showAddAddressForm.value = false
    await refreshSigner()
  } else {
    toast.error('Failed to add address', {
      description: mutationError.value?.message || 'Unknown error'
    })
  }
}

// Delete confirmation dialog
const isDeleteDialogOpen = ref(false)
const addressToDelete = ref<{ id: string, address: string } | null>(null)

const openDeleteDialog = (addressId: string, addressValue: string) => {
  addressToDelete.value = { id: addressId, address: addressValue }
  isDeleteDialogOpen.value = true
}

const handleDeleteConfirm = async () => {
  if (!addressToDelete.value) {
    toast.error('Cannot delete address: Invalid data provided.')
    isDeleteDialogOpen.value = false
    addressToDelete.value = null
    return
  }

  const { id, address } = addressToDelete.value
  
  const success = await deleteAddress(signerId, id)

  if (success) {
    toast.success(`Address deleted successfully.`)
    await refreshSigner()
  } else {
    toast.error(`Failed to delete address`, {
      description: mutationError.value?.message || 'Unknown error'
    })
  }

  isDeleteDialogOpen.value = false
  addressToDelete.value = null
}

// Format date helper
const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleString()
}
</script>

<template>
  <div>
    <!-- Loading state -->
    <div v-if="isLoading" class="flex justify-center p-6">
      <Spinner class="h-6 w-6" />
    </div>

    <!-- Error state -->
    <div v-else-if="signerError || !signer">
      <Alert variant="destructive">
        <Icon name="lucide:alert-triangle" class="w-4 h-4" />
        <AlertTitle>Error Loading Signer</AlertTitle>
        <AlertDescription>
          {{ signerError?.message || 'Signer not found' }}
        </AlertDescription>
      </Alert>
    </div>

    <!-- Signer details -->
    <div v-else>
      <div class="flex justify-between items-center mb-6">
        <h1 class="text-2xl font-bold">{{ signer.name }}</h1>
      </div>

      <!-- Signer info card -->
      <Card class="mb-8">
        <CardHeader>
          <CardTitle>Signer Information</CardTitle>
          <CardDescription>Details about this signer</CardDescription>
        </CardHeader>
        <CardContent>
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <p class="text-sm font-medium text-muted-foreground mb-1">Name</p>
              <p>{{ signer.name }}</p>
            </div>
            <div>
              <p class="text-sm font-medium text-muted-foreground mb-1">Type</p>
              <Badge variant="secondary">{{ signer.type }}</Badge>
            </div>
            <div>
              <p class="text-sm font-medium text-muted-foreground mb-1">Created</p>
              <p>{{ formatDate(signer.createdAt) }}</p>
            </div>
            <div>
              <p class="text-sm font-medium text-muted-foreground mb-1">Last Updated</p>
              <p>{{ formatDate(signer.updatedAt) }}</p>
            </div>
            <div v-if="signer.userId">
              <p class="text-sm font-medium text-muted-foreground mb-1">User ID</p>
              <p>{{ signer.userId }}</p>
            </div>
          </div>
        </CardContent>
      </Card>

      <!-- Addresses section -->
      <div class="mb-8">
        <div class="flex justify-between items-center mb-4">
          <h2 class="text-xl font-semibold">Addresses</h2>
          <Button v-if="!showAddAddressForm" variant="outline" @click="showAddAddressForm = true">
            <Icon name="lucide:plus" class="h-4 w-4 mr-2" />
            Add Address
          </Button>
        </div>

        <!-- Add Address Form -->
        <Card v-if="showAddAddressForm" class="mb-4">
          <CardHeader>
            <CardTitle>Add New Address</CardTitle>
          </CardHeader>
          <CardContent>
            <form @submit.prevent="handleAddAddress" class="space-y-4">
              <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div class="space-y-2">
                  <Label for="chainType">Chain Type</Label>
                  <Input 
                    id="chainType" 
                    v-model="newAddress.chainType" 
                    placeholder="e.g., ethereum, bitcoin" 
                    required
                  />
                </div>
                <div class="space-y-2">
                  <Label for="address">Address</Label>
                  <Input 
                    id="address" 
                    v-model="newAddress.address" 
                    placeholder="Wallet address" 
                    required
                  />
                </div>
              </div>
              <div class="flex justify-end gap-2">
                <Button 
                  type="button" 
                  variant="ghost" 
                  @click="showAddAddressForm = false"
                >
                  Cancel
                </Button>
                <Button 
                  type="submit" 
                  :disabled="isAddingAddress || !newAddress.chainType || !newAddress.address"
                >
                  Add Address
                </Button>
              </div>
            </form>
          </CardContent>
        </Card>

        <!-- Addresses table -->
        <div v-if="signer.addresses && signer.addresses.length > 0" class="border rounded-lg overflow-hidden">
          <Table>
            <TableHeader class="bg-muted">
              <TableRow>
                <TableHead class="w-[20%]">Chain Type</TableHead>
                <TableHead class="w-[60%]">Address</TableHead>
                <TableHead class="w-[15%]">Created</TableHead>
                <TableHead class="w-[5%] text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow v-for="address in signer.addresses" :key="address.id">
                <TableCell>
                  <Badge variant="outline">{{ address.chainType }}</Badge>
                </TableCell>
                <TableCell class="font-mono text-sm">{{ address.address }}</TableCell>
                <TableCell>{{ formatDate(address.createdAt) }}</TableCell>
                <TableCell class="text-right">
                  <DropdownMenu>
                    <DropdownMenuTrigger as-child>
                      <Button variant="ghost" class="h-8 w-8 p-0">
                        <span class="sr-only">Open menu</span>
                        <Icon name="lucide:more-horizontal" class="size-4" />
                      </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end">
                      <DropdownMenuItem 
                        @click="openDeleteDialog(address.id.toString(), address.address)"
                        class="text-destructive focus:text-destructive focus:bg-destructive/10"
                      >
                        <Icon name="lucide:trash-2" class="mr-2 size-4" />
                        <span>Delete</span>
                      </DropdownMenuItem>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </div>

        <!-- No addresses -->
        <Alert v-else>
          <Icon name="lucide:inbox" class="w-4 h-4" />
          <AlertTitle>No Addresses</AlertTitle>
          <AlertDescription>
            This signer doesn't have any addresses associated with it yet.
          </AlertDescription>
        </Alert>
      </div>
    </div>
  </div>

  <AlertDialog :open="isDeleteDialogOpen" @update:open="isDeleteDialogOpen = $event">
    <AlertDialogContent>
      <AlertDialogHeader>
        <AlertDialogTitle>Are you absolutely sure?</AlertDialogTitle>
        <AlertDialogDescription>
          This action cannot be undone. This will permanently delete the address
          "{{ addressToDelete?.address }}" from this signer.
        </AlertDialogDescription>
      </AlertDialogHeader>
      <AlertDialogFooter>
        <AlertDialogCancel :disabled="isDeletingAddress" @click="isDeleteDialogOpen = false">Cancel</AlertDialogCancel>
        <AlertDialogAction @click="handleDeleteConfirm" variant="destructive" :disabled="isDeletingAddress">
          <span v-if="isDeletingAddress">Deleting...</span>
          <span v-else>Delete Address</span>
        </AlertDialogAction>
      </AlertDialogFooter>
    </AlertDialogContent>
  </AlertDialog>
</template> 