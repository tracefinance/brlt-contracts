<script setup lang="ts">
import { computed } from 'vue'
import type { ITokenBalanceResponse, IWallet } from '~/types'

// API client
const { $api } = useNuxtApp()

// Route handling
const route = useRoute()
const activeTokenAddress = computed(() => 
  typeof route.params.tokenAddress === 'string' 
    ? route.params.tokenAddress 
    : undefined
)

// Async data fetching
const { data: walletsData, status: walletsStatus } = await useAsyncData(
  'wallets',
  () => $api.wallet.listWallets(10, 0),
  { 
    immediate: true,
  }
)

// Make sure walletsData is always treated as an array
const wallets = computed<IWallet[]>(() => walletsData.value?.items || [])

// Fetch wallet data based on route or first wallet
const fetchCurrentWallet = async () => {
  if (route.params.address && route.params.chainType) {
    const address = typeof route.params.address === 'string' ? route.params.address : route.params.address[0]
    const chainType = typeof route.params.chainType === 'string' ? route.params.chainType : route.params.chainType[0]
    
    return { chainType, address }
  } else if (walletsData.value && walletsData.value.items.length > 0) {
    const wallet = walletsData.value.items[0]
    return { chainType: wallet.chainType, address: wallet.address }
  }
  
  return null
}

// Fetch current wallet data
const { data: currentWallet, status: walletStatus } = await useAsyncData(
  'currentWallet',
  async () => {
    const params = await fetchCurrentWallet()
    if (params) {
      return $api.wallet.getWallet(params.chainType, params.address)
    }
    return null
  },
  { 
    watch: [walletsData, route]
  }
)

// Fetch balances for current wallet
const { data: balancesData, status: balancesStatus } = await useAsyncData(
  'walletBalances',
  async () => {
    if (currentWallet.value) {
      return $api.wallet.getWalletBalance(
        currentWallet.value.chainType, 
        currentWallet.value.address
      )
    }
    return [] as ITokenBalanceResponse[]
  },
  { 
    watch: [currentWallet]
  }
)

// Make sure balances is always treated as an array
const balances = computed<ITokenBalanceResponse[]>(() => balancesData.value || [])

// Combined loading state
const isLoading = computed(() => 
  walletsStatus.value === 'pending' || 
  walletStatus.value === 'pending' || 
  balancesStatus.value === 'pending'
)

// Handle wallet change
const handleWalletChange = async (wallet: IWallet) => {
  // Navigate to the wallet's transactions page
  // This will trigger the watcher to update the wallet data
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