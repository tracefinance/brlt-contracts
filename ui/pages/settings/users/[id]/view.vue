<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { formatDateTime, shortenAddress } from '~/lib/utils' // Import date formatting and address shortening utility
import { toast } from 'vue-sonner' // Added toast

definePageMeta({
  layout: 'settings'
})

const router = useRouter()
const route = useRoute()
const userId = computed(() => route.params.id as string)

const {
  user,
  error: fetchError,
  isLoading,
  refresh
} = useUserDetails(userId) // Use the same composable as edit page

// User Mutations (for delete) - Added
const {
  deleteUser,
  isDeleting,
  error: deleteError // Renamed from 'error' to avoid conflict
} = useUserMutations()

// Dialog state - Added
const isDeleteDialogOpen = ref(false)

// Function to navigate to the edit page
const goToEditPage = () => {
  if (userId.value) {
    router.push(`/settings/users/${userId.value}/edit`)
  }
}

// Functions for delete dialog - Added
const openDeleteDialog = () => {
  isDeleteDialogOpen.value = true
}

const handleDeleteConfirm = async () => {
  if (!user.value) {
    toast.error('Cannot delete user: User data not available.')
    isDeleteDialogOpen.value = false
    return
  }

  const { id, email } = user.value
  const success = await deleteUser(id) // Use deleteUser

  if (success) {
    toast.success(`User "${email}" (ID: ${shortenAddress(id)}) deleted successfully.`)
    router.push('/settings/users') // Redirect to user list after delete
  } else {
    toast.error(`Failed to delete user "${email}"`, {
      description: deleteError.value?.message || 'Unknown error'
    })
  }

  isDeleteDialogOpen.value = false
}

// Added copyToClipboard function
const copyToClipboard = (text: string | undefined) => {
  if (text) {
    navigator.clipboard.writeText(text)
    toast.success('Copied to clipboard')
  }
}

</script>

<template>
  <div class="flex justify-center">
    <Card class="w-full max-w-2xl">
      <CardHeader>
        <CardTitle>View User Details</CardTitle>
        <CardDescription>Read-only information for the selected user.</CardDescription>
      </CardHeader>
      
      <div v-if="isLoading" class="flex justify-center p-6">
        <Icon name="svg-spinners:180-ring-with-bg" class="h-8 w-8" />
      </div>
      
      <div v-else-if="fetchError" class="p-6">
        <Alert variant="destructive">
          <Icon name="lucide:alert-triangle" class="h-4 w-4" />
          <AlertTitle>Error Loading User</AlertTitle>
          <AlertDescription>
            {{ fetchError.message || 'Failed to load user data' }}
            <Button variant="link" class="p-0 h-auto ml-1" @click="refresh">Retry</Button>
          </AlertDescription>
        </Alert>
      </div>
      
      <template v-else-if="user">
        <CardContent class="space-y-4">
          <div class="space-y-1">
            <Label>User ID</Label>
            <div class="flex items-center gap-2 text-sm">
              <p class="font-mono">{{ user.id }}</p>
              <Button variant="ghost" size="icon" @click="copyToClipboard(user.id)">
                <Icon name="lucide:copy" class="size-4"/>
              </Button>
            </div>
          </div>
          <div class="space-y-1">
            <Label>Email</Label>
            <p class="text-sm font-medium">{{ user.email }}</p>
          </div>
          <div class="space-y-1">
            <Label>Created At</Label>
            <p class="text-sm text-muted-foreground">{{ formatDateTime(user.createdAt) }}</p>
          </div>
          <div class="space-y-1">
            <Label>Updated At</Label>
            <p class="text-sm text-muted-foreground">{{ formatDateTime(user.updatedAt) }}</p>
          </div>
        </CardContent>
        
        <CardFooter class="flex justify-between">
          <Button variant="destructive" @click="openDeleteDialog">
            <Icon name="lucide:trash-2" class="w-4 h-4 mr-2" />
            Delete User
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
          <AlertTitle>User Not Found</AlertTitle>
          <AlertDescription>
            The requested user (ID: {{ userId }}) could not be found.
          </AlertDescription>
        </Alert>
      </div>
    </Card>

    <AlertDialog :open="isDeleteDialogOpen" @update:open="isDeleteDialogOpen = $event">
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Are you absolutely sure?</AlertDialogTitle>
          <AlertDialogDescription>
            This action cannot be undone. This will permanently delete the user
            "{{ user?.email }}" (ID: {{ shortenAddress(user?.id || '') }}).
            This user will lose access and all associated data might be affected.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel :disabled="isDeleting" @click="isDeleteDialogOpen = false">Cancel</AlertDialogCancel>
          <AlertDialogAction :disabled="isDeleting" variant="destructive" @click="handleDeleteConfirm">
            <Icon v-if="isDeleting" name="svg-spinners:3-dots-fade" class="w-4 h-4 mr-2" />
            <span v-if="isDeleting">Deleting...</span>
            <span v-else>Delete User</span>
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>

  </div>
</template> 