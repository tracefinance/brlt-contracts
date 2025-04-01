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

// Get current wallet
const { currentWallet, balances } = useWallets()
</script>

<template>
  <div class="container p-8">
    <div v-if="currentWallet">
      <div class="flex items-center gap-4 mb-8">
        <div class="flex aspect-square size-10 items-center justify-center rounded-md bg-primary/10">
          <Web3Icon :symbol="chainType" size="28" />
        </div>
        <div>
          <h1 class="text-3xl font-bold mb-1">{{ currentWallet.name }}</h1>
          <p class="text-muted-foreground font-mono text-sm">{{ currentWallet.address }}</p>
        </div>
      </div>
      
      <div class="border rounded-lg p-6 mb-8">
        <h2 class="text-xl font-semibold mb-4">Wallet Overview</h2>
        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div>
            <p class="text-sm text-muted-foreground mb-2">Chain Type</p>
            <p class="flex items-center gap-2">
              <Web3Icon :symbol="chainType" size="16" />
              {{ chainType }}
            </p>
          </div>
          <div>
            <p class="text-sm text-muted-foreground mb-2">Tokens</p>
            <p>{{ balances.length }} tokens</p>
          </div>
        </div>
      </div>
      
      <div class="rounded-lg border">
        <div class="p-6">
          <h2 class="text-xl font-semibold mb-4">Recent Transactions</h2>
          <p class="text-muted-foreground">No transactions available for this wallet yet.</p>
          
          <p class="mt-4 text-sm text-muted-foreground">Select a token from the sidebar to view token-specific transactions.</p>
        </div>
      </div>
    </div>
    <div v-else class="p-8 text-center">
      <p class="text-lg text-muted-foreground">Select a wallet to view transactions</p>
    </div>
  </div>
</template> 