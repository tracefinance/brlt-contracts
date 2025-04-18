<script setup lang="ts">
import { computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { formatDateTime } from '~/lib/utils' // Import date formatting utility

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

// Function to navigate to the edit page
const goToEditPage = () => {
  if (userId.value) {
    router.push(`/settings/users/${userId.value}/edit`)
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
            <p class="text-sm font-medium">{{ user.id }}</p>
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
          <AlertTitle>User Not Found</AlertTitle>
          <AlertDescription>
            The requested user could not be found.
          </AlertDescription>
        </Alert>
      </div>
    </Card>
  </div>
</template> 