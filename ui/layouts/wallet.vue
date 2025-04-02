<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import type { IWallet, ITokenBalanceResponse } from '~/types'
import TokenSidebarSkeleton from '~/components/wallet/TokenSidebarSkeleton.vue'

// API client
const { $api } = useNuxtApp()

// State refs
const wallets = ref<IWallet[]>([])
const currentWallet = ref<IWallet | null>(null)
const balances = ref<ITokenBalanceResponse[]>([])

// Route handling
const route = useRoute()
const activeTokenAddress = computed(() => 
  typeof route.params.tokenAddress === 'string' 
    ? route.params.tokenAddress 
    : undefined
)

// Async data fetching
const { data: walletsData, status: walletsStatus } = useAsyncData(
  'wallets',
  () => $api.wallet.listWallets(10, 0),
  { immediate: true }
)

// Watch for wallets data and update local state
watch(walletsData, (newData) => {
  if (newData?.items) {
    wallets.value = newData.items
  }
})

// Fetch wallet data based on route or first wallet
const fetchCurrentWallet = async () => {
  if (route.params.address && route.params.chainType) {
    const address = typeof route.params.address === 'string' ? route.params.address : route.params.address[0]
    const chainType = typeof route.params.chainType === 'string' ? route.params.chainType : route.params.chainType[0]
    
    return { chainType, address }
  } else if (walletsData.value?.items && walletsData.value.items.length > 0) {
    const wallet = walletsData.value.items[0]
    return { chainType: wallet.chainType, address: wallet.address }
  }
  
  return null
}

// Fetch current wallet data
const { data: walletData, status: walletStatus } = useAsyncData(
  'currentWallet',
  async () => {
    const params = await fetchCurrentWallet()
    if (params) {
      return $api.wallet.getWallet(params.chainType, params.address)
    }
    return null
  },
  { watch: [walletsData] }
)

// Watch for wallet data and update current wallet
watch(walletData, (newData) => {
  if (newData) {
    currentWallet.value = newData
  }
})

// Fetch balances for current wallet
const { data: balanceData, status: balancesStatus } = useAsyncData(
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
  { watch: [() => currentWallet.value] }
)

// Watch for balance data and update balances
watch(balanceData, (newData) => {
  if (newData) {
    balances.value = newData
  }
})

// Combined loading state
const isLoading = computed(() => 
  walletsStatus.value === 'pending' || 
  walletStatus.value === 'pending' || 
  balancesStatus.value === 'pending'
)

// Handle wallet change
const handleWalletChange = async (wallet: IWallet) => {
  currentWallet.value = wallet
  
  // Navigate to the wallet's transactions page
  navigateTo(`/wallets/${wallet.chainType}/${wallet.address}/transactions`)
}
</script>

<template>
  <AppHeader />
  <div class="flex mt-16">
    <SidebarProvider>
      <TokenSidebarSkeleton v-if="isLoading" />
      
      <WalletTokenSidebar
        v-else-if="currentWallet"
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