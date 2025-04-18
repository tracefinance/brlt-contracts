<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { formatDateTime, getErrorMessage } from '~/lib/utils'

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

// Combined refresh function
const refresh = () => {
  refreshSigner()
  // Potentially refresh user if needed, but useUserDetails might handle it via watch
}

// Helper to display user info
const userDisplay = computed(() => {
  if (isLoadingUser.value) return 'Loading user info...'
  if (userError.value) return 'Error loading user'
  if (associatedUser.value) return `${associatedUser.value.email} (ID: ${associatedUser.value.id})`
  return 'None'
})
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
              <p class="text-sm text-muted-foreground rounded-md p-2 bg-muted">
                {{ userDisplay }} 
              </p>
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
          
          <CardFooter class="flex justify-end gap-2">
            <Button variant="outline" @click="router.back()">Back</Button>
            <Button @click="goToEditPage">
              <Icon name="lucide:edit" class="w-4 h-4 mr-2" />
              Edit
            </Button>
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
  </div>
</template> 