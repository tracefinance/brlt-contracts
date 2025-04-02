<script setup lang="ts">
import { formatDistanceToNow } from 'date-fns'
import { computed } from 'vue'
import { shortenAddress, formatCurrency } from '~/lib/utils'

// Define page metadata
definePageMeta({
  layout: 'wallet'
})

// Get API client
const { $api } = useNuxtApp()

// Get route params
const route = useRoute()
const router = useRouter()

const address = computed(() => route.params.address as string)
const chainType = computed(() => route.params.chainType as string)
const tokenAddress = computed(() => route.params.tokenAddress as string)

// Get pagination parameters from route query
const limit = computed(() => {
  const queryLimit = route.query.limit ? Number(route.query.limit) : 10
  return isNaN(queryLimit) ? 10 : queryLimit
})

const offset = computed(() => {
  const queryOffset = route.query.offset ? Number(route.query.offset) : 0
  return isNaN(queryOffset) ? 0 : queryOffset
})

// Fetch wallet data
const { data: currentWallet, status: walletStatus } = await useAsyncData(
  'currentWallet',
  () => $api.wallet.getWallet(chainType.value, address.value),
  {
    watch: [chainType, address]
  }
)

// Fetch token data
const { data: currentToken, status: tokenStatus } = await useAsyncData(
  'currentToken',
  () => $api.token.getToken(chainType.value, tokenAddress.value),
  {
    watch: [chainType, tokenAddress]
  }
)

// Fetch transactions
const { data: transactionsData, status: transactionsStatus } = await useAsyncData(
  'transactions',
  () => $api.transaction.getWalletTransactions(
    address.value,
    chainType.value,
    limit.value,
    offset.value,
    tokenAddress.value
  ),
  {
    watch: [chainType, address, tokenAddress, limit, offset]
  }
)

// Extract transactions and pagination info
const transactions = computed(() => transactionsData.value?.items || [])
const hasMoreTransactions = computed(() => {
  if (!transactionsData.value) return false
  return transactionsData.value.hasMore || false
})

// Handle errors
const error = computed(() => {
  if (walletStatus.value === 'error') return 'Failed to load wallet data'
  if (tokenStatus.value === 'error') return 'Failed to load token data'
  if (transactionsStatus.value === 'error') return 'Failed to load transactions'
  return null
})

// Compute overall loading state
const isLoading = computed(() => 
  walletStatus.value === 'pending' || 
  tokenStatus.value === 'pending' || 
  transactionsStatus.value === 'pending'
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
    <div v-if="currentWallet && currentToken">          
        <div class="p-4">          
          <div v-if="error" class="p-4 bg-red-50 text-red-700 rounded-md">
            {{ error }}
          </div>
          
          <div v-else-if="transactions.length === 0" class="text-muted-foreground">
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