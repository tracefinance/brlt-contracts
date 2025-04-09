<script setup lang="ts">
import { formatDistanceToNow } from 'date-fns'
import { computed } from 'vue'
import { shortenAddress, formatCurrency } from '~/lib/utils'

// Define page metadata
definePageMeta({
  layout: 'wallet'
})

// Get route params and router
const route = useRoute()
const router = useRouter()

// Reactive route parameters
const address = computed(() => route.params.address as string)
const chainType = computed(() => route.params.chainType as string)
const tokenAddress = computed(() => route.params.tokenAddress as string)

// Reactive pagination parameters from route query
const limit = computed(() => {
  const queryLimit = route.query.limit ? Number(route.query.limit) : 10
  return isNaN(queryLimit) || queryLimit <= 0 ? 10 : queryLimit // Ensure limit is positive
})

const offset = computed(() => {
  const queryOffset = route.query.offset ? Number(route.query.offset) : 0
  return isNaN(queryOffset) || queryOffset < 0 ? 0 : queryOffset // Ensure offset is non-negative
})

// Use composables for data fetching
const { 
  currentWallet, 
  isLoading: isLoadingWallet, 
  error: errorWallet 
} = useCurrentWallet(chainType, address)

const { 
  currentToken, 
  isLoading: isLoadingToken, 
  error: errorToken 
} = useTokenDetails(chainType, tokenAddress)

const { 
  transactions, 
  hasMore: hasMoreTransactions, 
  isLoading: isLoadingTransactions, 
  error: errorTransactions, 
} = useWalletTransactions(chainType, address, tokenAddress, limit, offset)


const error = computed(() => {
  if (errorWallet.value) return `Failed to load wallet: ${errorWallet.value.message}`
  if (errorToken.value) return `Failed to load token: ${errorToken.value.message}`
  if (errorTransactions.value) return `Failed to load transactions: ${errorTransactions.value.message}`
  return null
})

const isLoading = computed(() => 
  isLoadingWallet.value || isLoadingToken.value || isLoadingTransactions.value
)

// Handle page size change
function handleLimitChange(newLimit: number) {
  // Reset offset to 0 when limit changes to avoid invalid page numbers
  router.push({ 
    query: { 
      ...route.query, 
      limit: newLimit,
      offset: 0 // Reset offset when changing page size
    } 
  })
}

// Handle pagination events
function handlePreviousPage() {
  const newOffset = Math.max(0, offset.value - limit.value);
  router.push({ 
    query: { 
      ...route.query, 
      offset: newOffset 
    } 
  });
}

function handleNextPage() {
  const newOffset = offset.value + limit.value;
  router.push({ 
    query: { 
      ...route.query, 
      offset: newOffset 
    } 
  });
}

// Determine the explorer URL based on chainType
function getExplorerBaseUrl(chainType: string) {
  // Simple example, replace with actual logic
  if (chainType.toLowerCase() === 'ethereum') {
    return "https://etherscan.io"
  }
  // Add other chains as needed
  return "https://etherscan.io" // Default fallback
}

const explorerBaseUrl = computed(() => getExplorerBaseUrl(chainType.value))
</script>

