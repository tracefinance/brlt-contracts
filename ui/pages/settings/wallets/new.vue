<script setup lang="ts">
import { ref, reactive, watch } from 'vue'
import { useRouter } from 'vue-router'
import type { ICreateWalletRequest } from '~/types'
import { toast } from 'vue-sonner'

definePageMeta({
  layout: 'settings'
})

// Composables
const router = useRouter()
const { 
  createWallet: mutateCreateWallet,
  isCreating, 
  error: mutationError 
} = useWalletMutations()
const { chains, isLoading: isLoadingChains, error: chainsError, refresh: refreshChains } = useChains()

// Form state
const formData = reactive<ICreateWalletRequest>({
  name: '',
  chainType: '', // Initialize as empty, will be set by Select
  tags: {}
})
const tagsList = ref([{ key: '', value: '' }])

// Add a new tag input row
const addTag = () => {
  tagsList.value.push({ key: '', value: '' })
}

// Remove a tag input row
const removeTag = (index: number) => {
  tagsList.value.splice(index, 1)
}

// Watch for errors from the mutation composable
watch(mutationError, (newError) => {
  if (newError) {
    let errorMessage = 'An unknown error occurred while creating the wallet.'
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

// Handle form submission
const handleSubmit = async () => {
  mutationError.value = null

  // Convert tagsList to the Record<string, string> format
  const tags: Record<string, string> = tagsList.value
    .filter(tag => tag.key.trim() !== '' && tag.value.trim() !== '') // Filter out empty tags
    .reduce((acc, tag) => {
      acc[tag.key.trim()] = tag.value.trim()
      return acc
    }, {} as Record<string, string>)

  const payload: ICreateWalletRequest = {
    name: formData.name.trim(),
    chainType: formData.chainType.trim(),
    tags: Object.keys(tags).length > 0 ? tags : undefined // Only send tags if not empty
  }

  // Basic validation
  if (!payload.name || !payload.chainType) {
    toast.error('Wallet Name and Chain Type are required.')
    return
  }

  const newWallet = await mutateCreateWallet(payload)

  if (newWallet) {
    toast.success('Wallet created successfully!')
    router.push('/settings/wallets')
  }
}
</script>

<template>
  <div class="flex justify-center">
    <Card class="w-full max-w-2xl">
      <CardHeader>
        <CardTitle>Create New Wallet</CardTitle>
      </CardHeader>
      <CardContent>
        <form @submit.prevent="handleSubmit" class="space-y-6">
          <!-- Wallet Name -->
          <div class="space-y-2">
            <Label for="name">Wallet Name</Label>
            <Input id="name" v-model="formData.name" required placeholder="My Ethereum Wallet" />
          </div>

          <!-- Chain Type -->
          <div class="space-y-2">
            <Label for="chainType">Chain Type</Label>
            <!-- Loading State -->
            <div v-if="isLoadingChains" class="flex items-center space-x-2 text-muted-foreground">
              <Icon name="svg-spinners:180-ring-with-bg" class="h-4 w-4" />
              <span>Loading supported chains...</span>
            </div>
            <!-- Error State -->
            <div v-else-if="chainsError" class="text-red-500 text-sm">
              <span>Error loading chains: {{ chainsError.message }}.</span>
              <Button variant="link" size="sm" @click="refreshChains" class="p-0 h-auto ml-1">Retry</Button>
            </div>
            <!-- Select Input -->
            <Select v-else v-model="formData.chainType" required>
              <SelectTrigger id="chainType">
                <SelectValue placeholder="Select chain type..." />
              </SelectTrigger>
              <SelectContent>
                <SelectItem v-for="chain in chains" :key="chain.type" :value="chain.type">
                  <div class="flex items-center gap-2"> 
                    <Web3Icon :symbol="chain.type" size="16px" /> 
                    <span>{{ chain.name }}</span>
                  </div>
                </SelectItem>
                <div v-if="chains.length === 0" class="p-2 text-center text-sm text-muted-foreground">
                   No chains available.
                </div>
              </SelectContent>
            </Select>
          </div>

          <!-- Tags -->
          <div class="space-y-4">
            <Label>Tags (Optional)</Label>
            <div v-for="(tag, index) in tagsList" :key="index" class="flex items-center gap-2">
              <Input v-model="tag.key" placeholder="Key" class="flex-1" />
              <Input v-model="tag.value" placeholder="Value" class="flex-1" />
              <Button type="button" variant="outline" size="icon" @click="removeTag(index)" :disabled="tagsList.length <= 1">
                <Icon name="lucide:trash-2" class="h-4 w-4" />
              </Button>
            </div>
            <Button type="button" variant="outline" size="sm" @click="addTag">
              <Icon name="lucide:plus" class="h-4 w-4 mr-1" />
              Add Tag
            </Button>
          </div>
        </form>
      </CardContent>
      <CardFooter class="flex justify-end gap-2">
         <NuxtLink to="/settings/wallets">
            <Button variant="outline">Cancel</Button>
          </NuxtLink>
        <Button type="submit" @click="handleSubmit" :disabled="isCreating">
          <Icon v-if="isCreating" name="svg-spinners:3-dots-fade" class="w-4 h-4 mr-2" />
          {{ isCreating ? 'Creating...' : 'Create Wallet' }}
        </Button>
      </CardFooter>
    </Card>
  </div>
</template>

<style scoped>
/* Add any specific styles if needed */
</style> 