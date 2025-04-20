<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { formatDateTime, getErrorMessage, shortenAddress } from '~/lib/utils'
import { toast } from 'vue-sonner'

definePageMeta({
  layout: 'settings'
})

const router = useRouter()
const route = useRoute()
const signerId = ref(route.params.id as string)

// Fetch signer details
const {
  signer,
  error: signerError,
  isLoading: isLoadingSigner,
  refresh: refreshSigner
} = useSignerDetails(signerId)

// Fetch associated user details if userId exists
const userId = computed(() => signer.value?.userId)
const userIdRef = ref<string | undefined>(undefined)

watch(userId, (newUserId) => {
  userIdRef.value = newUserId ? newUserId.toString() : undefined
}, { immediate: true })

const { 
  user: associatedUser, 
  isLoading: isLoadingUser, 
  error: userError 
} = useUserDetails(userIdRef)

// Combined loading and error states
const isLoading = computed(() => isLoadingSigner.value || (userId.value && isLoadingUser.value))
const fetchError = computed(() => signerError.value || userError.value)

// Function to navigate to the edit page
const goToEditPage = () => {
  if (signerId.value) {
    router.push(`/settings/signers/${signerId.value}/edit`)
  }
}

// Signer Mutations (for delete)
const {
  deleteSigner,
  isDeleting,
  error: deleteError
} = useSignerMutations()

// Dialog state
const isDeleteDialogOpen = ref(false)

// Functions for delete dialog
const openDeleteDialog = () => {
  isDeleteDialogOpen.value = true
}

const handleDeleteConfirm = async () => {
  if (!signer.value) {
    toast.error('Cannot delete signer: Signer data not available.')
    isDeleteDialogOpen.value = false
    return
  }

  const { id, name } = signer.value
  const success = await deleteSigner(id)

  if (success) {
    toast.success(`Signer "${name}" (ID: ${shortenAddress(id)}) deleted successfully.`)
    router.push('/settings/signers')
  } else {
    toast.error(`Failed to delete signer "${name}"`, {
      description: deleteError.value?.message || 'Unknown error'
    })
  }

  isDeleteDialogOpen.value = false
}

// Combined refresh function
const refresh = () => {
  refreshSigner()
  // Potentially refresh user if needed, but useUserDetails might handle it via watch
}
</script>

