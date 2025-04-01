<script setup lang="ts">
import { formatDistanceToNow } from 'date-fns'
import { computed, onMounted } from 'vue'
import { shortenAddress } from '~/lib/utils'

// Define page metadata
definePageMeta({
  layout: 'wallet'
})

// Get route params
const route = useRoute()
const address = computed(() => route.params.address as string)
const chainType = computed(() => route.params.chainType as string)
const tokenAddress = computed(() => route.params.tokenAddress as string)

// Get current wallet and its tokens
const { currentWallet, loadWallet, isLoading: isWalletLoading } = useWallets()

// Get token information
const { currentToken, loadToken, isLoading: isTokenLoading } = useTokens()

// Get transaction data
const { transactions, isLoading: isTransactionsLoading, error, getTransactionsByAddress } = useTransactions()

// Compute overall loading state
const isLoading = computed(() => isWalletLoading || isTransactionsLoading || isTokenLoading)

// Get token and transactions on component mount
onMounted(async () => {
  // First load the wallet data
  await loadWallet(chainType.value, address.value)
  
  // Load token data
  await loadToken(chainType.value, tokenAddress.value)
    
  // Finally, load transactions
  await loadTransactions()

  console.log('transactions', transactions.value)
  console.log('currentToken', currentToken.value)
  console.log('currentWallet', currentWallet.value)
})

// Fetch transactions function
async function loadTransactions() {
  await getTransactionsByAddress(
    chainType.value,
    address.value,
    10, // Default limit
    0,  // Default offset
    tokenAddress.value
  )
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
      <div class="flex items-center gap-4 mb-8">
        <Web3Icon :symbol="currentToken.symbol" size="32" />
        <div>
          <h1 class="text-3xl font-bold mb-1">{{ currentToken.name }}</h1>
          <p class="text-muted-foreground">{{ currentToken.symbol }}</p>
        </div>
      </div>
      
      <div class="border rounded-lg p-6 mb-8">
        <h2 class="text-xl font-semibold mb-4">Token Details</h2>
        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div>
            <p class="text-sm text-muted-foreground mb-2">Address</p>
            <p class="font-mono text-sm break-all">{{ currentToken.address }}</p>
          </div>
          <div>
            <p class="text-sm text-muted-foreground mb-2">Chain</p>
            <p class="flex items-center gap-2">
              <Web3Icon :symbol="chainType" size="16" />
              {{ chainType }}
            </p>
          </div>
          <div>
            <p class="text-sm text-muted-foreground mb-2">Decimals</p>
            <p>{{ currentToken.decimals }}</p>
          </div>
        </div>
      </div>
      
      <div class="rounded-lg border">
        <div class="p-6">
          <h2 class="text-xl font-semibold mb-4">Transactions</h2>
          
          <div v-if="isLoading" class="flex justify-center p-6">
            <Icon name="lucide:loader" class="animate-spin h-8 w-8 text-muted-foreground" />
          </div>
          
          <div v-else-if="error" class="p-4 bg-red-50 text-red-700 rounded-md">
            {{ error }}
          </div>
          
          <div v-else-if="transactions.length === 0" class="text-muted-foreground">
            No transactions available for this token yet.
          </div>
          
          <div v-else>
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
                    <Badge variant="outline" class="flex items-center">
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
                  <TableCell class="text-right">{{ tx.value }}</TableCell>
                  <TableCell :title="new Date(tx.timestamp * 1000).toLocaleString()">
                    {{ formatDistanceToNow(new Date(tx.timestamp * 1000), { addSuffix: true }) }}
                  </TableCell>
                  <TableCell>
                    <Badge variant="outline" class="flex items-center">
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
        </div>
      </div>
    </div>
    <div v-else class="p-8 text-center">
      <p class="text-lg text-muted-foreground">Select a wallet and token to view transactions</p>
    </div>
  </div>
</template> 