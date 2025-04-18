<script setup lang="ts">
import { computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { formatDateTime } from '~/lib/utils' // Import utilities
import { getAddressExplorerUrl } from '~/lib/explorers' // Import explorer utility
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

// Compute the explorer URL for the wallet's address
const explorerAddressUrl = computed(() => {
  if (isLoadingChains.value || chainsError.value || !currentWallet.value?.address) return undefined
  const chain = chains.value.find(c => c.type?.toLowerCase() === currentWallet.value?.chainType?.toLowerCase())
  if (!chain?.explorerUrl) return undefined
  return getAddressExplorerUrl(chain.explorerUrl, currentWallet.value.address)
})

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
            <p class="text-sm font-medium">{{ currentWallet.name }}</p>
          </div>
           <div class="space-y-1">
            <Label>Chain Type</Label>
            <div class="flex items-center gap-2">
                <Web3Icon :symbol="currentWallet.chainType" class="size-5" variant="branded" />
                <span class="capitalize">{{ currentWallet.chainType }}</span>
            </div>
          </div>
          <div class="space-y-1">
            <Label>Address</Label>
            <div class="flex items-center gap-2">
            <a v-if="explorerAddressUrl" :href="explorerAddressUrl" target="_blank" rel="noopener noreferrer" class="text-sm font-mono hover:underline block truncate">
              {{ currentWallet.address }}
            </a>
            <p v-else class="text-sm font-mono truncate">{{ currentWallet.address }}</p>
            <Button variant="ghost" size="icon" @click="copyAddress">
              <Icon name="lucide:copy" class="size-4" />
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
            <p class="text-sm font-medium">{{ currentWallet.lastBlockNumber ?? 'N/A' }}</p>
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

        <CardFooter class="flex justify-end gap-2">
          <Button variant="outline" @click="router.back()">Back</Button>
          <Button @click="goToEditPage">
            <Icon name="lucide:edit" class="w-4 h-4 mr-2" />
            Edit
          </Button>
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
  </div>
</template>