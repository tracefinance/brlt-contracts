<script setup lang="ts">
import { formatDistanceToNow } from 'date-fns'
import { computed, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { formatCurrency, shortenAddress } from '~/lib/utils'
import { getTransactionExplorerUrl, getAddressExplorerUrl } from '~/lib/explorers'
import type { IChain } from '~/types'

// Define page metadata
definePageMeta({
  layout: 'wallet'
})

// Get route params
const route = useRoute()

// Reactive route parameters
const address = computed(() => route.params.address as string)
const chainType = computed(() => route.params.chainType as string)
const tokenAddress = computed(() => route.params.tokenAddress as string)

// Use the pagination composable
const { limit, offset, setLimit, previousPage, nextPage } = usePagination(10)

// Use composables for data fetching
const { 
  transactions, 
  error: walletTransactionsError, 
  hasMore,
  refresh
} = useWalletTransactions(chainType, address, tokenAddress, limit, offset)

// Set up auto-refresh interval
let refreshInterval: ReturnType<typeof setInterval> | null = null

onMounted(() => {
  refreshInterval = setInterval(() => {
    refresh()
  }, 3000) // Refresh every 3 seconds
})

onUnmounted(() => {
  if (refreshInterval) {
    clearInterval(refreshInterval)
  }
})

// Use the useChains composable
const { chains, isLoading: isLoadingChains, error: chainsError } = useChains()

// Find the current chain based on the route parameter
const currentChain = computed(() => {
  if (isLoadingChains.value || chainsError.value) return null // Guard against loading/error state
  return chains.value.find((chain: IChain) => chain.type?.toLowerCase() === chainType.value?.toLowerCase())
})

// Computed property for the base explorer URL
const explorerBaseUrl = computed(() => {
  return currentChain.value?.explorerUrl
})

// Combine loading and error states
const error = computed(() => walletTransactionsError.value || chainsError.value)
</script>

<template>
  <!-- Show error state -->
  <div v-if="error">
    <Alert variant="destructive">
      <Icon name="lucide:alert-triangle" class="w-4 h-4" />
      <AlertTitle>Error</AlertTitle>
      <AlertDescription>
        {{ error.message || 'Failed to load data' }}
      </AlertDescription>
    </Alert>
  </div>

  <!-- Show content when loaded and no errors -->
  <div v-else-if="currentChain">                   
    <div v-if="transactions.length === 0">
      <Alert>
        <Icon name="lucide:inbox" class="w-4 h-4" />
          <AlertTitle>No Transactions</AlertTitle>
          <AlertDescription>
            No transactions available for this token yet.
          </AlertDescription>
      </Alert>
    </div>

    <div v-else>                      
      <div class="overflow-auto rounded-lg border">
        <Table>
          <TableHeader>
            <TableRow class="bg-muted hover:bg-muted">
              <TableHead class="w-auto">Hash</TableHead>
              <TableHead class="w-[10%]">Type</TableHead>
              <TableHead class="w-[10%]">From</TableHead>
              <TableHead class="w-[10%]">To</TableHead>
              <TableHead class="w-[8%]">Token</TableHead>
              <TableHead class="w-[10%] text-right">Value</TableHead>
              <TableHead class="w-[15%]">Age</TableHead>
              <TableHead class="w-[110px]">Status</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="tx in transactions" :key="tx.hash">
              <TableCell>
                <a :href="getTransactionExplorerUrl(explorerBaseUrl, tx.hash)" target="_blank" rel="noopener noreferrer" class="hover:underline">
                  {{ shortenAddress(tx.hash) }}
                </a>
              </TableCell>
              <TableCell>
                <TransactionTypeBadge :wallet-address="address" :from-address="tx.fromAddress" />
              </TableCell>
              <TableCell>
                <a :href="getAddressExplorerUrl(explorerBaseUrl, tx.fromAddress)" target="_blank" rel="noopener noreferrer" class="hover:underline">
                  {{ shortenAddress(tx.fromAddress) }}
                </a>
              </TableCell>
              <TableCell>
                <a :href="getAddressExplorerUrl(explorerBaseUrl, tx.toAddress)" target="_blank" rel="noopener noreferrer" class="hover:underline">
                  {{ shortenAddress(tx.toAddress) }}
                </a>
              </TableCell>
              <TableCell class="flex items-center">
                <div class="flex items-center gap-2">
                  <Web3Icon v-if="tx.tokenSymbol" :symbol="tx.tokenSymbol" class="size-5" />
                  <Icon v-else name="lucide:help-circle" class="size-5 text-muted-foreground" />
                  {{ tx.tokenSymbol || 'N/A' }}
                </div>
              </TableCell>
              <TableCell class="text-right font-mono">{{ formatCurrency(tx.value) }}</TableCell>
              <TableCell>
                {{ formatDistanceToNow(new Date(tx.timestamp * 1000), { addSuffix: true }) }}
              </TableCell>
              <TableCell>
                <TransactionStatusBadge :status="tx.status" />
              </TableCell>
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
  </div>
</template>
