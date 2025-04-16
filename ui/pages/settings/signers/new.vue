<script setup lang="ts">
import { ref, reactive, watch } from 'vue'
import { useRouter } from 'vue-router'
import { toast } from 'vue-sonner'
import type { ICreateSignerRequest } from '~/types'

// Define page metadata
definePageMeta({
  layout: 'settings'
})

// Composables
const router = useRouter()
const { 
  createSigner,
  isCreating,
  error: mutationError 
} = useSignerMutations()

// Form state
const formData = reactive<ICreateSignerRequest>({
  name: '',
  type: 'internal',
  userId: undefined
})

// Watch for errors from the mutation composable
watch(mutationError, (newError) => {
  if (newError) {
    let errorMessage = 'An unknown error occurred while creating the signer.'
    const errorValue = newError
    const errorAsAny = errorValue as any
    if (errorAsAny?.data?.message) {
      errorMessage = String(errorAsAny.data.message)
    } else if (errorAsAny?.message) {
      errorMessage = String(errorAsAny.message)
    } else if (typeof errorValue === 'string') {
      errorMessage = errorValue
    }
    toast.error(errorMessage)
  }
})

// Form submission handler
const handleSubmit = async () => {
  mutationError.value = null

  // Basic validation
  if (!formData.name || !formData.type) {
    toast.error('Signer Name and Type are required.')
    return
  }

  const newSigner = await createSigner(formData)
  
  if (newSigner) {
    toast.success(`Signer "${newSigner.name}" created successfully!`)
    router.push('/settings/signers')
  }
}
</script>

<template>
  <div class="flex justify-center">
    <Card class="w-full max-w-2xl">
      <CardHeader>
        <CardTitle>Create New Signer</CardTitle>
      </CardHeader>
      <CardContent>
        <form @submit.prevent="handleSubmit" class="space-y-6">
          <!-- Signer Name -->
          <div class="space-y-2">
            <Label for="name">Signer Name</Label>
            <Input 
              id="name" 
              v-model="formData.name"
              placeholder="Enter signer name"
              required
            />
          </div>

          <!-- Signer Type -->
          <div class="space-y-2">
            <Label for="type">Signer Type</Label>
            <Select v-model="formData.type" required>
              <SelectTrigger id="type">
                <SelectValue placeholder="Select signer type" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="internal">
                  <div class="flex items-center gap-2">
                    <Icon name="lucide:shield" class="h-4 w-4" />
                    <span>Internal</span>
                  </div>
                </SelectItem>
                <SelectItem value="external">
                  <div class="flex items-center gap-2">
                    <Icon name="lucide:key" class="h-4 w-4" />
                    <span>External</span>
                  </div>
                </SelectItem>
              </SelectContent>
            </Select>
            <p class="text-sm text-muted-foreground">
              Internal signers are managed by the application, external signers are provided by users
            </p>
          </div>

          <!-- User ID (optional) -->
          <div class="space-y-2">
            <Label for="userId">User ID (Optional)</Label>
            <Input 
              id="userId" 
              v-model="formData.userId"
              type="number"
              placeholder="Enter user ID if applicable"
            />
            <p class="text-sm text-muted-foreground">
              Associate this signer with a specific user
            </p>
          </div>
        </form>
      </CardContent>
      <CardFooter class="flex justify-end gap-2">
        <NuxtLink to="/settings/signers">
          <Button variant="outline">Cancel</Button>
        </NuxtLink>
        <Button type="submit" @click="handleSubmit" :disabled="isCreating">
          <Icon v-if="isCreating" name="svg-spinners:3-dots-fade" class="w-4 h-4 mr-2" />
          {{ isCreating ? 'Creating...' : 'Create Signer' }}
        </Button>
      </CardFooter>
    </Card>
  </div>
</template>

<style scoped>
/* Add any specific styles if needed */
</style> 