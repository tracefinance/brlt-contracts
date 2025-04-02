<script setup lang="ts">
import { computed, onMounted } from 'vue'
import type { Wallet } from '~/types/wallet'

// Get wallets functionality from composable
const { 
  wallets,
  currentWallet, 
  balances,
  isLoading,
  loadWallets, 
  loadWallet,
  loadWalletBalances
} = useWallets()

// Get route for active token
const route = useRoute()
const activeTokenAddress = computed(() => 
  typeof route.params.tokenAddress === 'string' 
    ? route.params.tokenAddress 
    : undefined
)

// Initialize with data
onMounted(async () => {
  await loadWallets()
  
  // If we have a wallet in the route, load it
  if (route.params.address && route.params.chainType) {
    const address = typeof route.params.address === 'string' ? route.params.address : route.params.address[0]
    const chainType = typeof route.params.chainType === 'string' ? route.params.chainType : route.params.chainType[0]
    
    await loadWallet(chainType, address)
    
    if (currentWallet.value) {
      await loadWalletBalances(currentWallet.value.chainType, currentWallet.value.address)
    }
  } else if (wallets.value.length > 0) {
    // If no wallet in route but we have wallets, load the first one
    const wallet = wallets.value[0]
    await loadWallet(wallet.chainType, wallet.address)
    await loadWalletBalances(wallet.chainType, wallet.address)
  }
})

// Handle wallet change
const handleWalletChange = async (wallet: Wallet) => {
  await loadWallet(wallet.chainType, wallet.address)
  await loadWalletBalances(wallet.chainType, wallet.address)
  
  // Navigate to the wallet's transactions page
  navigateTo(`/wallets/${wallet.address}/${wallet.chainType}/transactions`)
}
</script>

<template>
  <AppHeader />
  <div class="flex mt-16">
    <SidebarProvider>
      <WalletSidebar
        v-if="currentWallet"
        :wallets="wallets"
        :selected-wallet="currentWallet"
        :on-wallet-change="handleWalletChange"
        :balances="balances"
        :active-token-address="activeTokenAddress"
      />
      
      <SidebarInset>
        <header class="flex h-16 shrink-0 items-center gap-2 border-b px-4">
           <SidebarTrigger class="size-10 -ml-2" />
           <Breadcrumb>
            <BreadcrumbList>
              <BreadcrumbItem>
                {{ currentWallet?.name || 'Wallet' }}
              </BreadcrumbItem>
              <BreadcrumbSeparator/>
              <BreadcrumbItem>Transactions</BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
        </header>
        <div class="flex flex-1 flex-col gap-4">
          <slot />
        </div>
      </SidebarInset>
    </SidebarProvider>
  </div>
</template> 