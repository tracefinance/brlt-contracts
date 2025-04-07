<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import type { ICreateWalletRequest } from '~/types/wallet'

definePageMeta({
  layout: 'settings'
})

// API client & Router
const { $api } = useNuxtApp()
const router = useRouter()

// Form state
const formData = reactive<ICreateWalletRequest>({
  name: '',
  chainType: '', // Assuming free text for now, could be a <Select> later
  tags: {}
})
const tagsList = ref([{ key: '', value: '' }])

const isLoading = ref(false)
const error = ref<string | null>(null)

// Add a new tag input row
const addTag = () => {
  tagsList.value.push({ key: '', value: '' })
}

// Remove a tag input row
const removeTag = (index: number) => {
  tagsList.value.splice(index, 1)
}

// Handle form submission
const handleSubmit = async () => {
  isLoading.value = true
  error.value = null

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
    error.value = 'Wallet Name and Chain Type are required.'
    isLoading.value = false
    return
  }

  try {
    await $api.wallet.createWallet(payload)
    // Navigate back to the wallets list on success
    router.push('/settings/wallets')
    // TODO: Add success notification/toast
  } catch (err) {
    console.error("Error creating wallet:", err)
    error.value = err instanceof Error ? err.message : 'Failed to create wallet'
  } finally {
    isLoading.value = false
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
            <Input id="chainType" v-model="formData.chainType" required placeholder="ethereum" />
            <!-- TODO: Consider using a <Select> component if chain types are predefined -->
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

          <!-- Error Message -->
          <div v-if="error" class="text-red-500 text-sm">
            {{ error }}
          </div>
        </form>
      </CardContent>
      <CardFooter class="flex justify-end gap-2">
         <NuxtLink to="/settings/wallets">
            <Button variant="outline">Cancel</Button>
          </NuxtLink>
        <Button type="submit" @click="handleSubmit" :disabled="isLoading">
          {{ isLoading ? 'Creating...' : 'Create Wallet' }}
        </Button>
      </CardFooter>
    </Card>
  </div>
</template>

<style scoped>
/* Add any specific styles if needed */
</style> 