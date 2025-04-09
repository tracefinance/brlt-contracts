<script setup lang="ts">
import { formatDistanceToNow } from 'date-fns'
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { formatCurrency, shortenAddress } from '~/lib/utils'
import { getTransactionExplorerUrl, getAddressExplorerUrl } from '~/lib/explorers'

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
  transactions, 
  isLoading, 
  error: walletTransactionsError, 
  hasMore 
} = useWalletTransactions(chainType, address, tokenAddress, limit, offset)

// Use the useChains composable
const { chains, isLoading: chainsLoading, error: chainsError } = useChains()

// Find the current chain based on the route parameter
const currentChain = computed(() => {
  if (chainsLoading.value || chainsError.value) return null // Guard against loading/error state
  return chains.value.find(chain => chain.type.toLowerCase() === chainType.value.toLowerCase())
})

// Computed property for the base explorer URL
const explorerBaseUrl = computed(() => {
  return currentChain.value?.explorerUrl
})

const error = computed(() => walletTransactionsError.value || chainsError.value)

// Handle pagination events
function handlePreviousPage() {
  const newOffset = Math.max(0, offset.value - limit.value)
  router.push({ query: { ...route.query, offset: newOffset } })
}

function handleNextPage() {
  const newOffset = offset.value + limit.value
  router.push({ query: { ...route.query, offset: newOffset } })
}

function handleLimitChange(newLimit: number) {
  router.push({ query: { ...route.query, limit: newLimit, offset: 0 } })
}
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
    <div v-else-if="currentChain">          
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
                      <a :href="getTransactionExplorerUrl(explorerBaseUrl, tx.hash)" target="_blank" rel="noopener noreferrer" class="hover:underline">
                        {{ shortenAddress(tx.hash, 6, 4) }}
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
                  :has-more="hasMore" 
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