<script setup lang="ts">
import { computed, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
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
const { limit, nextToken, setLimit, previousPage, nextPage } = usePagination(10)

// Use composables for data fetching
const { 
  transactions, 
  isLoading, 
  hasInitiallyLoaded,
  error: walletTransactionsError, 
  nextPageToken,
  refresh
} = useWalletTransactions(chainType, address, tokenAddress, limit, nextToken)

// Use the useChains composable
const { chains, isLoading: isLoadingChains, error: chainsError } = useChains()

// Use the useNativeTokens composable
const { nativeTokens, isLoading: isLoadingNativeTokens, error: nativeTokensError } = useNativeTokens()

// Find the current chain based on the route parameter
const currentChain = computed(() => {
  if (isLoadingChains.value || chainsError.value) return null // Guard against loading/error state
  return chains.value.find((chain: IChain) => chain.type?.toLowerCase() === chainType.value?.toLowerCase())
})

// Find native token for current chain
const nativeToken = computed(() => {
  if (isLoadingNativeTokens.value || !currentChain.value) return null
  return nativeTokens.value.find(token => token.chainType?.toLowerCase() === chainType.value?.toLowerCase())
})

// Computed property for the base explorer URL
const explorerBaseUrl = computed(() => {
  return currentChain.value?.explorerUrl
})

// Combine loading and error states
const error = computed(() => walletTransactionsError.value || chainsError.value || nativeTokensError.value)

// Set up auto-refresh interval
let refreshInterval: ReturnType<typeof setInterval> | null = null

onMounted(() => {
  refreshInterval = setInterval(() => {
    refresh()
  }, 3000) // Refresh every 3 seconds
})

onUnmounted(() => {
  if (refreshInterval) {
    clearInterval(refreshInterval)
  }
})
</script>

<template>
  <div>
    <!-- Show error state -->
    <div v-if="error">
      <Alert variant="destructive">
        <Icon name="lucide:alert-triangle" class="w-4 h-4" />
        <AlertTitle>Error</AlertTitle>
        <AlertDescription>
          {{ error.message || 'Failed to load data' }}
        </AlertDescription>
      </Alert>
    </div>

    <!-- Show content only after initial load attempt -->
    <div>
      <TransactionListTable
        :transactions="transactions"
        :is-loading="isLoading"
        :has-initially-loaded="hasInitiallyLoaded"
        :wallet-address="address"
        :explorer-base-url="explorerBaseUrl"
        :native-token-symbol="nativeToken?.symbol"
        :rows="3"
      />
      
      <!-- Only show pagination controls when not in loading state or after initial load -->
      <div v-if="!isLoading || hasInitiallyLoaded" class="flex items-center gap-2 mt-2">
        <PaginationSizeSelect :current-limit="limit" @update:limit="setLimit" />
        <PaginationControls 
          :next-token="nextPageToken" 
          :current-token="nextToken"
          @previous="previousPage"
          @next="nextPage(nextPageToken)"
        />
      </div>
    </div>
  </div>
</template>
