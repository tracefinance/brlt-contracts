<script setup lang="ts">
import { formatDistanceToNow } from 'date-fns'
import { computed } from 'vue'
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
  isLoading: isLoadingTransactions, 
  error: walletTransactionsError, 
  hasMore 
} = useWalletTransactions(chainType, address, tokenAddress, limit, offset)

// Use the useChains composable
const { chains, isLoading: isLoadingChains, error: chainsError } = useChains()

// Find the current chain based on the route parameter
const currentChain = computed(() => {
  if (isLoadingChains.value || chainsError.value) return null // Guard against loading/error state
  return chains.value.find((chain: IChain) => chain.type.toLowerCase() === chainType.value.toLowerCase())
})

// Computed property for the base explorer URL
const explorerBaseUrl = computed(() => {
  return currentChain.value?.explorerUrl
})

// Combine loading and error states
const isLoading = computed(() => isLoadingTransactions.value || isLoadingChains.value)
const error = computed(() => walletTransactionsError.value || chainsError.value)
</script>

<template>
  <div>
    <!-- Show loading state -->
    <div v-if="isLoading">
      <TableSkeleton />
    </div>

    <!-- Show error state -->
    <div v-else-if="error">
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
        <div>          
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
                    <TableHead class="w-[15%]">Hash</TableHead>
                    <TableHead class="w-[10%]">Type</TableHead>
                    <TableHead class="w-[15%]">From</TableHead>
                    <TableHead class="w-[15%]">To</TableHead>
                    <TableHead class="w-[11%]">Token</TableHead>
                    <TableHead class="w-[12%] text-right">Value</TableHead>
                    <TableHead class="w-[12%]">Age</TableHead>
                    <TableHead class="w-[10%]">Status</TableHead>
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
      </div>
    <div v-else class="p-8 text-center">
      <Alert>
        <Icon name="lucide:info" class="w-4 h-4" />
        <AlertTitle>Select Wallet and Token</AlertTitle>
        <AlertDescription>
          Please select a wallet and token from the sidebar to view transactions.
        </AlertDescription>
      </Alert>
    </div>
  </div>
</template> 