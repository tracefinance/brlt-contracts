<script setup lang="ts">
import { watch, ref } from 'vue'
import { toast } from 'vue-sonner'
import type { ChainType, IToken } from '~/types'

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
const { deleteToken, isDeleting, error: tokenMutationsError } = useTokenMutations()

const router = useRouter()
const route = useRoute()

// State for delete dialog
const isDeleteDialogOpen = ref(false)
const tokenToDelete = ref<IToken | null>(null)

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

function handleEditToken(token: IToken) {
  if (!token || !token.chainType || !token.address) {
    console.error('Invalid token data for edit navigation:', token)
    toast.error('Invalid token data. Cannot navigate to edit page.')
    return
  }
  router.push(`/settings/tokens/${token.chainType}/${token.address}/edit`)
}

function handleDeleteToken(token: IToken) {
  tokenToDelete.value = token
  isDeleteDialogOpen.value = true
}

async function handleDeleteConfirm() {
  if (!tokenToDelete.value || !tokenToDelete.value.address) {
    toast.error('Cannot delete token: Invalid data provided.')
    isDeleteDialogOpen.value = false
    tokenToDelete.value = null
    return
  }

  const { address, symbol } = tokenToDelete.value

  const success = await deleteToken(address)

  if (success) {
    toast.success(`Token ${symbol} deleted successfully.`)
    await refresh()
  } else {
    toast.error(`Failed to delete token ${symbol}`, {
      description: tokenMutationsError.value?.message || 'Unknown error'
    })
  }

  isDeleteDialogOpen.value = false
  tokenToDelete.value = null
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
       @edit="handleEditToken"
       @delete="handleDeleteToken"
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

    <!-- Delete Confirmation Dialog -->
    <AlertDialog :open="isDeleteDialogOpen" @update:open="isDeleteDialogOpen = $event">
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Are you absolutely sure?</AlertDialogTitle>
          <AlertDialogDescription>
            This action cannot be undone. This will permanently delete the token
            "{{ tokenToDelete?.symbol }}" ({{ tokenToDelete?.address }}) from the {{ tokenToDelete?.chainType }} chain.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel :disabled="isDeleting">Cancel</AlertDialogCancel>
          <AlertDialogAction 
            :disabled="isDeleting" 
            variant="destructive" 
            @click="handleDeleteConfirm"
          >
            <span v-if="isDeleting">Deleting...</span>
            <span v-else>Delete Token</span>
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>

  </div>
</template>