<template>
  <div class="flex justify-center">
    <div class="w-full max-w-2xl space-y-6">
      <!-- Main Signer Details Card -->
      <Card>
        <CardHeader>
          <CardTitle>View Signer Details</CardTitle>
          <CardDescription>Read-only information for the selected signer.</CardDescription>
        </CardHeader>
        
        <div v-if="isLoading" class="flex justify-center p-6">
          <Icon name="svg-spinners:180-ring-with-bg" class="h-8 w-8" />
        </div>
        
        <div v-else-if="fetchError" class="p-6">
          <Alert variant="destructive">
            <Icon name="lucide:alert-triangle" class="h-4 w-4" />
            <AlertTitle>Error Loading Signer Data</AlertTitle>
            <AlertDescription>
              {{ getErrorMessage(fetchError, 'Failed to load signer or user data') }}
              <Button variant="link" class="p-0 h-auto ml-1" @click="refresh">Retry</Button>
            </AlertDescription>
          </Alert>
        </div>
        
        <template v-else-if="signer">
          <CardContent class="space-y-4">
            <div class="space-y-1">
              <Label>Signer ID</Label>
              <p class="text-sm font-medium">{{ signer.id }}</p>
            </div>
            <div class="space-y-1">
              <Label>Name</Label>
              <p class="text-sm font-medium">{{ signer.name }}</p>
            </div>
            <div class="space-y-1">
              <Label>Signer Type</Label>
              <div>
                <SignerTypeBadge :type="signer.type" /> 
              </div>
            </div>
            <div class="space-y-1">
              <Label>Associated User</Label>
              <div class="text-sm">
                <span v-if="isLoadingUser">Loading user info...</span>
                <span v-else-if="userError" class="text-destructive">Error loading user</span>
                <NuxtLink 
                  v-else-if="associatedUser" 
                  :to="`/settings/users/${associatedUser.id}/view`" 
                  class="hover:underline"
                >
                  {{ associatedUser.email }}
                </NuxtLink>
                <span v-else class="text-muted-foreground">None</span>
              </div>
            </div>
            <div class="space-y-1">
              <Label>Created At</Label>
              <p class="text-sm text-muted-foreground">{{ formatDateTime(signer.createdAt) }}</p>
            </div>
            <div class="space-y-1">
              <Label>Updated At</Label>
              <p class="text-sm text-muted-foreground">{{ formatDateTime(signer.updatedAt) }}</p>
            </div>
          </CardContent>
          
          <CardFooter class="flex justify-between">
            <Button variant="destructive" @click="openDeleteDialog">
              <Icon name="lucide:trash-2" class="w-4 h-4 mr-2" />
              Delete Signer
            </Button>
            <div class="flex gap-2">
              <Button variant="outline" @click="router.back()">Back</Button>
              <Button @click="goToEditPage">
                <Icon name="lucide:edit" class="w-4 h-4 mr-2" />
                Edit
              </Button>
            </div>
          </CardFooter>
        </template>
        
        <div v-else class="p-6">
          <Alert>
            <Icon name="lucide:search-x" class="h-4 w-4" />
            <AlertTitle>Signer Not Found</AlertTitle>
            <AlertDescription>
              The requested signer could not be found.
            </AlertDescription>
          </Alert>
        </div>
      </Card>

      <!-- Associated Addresses Card (only if signer exists) -->
      <Card v-if="signer">
        <CardHeader>
          <CardTitle>Associated Addresses</CardTitle>
          <CardDescription>Blockchain addresses managed by this signer.</CardDescription>
        </CardHeader>
        <CardContent>
          <div v-if="signer.addresses && signer.addresses.length > 0" class="space-y-3">
            <div 
              v-for="address in signer.addresses" 
              :key="address.id"
              class="flex items-center justify-between p-3 bg-card border rounded-lg"
            >
              <div class="space-y-1">
                 <p class="flex items-center capitalize font-medium text-sm text-muted-foreground">
                  <Web3Icon :symbol="address.chainType" variant="branded" class="size-5 mr-1"/> 
                  <span>{{ address.chainType }}</span>
                </p>
                <p class="font-mono text-sm break-all">{{ address.address }}</p>
              </div>
              <!-- No actions needed in view mode -->
            </div>
          </div>
          <div v-else>
            <Alert>
              <Icon name="lucide:notebook" class="w-4 h-4" />
              <AlertTitle>No Addresses Found</AlertTitle>
              <AlertDescription>
                No blockchain addresses are associated with this signer yet.
              </AlertDescription>
            </Alert>
          </div>
        </CardContent>
      </Card>
    </div>

    <!-- Delete Confirmation Dialog -->
    <AlertDialog :open="isDeleteDialogOpen" @update:open="isDeleteDialogOpen = $event">
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Are you absolutely sure?</AlertDialogTitle>
          <AlertDialogDescription>
            This action cannot be undone. This will permanently delete the signer
            "{{ signer?.name }}" (ID: {{ shortenAddress(signer?.id || '') }}).
            Any associated addresses will no longer be manageable through this signer.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel :disabled="isDeleting" @click="isDeleteDialogOpen = false">Cancel</AlertDialogCancel>
          <AlertDialogAction :disabled="isDeleting" variant="destructive" @click="handleDeleteConfirm">
            <Icon v-if="isDeleting" name="svg-spinners:3-dots-fade" class="w-4 h-4 mr-2" />
            <span v-if="isDeleting">Deleting...</span>
            <span v-else>Delete Signer</span>
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>

  </div>
</template> 