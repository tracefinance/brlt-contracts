<script setup lang="ts">
import { reactive, watch, computed } from 'vue'
import { useRouter } from 'vue-router'
import { toast } from 'vue-sonner'
import type { ICreateSignerRequest } from '~/types'
import { getErrorMessage } from '~/lib/utils'

definePageMeta({
  layout: 'settings'
})

const router = useRouter()
const { 
  createSigner,
  isCreating,
  error: mutationError 
} = useSignerMutations()

// Initialize user pagination
const limit = ref(100) // Get a larger list for the dropdown
const offset = ref(0)

// Fetch users for the dropdown
const { users } = useUsersList(limit, offset)

const formData = reactive<ICreateSignerRequest>({
  name: '',
  type: 'internal',
  userId: undefined
})

// Selected user ID for the dropdown
const selectedUserId = ref<number | null>(null)

// Update formData when selectedUserId changes
watch(selectedUserId, (newUserId) => {
  formData.userId = newUserId === null ? undefined : newUserId
})

// Format user options for select
const userOptions = computed(() => {
  return users.value.map(user => ({
    label: `${user.email} (ID: ${user.id})`,
    value: user.id
  }))
})

watch(mutationError, (newError) => {
  if (newError) {
    toast.error(getErrorMessage(newError, 'An unknown error occurred while creating the signer.'))
  }
})

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
        <form class="space-y-6" @submit.prevent="handleSubmit">
          <div class="space-y-2">
            <Label for="name">Signer Name</Label>
            <Input 
              id="name" 
              v-model="formData.name"
              placeholder="Enter signer name"
              required
            />
          </div>

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

          <div class="space-y-2">
            <Label for="userId">Associated User (Optional)</Label>
            <Select v-model="selectedUserId">
              <SelectTrigger id="userId">
                <SelectValue placeholder="Select a user" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem :value="null">None</SelectItem>
                <SelectItem v-for="option in userOptions" :key="option.value" :value="option.value">
                  {{ option.label }}
                </SelectItem>
              </SelectContent>
            </Select>
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
        <Button type="submit" :disabled="isCreating" @click="handleSubmit">
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