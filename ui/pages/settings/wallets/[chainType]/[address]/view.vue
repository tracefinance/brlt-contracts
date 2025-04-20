<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { formatDateTime, shortenAddress } from '~/lib/utils' // Import utilities, added shortenAddress
import { getAddressExplorerUrl, getBlockExplorerUrl } from '~/lib/explorers' // Import explorer utility
import { toast } from 'vue-sonner'

definePageMeta({
  layout: 'settings'
})

const router = useRouter()
const route = useRoute()

const chainType = computed(() => route.params.chainType as string)
const address = computed(() => route.params.address as string)

// Use currentWallet instead of wallet
const {
  currentWallet,
  error: fetchError,
  isLoading,
  refresh
} = useWalletDetails(chainType, address)

// Fetch chains to get explorer URL (similar to list page)
const {
  chains,
  isLoading: isLoadingChains,
  error: chainsError
} = useChains()

// Find the chain object (helper computed property)
const currentChain = computed(() => {
  if (isLoadingChains.value || chainsError.value) return undefined
  return chains.value.find(c => c.type?.toLowerCase() === currentWallet.value?.chainType?.toLowerCase())
})

// Compute the explorer URL for the wallet's address
const explorerAddressUrl = computed(() => {
  if (!currentChain.value?.explorerUrl || !currentWallet.value?.address) return undefined
  return getAddressExplorerUrl(currentChain.value.explorerUrl, currentWallet.value.address)
})

// Compute the explorer URL for the wallet's last block
const explorerBlockUrl = computed(() => {
  if (!currentChain.value?.explorerUrl || currentWallet.value?.lastBlockNumber === undefined || currentWallet.value?.lastBlockNumber === null) return undefined
  return getBlockExplorerUrl(currentChain.value.explorerUrl, currentWallet.value.lastBlockNumber)
})

// Wallet Mutations (for delete) - Added
const {
  deleteWallet,
  isDeleting,
  error: deleteError // Renamed from 'error' to avoid conflict
} = useWalletMutations()

// Dialog state - Added
const isDeleteDialogOpen = ref(false)

// Functions for delete dialog - Added
const openDeleteDialog = () => {
  isDeleteDialogOpen.value = true
}

const handleDeleteConfirm = async () => {
  if (!currentWallet.value) {
    toast.error('Cannot delete wallet: Wallet data not available.')
    isDeleteDialogOpen.value = false
    return
  }

  const { chainType, address: walletAddress, name } = currentWallet.value
  const success = await deleteWallet(chainType, walletAddress) // Use deleteWallet

  if (success) {
    toast.success(`Wallet "${name}" (${shortenAddress(walletAddress)}) deleted successfully.`)
    router.push('/settings/wallets') // Redirect to wallet list after delete
  } else {
    toast.error(`Failed to delete wallet "${name}"`, {
      description: deleteError.value?.message || 'Unknown error'
    })
  }

  isDeleteDialogOpen.value = false
}

// Function to navigate to the edit page
const goToEditPage = () => {
  if (chainType.value && address.value) {
    const chainTypeEncoded = encodeURIComponent(chainType.value)
    const addressEncoded = encodeURIComponent(address.value)
    router.push(`/settings/wallets/${chainTypeEncoded}/${addressEncoded}/edit`)
  }
}

const copyAddress = () => {
  if (currentWallet.value?.address) {
    navigator.clipboard.writeText(currentWallet.value.address)
    toast.success('Address copied to clipboard')
  }
}
</script>

