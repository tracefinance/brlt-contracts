<script setup lang="ts">
import { watch } from 'vue'
import type { ChainType } from '~/types'

definePageMeta({
  layout: 'settings',
})

const { limit, nextToken, setLimit, previousPage, nextPage } = usePagination()

const { chainTypeFilter, tokenTypeFilter, clearFilters } = useTokenFilters()

const { tokens, isLoading, error, refresh, nextToken: apiNextToken } = useTokensList(
  chainTypeFilter,
  tokenTypeFilter,
  limit,
  nextToken
)

const { chains, isLoading: isLoadingChains, error: chainsError } = useChains()

const router = useRouter()
const route = useRoute()

watch([chainTypeFilter, tokenTypeFilter], ([newChain, newType]) => {
  const currentQuery = { ...route.query }
  const newQuery: Record<string, string> = {}

  for (const key in currentQuery) {
    if (!['chain_type', 'token_type', 'next_token', 'limit'].includes(key)) {
      newQuery[key] = currentQuery[key] as string
    }
  }

  if (newChain) newQuery.chain_type = newChain
  if (newType) newQuery.token_type = newType

  router.push({ query: newQuery })
}, { deep: true })

function handleClearFilters() {
  clearFilters()
}

function handleLimitChange(newLimit: number) {
  setLimit(newLimit)
}

function handleNextPage() {
  nextPage(apiNextToken.value)
}

function handlePreviousPage() {
  previousPage()
}

function getChainExplorerUrl(chainType: ChainType): string | undefined {
  if (isLoadingChains.value || chainsError.value) return undefined
  const chain = chains.value.find(c => c.type?.toLowerCase() === chainType?.toLowerCase())
  return chain?.explorerUrl
}

</script>

<template>
  <div class="space-y-4">
    <div class="flex justify-between items-center">
      <TokenFilters
        :chain-type-filter="chainTypeFilter"
        :token-type-filter="tokenTypeFilter"
        @update:chain-type-filter="chainTypeFilter = $event"
        @update:token-type-filter="tokenTypeFilter = $event"
        @clear-filters="handleClearFilters"
      />
    </div>

    <Alert v-if="error" variant="destructive" class="flex items-start">
      <Icon name="lucide:alert-circle" class="h-5 w-5 flex-shrink-0 mr-2" />
      <div class="flex-grow">
        <AlertTitle>Error Fetching Tokens</AlertTitle>
        <AlertDescription>
          {{ error.message }}
          <Button variant="link" size="sm" class="p-0 h-auto ml-2" @click="refresh">Retry</Button>
        </AlertDescription>
      </div>
    </Alert>

    <TokenListTable 
       :tokens="tokens" 
       :is-loading="isLoading || isLoadingChains" 
       :get-explorer-url="getChainExplorerUrl"
    />

    <div class="flex items-center gap-2">
      <PaginationSizeSelect :current-limit="limit" @update:limit="handleLimitChange" />
      <PaginationControls
        :next-token="apiNextToken" 
        :current-token="nextToken"
        @previous="handlePreviousPage"
        @next="handleNextPage"
      />
    </div>

  </div>
</template>