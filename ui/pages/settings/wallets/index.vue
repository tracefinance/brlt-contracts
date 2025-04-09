<script setup lang="ts">
import { computed } from 'vue'
import { getAddressExplorerUrl } from '~/lib/explorers'
import type { IWallet } from '~/types'

definePageMeta({
  layout: 'settings'
})

// Use the pagination composable
const { limit, offset, setLimit, previousPage, nextPage } = usePagination(10) // Default limit 10

// Use the composable for wallets data fetching, passing reactive limit/offset
const { 
  wallets, 
  isLoading: isLoadingWallets, 
  error: walletsError,       
  hasMore 
} = useWalletsList(limit, offset)

// Use the composable for chains data fetching
const { 
  chains, 
  isLoading: isLoadingChains,
  error: chainsError 
} = useChains()

// Combine loading states
const isLoading = computed(() => isLoadingWallets.value || isLoadingChains.value)

// Combine errors (show the first error encountered)
const error = computed(() => walletsError.value || chainsError.value)

// Helper function to find the explorer URL for a given wallet
const getWalletExplorerBaseUrl = (wallet: IWallet): string | undefined => {
  if (isLoadingChains.value || chainsError.value) return undefined
  const chain = chains.value.find(c => c.type.toLowerCase() === wallet.chainType.toLowerCase())
  return chain?.explorerUrl
}

</script>

<template>
  <div class="flex flex-col">
    <div v-if="isLoading" class="text-center py-4">
      Loading wallets...
    </div>
    <div v-else-if="error" class="text-red-500 text-center py-4">
      Error: {{ error.message || 'Failed to load wallets' }}
    </div>
    <div v-else-if="wallets.length === 0" class="text-center py-4">
      No wallets found. Create one!
    </div>
    <div v-else class="border rounded-lg overflow-hidden">
      <Table>
        <TableHeader class="bg-muted">
          <TableRow>
            <TableHead>Name</TableHead>
            <TableHead>Chain</TableHead>
            <TableHead>Address</TableHead>
            <TableHead>Last Sync Block</TableHead>
            <TableHead>Tags</TableHead>
            <!-- <TableHead>Actions</TableHead> -->
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow v-for="wallet in wallets" :key="`${wallet.chainType}-${wallet.address}`">
            <TableCell class="font-medium">{{ wallet.name }}</TableCell>
            <TableCell class="flex items-center gap-2">
              <Web3Icon :symbol="wallet.chainType" class="size-5" variant="branded" />
              <span class="capitalize">{{ wallet.chainType }}</span>
            </TableCell>
            <TableCell class="font-mono text-xs">
              <a 
                :href="getAddressExplorerUrl(getWalletExplorerBaseUrl(wallet), wallet.address)"
                target="_blank" 
                rel="noopener noreferrer" 
                class="hover:underline"
                v-if="wallet.address"
              >
                {{ wallet.address }}
              </a>
              <span v-else>N/A</span>
            </TableCell>
            <TableCell>{{ wallet.lastBlockNumber || 'N/A' }}</TableCell>
            <TableCell>
               <div v-if="wallet.tags && Object.keys(wallet.tags).length > 0" class="flex flex-wrap gap-1">
                 <Badge v-for="(value, key) in wallet.tags" :key="key" variant="secondary">
                   {{ key }}: {{ value }}
                 </Badge>
               </div>
               <span v-else class="text-xs text-muted-foreground">No tags</span>
            </TableCell>
            <!-- <TableCell class="text-right">
              <Button variant="ghost" size="icon" @click="// TODO: Implement edit">
                <Icon name="lucide:pencil" class="h-4 w-4" />
              </Button>
              <Button variant="ghost" size="icon" @click="// TODO: Implement delete">
                 <Icon name="lucide:trash-2" class="h-4 w-4 text-red-500" />
              </Button>
            </TableCell> -->
          </TableRow>
        </TableBody>
      </Table>
    </div>
    <div class="flex items-center gap-2 mt-2">
      <PaginationSizeSelect :current-limit="limit" @update:limit="setLimit" />
      <PaginationControls 
        :offset="offset" 
        :limit="limit" 
        :has-more="hasMore" 
        @previous="previousPage"
        @next="nextPage"
      />
    </div>
  </div>
</template>

<style scoped>
/* Add any specific styles if needed */
</style> 