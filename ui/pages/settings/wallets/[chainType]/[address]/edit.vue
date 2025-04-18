<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import type { IUpdateWalletRequest } from '~/types'
import { toast } from 'vue-sonner'
import { getErrorMessage } from '~/lib/utils'

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
    toast.error(getErrorMessage(newError, 'An unknown error occurred while saving.'))
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
    router.back()
  }  
}

</script>

<template>
  <div class="flex flex-col justify-center space-y-6">
    <Card class="w-full max-w-2xl mx-auto">
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
               <Button variant="link" size="sm" class="p-0 h-auto mt-1" @click="refreshWallet">Retry</Button>
            </AlertDescription>
          </Alert>
        </div>
        
        <form v-else-if="wallet" class="space-y-6" @submit.prevent="handleSaveChanges">
          <div class="space-y-2">
            <Label for="wallet-name">Wallet Name</Label>
            <Input 
              id="wallet-name" 
              v-model="walletName" 
              placeholder="e.g. My Ethereum Hot Wallet"
              required 
            />
          </div>
          
          <div class="space-y-4">
            <Label>Tags (Optional)</Label>
            <div v-for="(tag, index) in tagsList" :key="index" class="flex items-center gap-2">
              <Input v-model="tag.key" placeholder="Key (e.g. environment)" class="flex-1" />
              <Input v-model="tag.value" placeholder="Value (e.g. production)" class="flex-1" />
              <Button 
                type="button" 
                variant="outline" 
                size="icon" 
                :disabled="tagsList.length === 1 && (!tag.key && !tag.value)"
                aria-label="Remove Tag" 
                @click="removeTag(index)"
              >
                <Icon name="lucide:trash-2" class="h-4 w-4" />
              </Button>
            </div>
            <Button type="button" variant="outline" size="sm" @click="addTag">
              <Icon name="lucide:plus" class="h-4 w-4 mr-1" />
              Add Tag
            </Button>
          </div>        
        </form>
        
        <div v-else-if="!isLoadingWallet && !fetchError">
           <p class="text-muted-foreground">Could not load wallet data for the specified chain and address.</p>
        </div>

      </CardContent>
      
      <CardFooter v-if="!isLoadingWallet && !fetchError && wallet" class="flex justify-end space-x-2">
        <Button variant="outline" :disabled="isUpdating" @click="router.back()">
          Cancel
        </Button>
        <Button
          type="submit"
          :disabled="isUpdating || !walletName.trim()"
          @click="handleSaveChanges"
        >
          <Icon v-if="isUpdating" name="svg-spinners:3-dots-fade" class="w-4 h-4 mr-2" />
          {{ isUpdating ? 'Saving...' : 'Save Changes' }}
        </Button>
      </CardFooter>
    </Card>
  </div>
</template>