<template>
  <div>
    <!-- Show loading state -->
    <div v-if="isLoading" class="p-8 text-center text-muted-foreground">
      Loading transactions...
    </div>

    <!-- Show error state -->
    <div v-else-if="error" class="p-4 m-4 bg-red-50 text-red-700 rounded-md">
      {{ error }}
    </div>

    <!-- Show content when loaded and no errors -->
    <div v-else-if="currentWallet && currentToken">          
        <div class="p-4">          
          <div v-if="transactions.length === 0" class="text-muted-foreground">
            No transactions available for this token yet.
          </div>
          
          <div v-else>                      
            <div class="overflow-auto rounded-lg border">
              <Table>
                <TableHeader>
                  <TableRow class="bg-muted hover:bg-muted">
                    <TableHead>Hash</TableHead>
                    <TableHead>Type</TableHead>
                    <TableHead>From</TableHead>
                    <TableHead>To</TableHead>
                    <TableHead>Token</TableHead>
                    <TableHead class="text-right">Value</TableHead>
                    <TableHead>Age</TableHead>
                    <TableHead>Status</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  <TableRow v-for="tx in transactions" :key="tx.hash">
                    <TableCell>
                      <a :href="`${explorerBaseUrl}/tx/${tx.hash}`" target="_blank" rel="noopener noreferrer" class="text-blue-600 hover:underline">
                        {{ shortenAddress(tx.hash, 6, 6) }}
                      </a>
                    </TableCell>
                    <TableCell>
                      <Badge variant="outline" class="rounded-full px-2 py-1">
                        <Icon 
                          :name="tx.fromAddress.toLowerCase() === address.toLowerCase() ? 'lucide:arrow-up-right' : 'lucide:arrow-down-left'" 
                          class="mr-1 h-4 w-4" 
                        />
                        {{ tx.fromAddress.toLowerCase() === address.toLowerCase() ? 'Send' : 'Receive' }}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <a :href="`${explorerBaseUrl}/address/${tx.fromAddress}`" target="_blank" rel="noopener noreferrer" class="hover:underline">
                        {{ shortenAddress(tx.fromAddress) }}
                      </a>
                    </TableCell>
                    <TableCell>
                      <a :href="`${explorerBaseUrl}/address/${tx.toAddress}`" target="_blank" rel="noopener noreferrer" class="hover:underline">
                        {{ shortenAddress(tx.toAddress) }}
                      </a>
                    </TableCell>
                    <TableCell class="flex items-center">
                      <Web3Icon v-if="tx.tokenSymbol" :symbol="tx.tokenSymbol" class="mr-2 h-5 w-5" />
                      <Icon v-else name="lucide:help-circle" class="mr-2 h-5 w-5 text-muted-foreground" />
                      {{ tx.tokenSymbol || 'N/A' }}
                    </TableCell>
                    <TableCell class="text-right">{{ formatCurrency(tx.value) }}</TableCell>
                    <TableCell :title="new Date(tx.timestamp * 1000).toLocaleString()">
                      {{ formatDistanceToNow(new Date(tx.timestamp * 1000), { addSuffix: true }) }}
                    </TableCell>
                    <TableCell>
                      <Badge variant="outline" class="rounded-full px-2 py-1">
                        <Icon 
                          v-if="tx.status?.toLowerCase() === 'success'" 
                          name="lucide:check-circle" 
                          class="mr-1 h-4 w-4 text-green-600" 
                        />
                        <Icon 
                          v-else-if="tx.status?.toLowerCase() === 'pending'" 
                          name="lucide:loader" 
                          class="mr-1 h-4 w-4 animate-spin text-muted-foreground" 
                        />
                        <Icon 
                          v-else-if="tx.status?.toLowerCase() === 'failed'" 
                          name="lucide:x-circle" 
                          class="mr-1 h-4 w-4 text-destructive" 
                        />
                        <Icon 
                          v-else 
                          name="lucide:help-circle" 
                          class="mr-1 h-4 w-4 text-muted-foreground" 
                        />
                        {{ tx.status || 'Unknown' }}
                      </Badge>
                    </TableCell>
                  </TableRow>
                </TableBody>
              </Table>              
            </div>
            <div class="flex items-center gap-2 mt-2">
                <PaginationSizeSelect :current-limit="limit" @update:limit="handleLimitChange" />
                <PaginationControls 
                  :offset="offset" 
                  :limit="limit" 
                  :has-more="hasMoreTransactions" 
                  @previous="handlePreviousPage"
                  @next="handleNextPage"
                />
            </div>
          </div>
        </div>
      </div>
    <div v-else class="p-8 text-center">
      <p class="text-lg text-muted-foreground">Select a wallet and token to view transactions</p>
    </div>
  </div>
</template> 