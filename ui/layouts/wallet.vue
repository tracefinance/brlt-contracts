<script setup lang="ts">
import { ref, computed } from 'vue'
import type { IWallet } from '~/types'

// Route handling
const route = useRoute()

// Reactive pagination state for wallet list (optional, can be static)
const walletLimit = ref(100) // Fetch a larger number if needed for the sidebar
const walletOffset = ref(0)

// Use the wallets list composable
const { wallets, isLoading: isLoadingWallets } = useWalletsList(walletLimit, walletOffset)

// Determine target chainType and address from route or first wallet
const targetChainType = computed(() => {
  if (route.params.chainType) {
    return route.params.chainType as string
  }
  // Fallback to the first wallet in the list if no route params
  return wallets.value.length > 0 ? wallets.value[0].chainType : undefined
})

const targetAddress = computed(() => {
  if (route.params.address) {
    return route.params.address as string
  }
  // Fallback to the first wallet in the list if no route params
  return wallets.value.length > 0 ? wallets.value[0].address : undefined
})

// Use the current wallet composable
const { currentWallet, isLoading: isLoadingCurrentWallet } = useCurrentWallet(targetChainType, targetAddress)

// Use the wallet balances composable
const { balances, isLoading: isLoadingBalances } = useWalletBalances(targetChainType, targetAddress)

// Active token address from route for highlighting in sidebar
const activeTokenAddress = computed(() => 
  route.params.tokenAddress as string | undefined
)

// Combined loading state
const isLoading = computed(() => 
  isLoadingWallets.value || isLoadingCurrentWallet.value || isLoadingBalances.value
)

// Handle wallet change (navigation triggers data refresh via route params)
const handleWalletChange = (wallet: IWallet) => {
  navigateTo(`/wallets/${wallet.chainType}/${wallet.address}/transactions`)
}
</script>

<template>
  <AppHeader />
  <div class="flex mt-16">
    <SidebarProvider>      
      <WalletTokenSidebar
        v-if="currentWallet"
        :wallets="wallets"
        :selected-wallet="currentWallet"
        :on-wallet-change="handleWalletChange"
        :balances="balances"
        :active-token-address="activeTokenAddress"
        :is-loading="isLoading" 
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
        <div class="flex flex-1 flex-col gap-4 p-4">
          <slot />
        </div>
      </SidebarInset>
    </SidebarProvider>
  </div>
</template> 