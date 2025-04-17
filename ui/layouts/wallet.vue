<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { toast } from 'vue-sonner'
import type { IToken, IWallet } from '~/types'

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
const { currentWallet, isLoading: isLoadingCurrentWallet } = useWalletDetails(targetChainType, targetAddress)

// Use the wallet balances composable
const { balances, isLoading: isLoadingBalances, refresh: refreshBalances } = useWalletBalances(targetChainType, targetAddress)

// Set up auto-refresh interval for balances
let balanceRefreshInterval: ReturnType<typeof setInterval> | null = null

onMounted(() => {
  balanceRefreshInterval = setInterval(() => {
    // Only refresh if we have a target wallet context
    if (targetChainType.value && targetAddress.value) {
      refreshBalances()
    }
  }, 3000) // Refresh every 3 seconds
})

onUnmounted(() => {
  if (balanceRefreshInterval) {
    clearInterval(balanceRefreshInterval)
  }
})

// Use the wallet mutations composable
const { error: tokenActivationError, activateToken } = useWalletMutations()

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

// Handle token activation with notifications
const handleTokenActivation = async (token: IToken) => {
  if (!targetChainType.value || !targetAddress.value) {
    toast.warning('Cannot activate token: Wallet context is missing.')
    return
  }

  await activateToken(targetChainType.value, targetAddress.value, token.address)

  if (tokenActivationError.value) {
    const errorMessage = tokenActivationError.value.message || 'An unknown error occurred'
    toast.error(`Failed to activate token: ${errorMessage}`)
    return
  }

  setTimeout(() => {
    toast.success(`Token ${token.symbol || token.address} activated successfully!`)
    refreshBalances()
  }, 100)
}

// copyAddress function removed
</script>

<template>
  <div>
    <AppHeader />
    <div class="flex mt-16">
      <SidebarProvider>
        <WalletTokenSidebar 
          v-if="currentWallet" :wallets="wallets" :selected-wallet="currentWallet"
          :on-wallet-change="handleWalletChange" :balances="balances" :active-token-address="activeTokenAddress"
          :is-loading="isLoading" :on-token-activation="handleTokenActivation" />

        <SidebarInset>
          <header class="flex h-16 shrink-0 items-center gap-2 border-b px-4">
            <SidebarTrigger class="size-10 -ml-2" />
            <Breadcrumb>
              <BreadcrumbList>
                <BreadcrumbItem>
                  {{ currentWallet?.name || 'Wallet' }}
                </BreadcrumbItem>
                <BreadcrumbSeparator />
                <BreadcrumbItem>Transactions</BreadcrumbItem>
              </BreadcrumbList>
            </Breadcrumb>
            <!-- Add Spacer and Receive Button -->
            <div class="ml-auto flex items-center gap-2">
              <!-- Use the new WalletReceiveModal component -->
              <WalletReceiveModal :current-wallet="currentWallet" />
            </div>
          </header>
          <div class="flex flex-1 flex-col gap-4 p-4">
            <slot />
          </div>
        </SidebarInset>
      </SidebarProvider>
    </div>
  </div>
</template>
