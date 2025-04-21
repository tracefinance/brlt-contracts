<script setup lang="ts">
import { ref } from 'vue'
import { toast } from 'vue-sonner'
import type { IAddTokenRequest } from '~/types'

// Settings
definePageMeta({
  layout: 'settings',
})

// State
const isLoading = ref(false)
const error = ref<Error | null>(null)
const router = useRouter()

// Use the API client directly for now
// TODO: Create and use a dedicated useTokenMutations composable later
const { $api } = useNuxtApp()

async function handleAddToken(formData: IAddTokenRequest) {
  isLoading.value = true
  error.value = null
  try {
    const newToken = await $api.token.addToken(formData)
    toast.success(`Token ${newToken.symbol} added successfully!`)
    router.push('/admin/tokens') // Navigate back to list on success
  } catch (err: any) {
    console.error('Failed to add token:', err)
    error.value = err
    toast.error('Failed to add token', {
      description: err.message || 'An unexpected error occurred.',
    })
  } finally {
    isLoading.value = false
  }
}

function handleCancel() {
  router.back()
}
</script>

<template>
  <div class="space-y-6">
    <PageHeader title="Add New Token" description="Register a new token in the system." />

    <TokenForm
      :is-loading="isLoading"
      @submit="handleAddToken"
      @cancel="handleCancel"
    />

    <Alert v-if="error" variant="destructive" class="mt-4">
      <AlertCircle class="h-4 w-4" />
      <AlertTitle>Error Adding Token</AlertTitle>
      <AlertDescription>
        {{ error.message }}
      </AlertDescription>
    </Alert>
  </div>
</template> 