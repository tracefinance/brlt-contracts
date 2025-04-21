<script setup lang="ts">
import { watch } from 'vue'
import { useRouter } from 'vue-router'
import { toast } from 'vue-sonner'
import type { IAddTokenRequest } from '~/types'
import { getErrorMessage } from '~/lib/utils'

// Settings
definePageMeta({
  layout: 'settings',
})

// State & Composables
const router = useRouter()
const { 
  addToken: mutateAddToken, 
  isCreating, 
  error: mutationError 
} = useTokenMutations()

// Watch for errors from the composable
watch(mutationError, (newError) => {
  if (newError) {
    toast.error('Failed to add token', {
      description: getErrorMessage(newError, 'An unexpected error occurred.'),
    })
  }
})

// Handle form submission
async function handleAddToken(formData: IAddTokenRequest) {
  mutationError.value = null
  
  // Basic validation (can be enhanced)
  if (!formData.address || !formData.chainType || !formData.symbol || formData.decimals === undefined || formData.decimals === null || !formData.type) {
    toast.error('All fields are required.');
    return;
  }
  
  const newToken = await mutateAddToken(formData)

  if (newToken) {
    toast.success(`Token ${newToken.symbol} added successfully!`)
    router.push('/settings/tokens') 
  }
}

function handleCancel() {
  router.back()
}
</script>

<template>
  <div class="flex justify-center">
    <Card class="w-full max-w-2xl">
      <CardHeader>
        <CardTitle>Add New Token</CardTitle>
        <CardDescription>Register a new token in the system.</CardDescription>
      </CardHeader>
      <CardContent>
        <TokenNewForm
          :is-loading="isCreating" 
          @submit="handleAddToken"
          @cancel="handleCancel"
        />
      </CardContent>
    </Card>
  </div>
</template> 