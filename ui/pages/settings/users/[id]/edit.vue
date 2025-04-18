<script setup lang="ts">
import { reactive, watch, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import type { IUpdateUserRequest } from '~/types'
import { toast } from 'vue-sonner'

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
} = useUserDetails(userId)

const { 
  updateUser,
  isUpdating,
  error: mutationError
} = useUserMutations()

const formData = reactive<IUpdateUserRequest>({
  email: '',
  password: ''
})

// Update form when user data is loaded
watch(user, newUser => {
  if (newUser) {
    formData.email = newUser.email
    formData.password = '' // Reset password field for security
  }
}, { immediate: true })

const handleSubmit = async () => {
  if (!user.value) {
    toast.error('User data not loaded.')
    return
  }

  // Create payload, only including non-empty fields
  const payload: IUpdateUserRequest = {}
  
  if (formData.email && formData.email.trim()) {
    payload.email = formData.email.trim()
  }
  
  if (formData.password) {
    payload.password = formData.password
  }

  // Skip update if nothing changed
  if (Object.keys(payload).length === 0) {
    toast.info('No changes to save.')
    return
  }

  const updatedUser = await updateUser(userId.value, payload)

  if (updatedUser) {
    toast.success('User updated successfully!')
    await refresh()
    router.push('/settings/users')
  } else if (mutationError.value) {
    toast.error('Failed to update user', {
      description: mutationError.value.message || 'Unknown error'
    })
  }
}
</script>

<template>
  <div class="flex justify-center">
    <Card class="w-full max-w-2xl">
      <CardHeader>
        <CardTitle>Edit User</CardTitle>
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
      
      <template v-else>
        <CardContent>
          <form class="space-y-6" @submit.prevent="handleSubmit">
            <div class="space-y-2">
              <Label for="email">Email</Label>
              <Input id="email" v-model="formData.email" type="email" required placeholder="user@example.com" />
            </div>

            <div class="space-y-2">
              <Label for="password">Password (leave empty to keep current)</Label>
              <Input id="password" v-model="formData.password" type="password" placeholder="••••••••" />
            </div>
          </form>
        </CardContent>
        
        <CardFooter class="flex justify-end gap-2">
          <Button variant="outline" @click="router.back()">Cancel</Button>
          <Button type="submit" :disabled="isUpdating" @click="handleSubmit">
            <Icon v-if="isUpdating" name="svg-spinners:3-dots-fade" class="w-4 h-4 mr-2" />
            {{ isUpdating ? 'Saving...' : 'Save Changes' }}
          </Button>
        </CardFooter>
      </template>
    </Card>
  </div>
</template> 