<template>
  <div class="flex justify-center">
    <Card class="w-full max-w-2xl">
      <CardHeader>
        <CardTitle>View Wallet Details</CardTitle>
        <CardDescription>Read-only information for the selected wallet.</CardDescription>
      </CardHeader>

      <div v-if="isLoading || isLoadingChains" class="flex justify-center p-6">
        <Icon name="svg-spinners:180-ring-with-bg" class="h-8 w-8" />
      </div>

      <div v-else-if="fetchError || chainsError" class="p-6">
        <Alert variant="destructive">
          <Icon name="lucide:alert-triangle" class="h-4 w-4" />
          <AlertTitle>Error Loading Wallet Data</AlertTitle>
          <AlertDescription>
            {{ (fetchError || chainsError)?.message || 'Failed to load wallet or chain data' }}
            <Button variant="link" class="p-0 h-auto ml-1" @click="refresh">Retry</Button>
          </AlertDescription>
        </Alert>
      </div>

      <template v-else-if="currentWallet">
        <CardContent class="space-y-4">
          <div class="space-y-1">
            <Label>Wallet ID</Label>
            <p class="text-sm font-medium">{{ currentWallet.id }}</p>
          </div>
           <div class="space-y-1">
            <Label>Name</Label>
            <p class="text-sm">{{ currentWallet.name }}</p>
          </div>
           <div class="space-y-1">
            <Label>Chain Type</Label>
            <div class="flex items-center gap-1 text-sm">
                <Web3Icon :symbol="currentWallet.chainType" class="size-5" variant="branded" />
                <span class="capitalize">{{ currentWallet.chainType }}</span>
            </div>
          </div>
          <div class="space-y-1">
            <Label>Address</Label>
            <div class="flex items-center gap-2 text-sm">
              <a v-if="explorerAddressUrl" :href="explorerAddressUrl" target="_blank" rel="noopener noreferrer" class="hover:underline block truncate">
                {{ currentWallet.address }}
              </a>
              <p v-else>{{ currentWallet.address }}</p>
              <Button variant="ghost" size="icon" @click="copyAddress">
                <Icon name="lucide:copy" class="size-4"/>
              </Button>
            </div>
          </div>
           <div class="space-y-1">
            <Label>Key ID</Label>
            <p class="text-sm font-medium">{{ currentWallet.keyId }}</p>
          </div>
          <div class="space-y-1">
            <Label>Tags</Label>
            <div v-if="currentWallet.tags && Object.keys(currentWallet.tags).length > 0" class="flex flex-wrap gap-1">
              <Badge v-for="(value, key) in currentWallet.tags" :key="key" variant="secondary">
                {{ key }}: {{ value }}
              </Badge>
            </div>
            <p v-else class="text-sm text-muted-foreground">No tags</p>
          </div>
           <div class="space-y-1">
            <Label>Last Synced Block</Label>
            <div class="text-sm">
              <a 
                v-if="explorerBlockUrl"
                :href="explorerBlockUrl" 
                target="_blank" 
                rel="noopener noreferrer" 
                class="hover:underline"
              >
                {{ currentWallet.lastBlockNumber }}
              </a>
              <p v-else>
                {{ currentWallet.lastBlockNumber ?? 'N/A' }}
              </p>
            </div>
          </div>
          <div class="space-y-1">
            <Label>Created At</Label>
            <p class="text-sm text-muted-foreground">{{ formatDateTime(currentWallet.createdAt) }}</p>
          </div>
          <div class="space-y-1">
            <Label>Updated At</Label>
            <p class="text-sm text-muted-foreground">{{ formatDateTime(currentWallet.updatedAt) }}</p>
          </div>
        </CardContent>

        <CardFooter class="flex justify-between">
          <Button variant="destructive" @click="openDeleteDialog">
            <Icon name="lucide:trash-2" class="w-4 h-4 mr-2" />
            Delete Wallet
          </Button>
          <div class="flex gap-2">
            <Button variant="outline" @click="router.back()">Back</Button>
            <Button @click="goToEditPage">
              <Icon name="lucide:edit" class="w-4 h-4 mr-2" />
              Edit
            </Button>
          </div>
        </CardFooter>
      </template>

      <div v-else class="p-6">
         <Alert>
          <Icon name="lucide:search-x" class="h-4 w-4" />
          <AlertTitle>Wallet Not Found</AlertTitle>
          <AlertDescription>
            The requested wallet could not be found.
          </AlertDescription>
        </Alert>
      </div>
    </Card>

    <AlertDialog :open="isDeleteDialogOpen" @update:open="isDeleteDialogOpen = $event">
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Are you absolutely sure?</AlertDialogTitle>
          <AlertDialogDescription>
            This action cannot be undone. This will permanently delete the wallet
            "{{ currentWallet?.name }}" ({{ shortenAddress(currentWallet?.address || '') }}).
            All associated tokens and transaction history for this wallet might be affected.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel :disabled="isDeleting" @click="isDeleteDialogOpen = false">Cancel</AlertDialogCancel>
          <AlertDialogAction :disabled="isDeleting" variant="destructive" @click="handleDeleteConfirm">
            <Icon v-if="isDeleting" name="svg-spinners:3-dots-fade" class="w-4 h-4 mr-2" />
            <span v-if="isDeleting">Deleting...</span>
            <span v-else>Delete Wallet</span>
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>

  </div>
</template>