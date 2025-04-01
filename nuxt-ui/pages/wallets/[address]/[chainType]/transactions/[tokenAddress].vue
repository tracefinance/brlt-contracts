<script setup lang="ts">
import { computed } from 'vue'

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
const { currentWallet, balances } = useWallets()

// Get current token details
const currentToken = computed(() => {
  if (!balances.value.length) return null
  
  return balances.value.find(balance => 
    (balance.token?.address || '').toLowerCase() === tokenAddress.value.toLowerCase()
  )?.token || null
})
</script>

<template>
  <div class="container p-8">
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
          <p class="text-muted-foreground">No transactions available for this token yet.</p>
        </div>
      </div>
    </div>
    <div v-else class="p-8 text-center">
      <p class="text-lg text-muted-foreground">Select a wallet and token to view transactions</p>
    </div>
  </div>
</template> 