<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import type { IUpdateWalletRequest } from '~/types'
import { toast } from 'vue-sonner'

definePageMeta({
  layout: 'settings'
})

const route = useRoute()
const router = useRouter()

const { 
  updateWallet: mutateUpdateWallet,
  isUpdating,
  error: mutationError
} = useWalletMutations()

const routeChainType = computed(() => {
  const param = route.params.chainType
  return typeof param === 'string' ? param : undefined
})

const routeAddress = computed(() => {
  const param = route.params.address
  return typeof param === 'string' ? param : undefined
})

const walletName = ref('')
const tagsList = ref([{ key: '', value: '' }])

const { 
  currentWallet: wallet,
  isLoading: isLoadingWallet,
  error: fetchError,
  refresh: refreshWallet 
} = useWalletDetails(routeChainType, routeAddress)

const addTag = () => {
  tagsList.value.push({ key: '', value: '' })
}

const removeTag = (index: number) => {
  if (tagsList.value.length > 1) {
     tagsList.value.splice(index, 1)
  } else if (tagsList.value.length === 1) {
    tagsList.value = [{ key: '', value: '' }]
  }
}

watch(wallet, (newWallet) => {
  if (newWallet) {
    walletName.value = newWallet.name || ''
    const loadedTags = newWallet.tags ? Object.entries(newWallet.tags) : []
    if (loadedTags.length > 0) {
      tagsList.value = loadedTags.map(([key, value]) => ({ key, value }))
    } else {
      tagsList.value = [{ key: '', value: '' }]
    }
  } else {
    walletName.value = ''
    tagsList.value = [{ key: '', value: '' }]
  }
}, { immediate: true })

watch(mutationError, (newError) => {
  if (newError) {
    // Attempt to extract a meaningful message, defaulting to a generic one
    const errorMessage = (newError as any)?.message
                      || (typeof newError === 'string' ? newError : null) // Handle string errors directly
                      || 'An unknown error occurred while saving.';
    toast.error(errorMessage);
  }
})

const handleSaveChanges = async () => {
  const chainType = routeChainType.value
  const address = routeAddress.value

  if (!chainType || !address || !wallet.value) {
    toast.error('Cannot save, wallet context is invalid or data is missing.')
    return
  }

  if (!walletName.value.trim()) {
    toast.error('Wallet Name is required.')
    return
  }

  mutationError.value = null

  // Convert tagsList back to Record<string, string>
  const tagsPayload: Record<string, string> = tagsList.value
    .map(tag => ({ key: tag.key.trim(), value: tag.value.trim() }))
    .filter(tag => tag.key !== '')
    .reduce((acc, tag) => {
      acc[tag.key] = tag.value
      return acc
    }, {} as Record<string, string>)

  const payload: IUpdateWalletRequest = {
    name: walletName.value.trim(),
    tags: Object.keys(tagsPayload).length > 0 ? tagsPayload : undefined
  }

  const updatedWallet = await mutateUpdateWallet(chainType, address, payload)

  if (updatedWallet) {
    toast.success('Wallet updated successfully!')
    router.push('/settings/wallets')
  }  
}

const handleCancel = () => {
  router.push('/settings/wallets')
}

</script>

<template>
  <Card>
    <CardHeader>
      <CardTitle>Edit Wallet</CardTitle>
      <CardDescription v-if="wallet" class="flex items-center gap-2 pt-1">
        <Web3Icon :symbol="wallet.chainType" class="size-5" variant="branded" />
        <span class="font-medium capitalize">{{ wallet.chainType }} - {{ wallet.address }}</span>
      </CardDescription>
      <CardDescription v-else-if="isLoadingWallet">Loading wallet details...</CardDescription>
      <CardDescription v-else-if="fetchError">Error loading wallet.</CardDescription>
      <CardDescription v-else>Wallet details unavailable.</CardDescription>
    </CardHeader>
    
    <CardContent>
      <div v-if="isLoadingWallet" class="flex items-center justify-center p-8">
        <Icon name="svg-spinners:pulse-3" class="w-6 h-6 mr-2" />
        <span>Loading wallet details...</span>
      </div>

      <div v-else-if="fetchError" class="my-4">
        <Alert variant="destructive">
          <Icon name="lucide:alert-triangle" class="w-4 h-4" />
          <AlertTitle>Error Loading Wallet</AlertTitle>
          <AlertDescription>
            {{ fetchError.message || 'Failed to load wallet details.' }}
             <Button variant="link" size="sm" @click="refreshWallet" class="p-0 h-auto mt-1">Retry</Button>
          </AlertDescription>
        </Alert>
      </div>
      
      <form v-else-if="wallet" @submit.prevent="handleSaveChanges" class="space-y-6">
        <!-- Name Input -->
        <div class="space-y-2">
          <Label for="wallet-name">Wallet Name</Label>
          <Input 
            id="wallet-name" 
            v-model="walletName" 
            placeholder="e.g. My Ethereum Hot Wallet"
            required 
          />
        </div>
        
        <!-- Tags Input (similar to new.vue) -->
        <div class="space-y-4">
          <Label>Tags (Optional)</Label>
          <div v-for="(tag, index) in tagsList" :key="index" class="flex items-center gap-2">
            <Input v-model="tag.key" placeholder="Key (e.g. environment)" class="flex-1" />
            <Input v-model="tag.value" placeholder="Value (e.g. production)" class="flex-1" />
            <Button 
              type="button" 
              variant="outline" 
              size="icon" 
              @click="removeTag(index)"
              :disabled="tagsList.length === 1 && (!tag.key && !tag.value)" 
              aria-label="Remove Tag"
            >
              <Icon name="lucide:trash-2" class="h-4 w-4" />
            </Button>
          </div>
          <Button type="button" variant="outline" size="sm" @click="addTag">
            <Icon name="lucide:plus" class="h-4 w-4 mr-1" />
            Add Tag
          </Button>
        </div>

        <!-- Submission Error -->
        <!-- Use mutationError from the composable if needed, but toast is primary -->
        <!-- 
        <div v-if="mutationError" class="my-4">
          <Alert variant="destructive">
            <Icon name="lucide:alert-triangle" class="w-4 h-4" />
            <AlertTitle>Save Failed</AlertTitle>
            <AlertDescription>
              {{ (mutationError as any)?.data?.message || mutationError.message || 'An error occurred.' }}
            </AlertDescription>
          </Alert>
        </div> 
        -->
        
      </form>
      
      <div v-else-if="!isLoadingWallet && !fetchError">
         <p class="text-muted-foreground">Could not load wallet data for the specified chain and address.</p>
      </div>

    </CardContent>
    
    <CardFooter v-if="!isLoadingWallet && !fetchError && wallet" class="flex justify-end space-x-2">
      <Button variant="outline" @click="handleCancel" :disabled="isUpdating">
        Cancel
      </Button>
      <Button
        type="submit"
        @click="handleSaveChanges"
        :disabled="isUpdating || !walletName.trim()"
      >
        <Icon v-if="isUpdating" name="svg-spinners:3-dots-fade" class="w-4 h-4 mr-2" />
        {{ isUpdating ? 'Saving...' : 'Save Changes' }}
      </Button>
    </CardFooter>
  </Card>
</template>

<style scoped>
/* Add specific styles if needed */
</style> 