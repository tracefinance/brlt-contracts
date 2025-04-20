<script setup lang="ts">
import { ref, watch } from 'vue'
// import type { ChainType, TokenType } from '~/types' // Types used implicitly via composables
import { Button } from '~/components/ui/button'
import { Alert, AlertTitle, AlertDescription } from '~/components/ui/alert' // Import Alert components
import { AlertCircle } from 'lucide-vue-next' // Import icon

// Settings
definePageMeta({
  layout: 'settings', // Assuming a common layout for settings/admin pages
})

// State
const limit = ref(15) // Items per page
const nextTokenInternal = ref<string | undefined>(undefined)
const router = useRouter()
const route = useRoute()

// Composables
const { chainTypeFilter, tokenTypeFilter, clearFilters } = useTokenFilters() // Manages filter state and reads initial URL
const { tokens, isLoading, error, refresh, hasMore, nextToken: apiNextToken } = useTokensList(
  chainTypeFilter,
  tokenTypeFilter,
  limit,
  nextTokenInternal // Pass internal pagination state to list composable
)

// URL Syncing: Watch filters managed by useTokenFilters
watch([chainTypeFilter, tokenTypeFilter], ([newChain, newType]) => {
  const currentQuery = { ...route.query }
  const newQuery: Record<string, string> = {}

  // Preserve other query params if they exist
  for (const key in currentQuery) {
    if (key !== 'chain_type' && key !== 'token_type' && key !== 'next_token') {
      newQuery[key] = currentQuery[key] as string
    }
  }

  // Add new filter values
  if (newChain) newQuery.chain_type = newChain
  if (newType) newQuery.token_type = newType

  // Reset pagination when filters change
  nextTokenInternal.value = undefined

  router.push({ query: newQuery })
}, { deep: true }) // Deep watch might be needed depending on filter structure

// Pagination: Update internal ref when API provides next token
watch(apiNextToken, (newApiNextToken) => {
  // Only update if the API actually provided a token (prevents infinite loops on last page)
  if (newApiNextToken !== nextTokenInternal.value) {
    nextTokenInternal.value = newApiNextToken
  }
})

function loadMore() {
  // The useTokensList composable re-fetches automatically when nextTokenInternal changes.
  // We need to update the URL based on the *next* token we will request,
  // which is the one we just received from the API.
  const nextTokenForUrl = apiNextToken.value
  if (hasMore.value && nextTokenForUrl) {
    // Update URL immediately to reflect the state *before* the next fetch starts
    router.push({ query: { ...route.query, next_token: nextTokenForUrl } })
    // Update the internal state to trigger the fetch by useTokensList watcher
    nextTokenInternal.value = nextTokenForUrl
  } else {
     // Clear next_token from URL if no more pages
     const query = { ...route.query }
     delete query.next_token
     router.push({ query })
     // Ensure internal state is also cleared if API confirms no more tokens
     if (!hasMore.value) {
       nextTokenInternal.value = undefined
     }
  }
}

function handleClearFilters() {
  clearFilters() // This triggers the watcher above via reactivity, which updates URL
}

</script>

<template>
  <div class="space-y-6">
    <PageHeader title="Token Management" description="Manage supported blockchain tokens." />

    <div class="flex justify-between items-center">
      <TokenFilters
        :chain-type-filter="chainTypeFilter"
        :token-type-filter="tokenTypeFilter"
        @update:chain-type-filter="chainTypeFilter = $event"
        @update:token-type-filter="tokenTypeFilter = $event"
        @clear-filters="handleClearFilters"
      />
      <NuxtLink to="/admin/tokens/new">
        <Button>Add Token</Button>
      </NuxtLink>
    </div>

    <Alert v-if="error" variant="destructive" class="flex items-start">
      <AlertCircle class="h-5 w-5 flex-shrink-0 mr-2" />
      <div class="flex-grow">
        <AlertTitle>Error Fetching Tokens</AlertTitle>
        <AlertDescription>
          {{ error.message }}
          <Button variant="link" size="sm" class="p-0 h-auto ml-2" @click="refresh">Retry</Button>
        </AlertDescription>
      </div>
    </Alert>

    <TokenListTable :tokens="tokens" :is-loading="isLoading" />

    <div v-if="hasMore && !isLoading" class="text-center mt-6">
      <Button variant="outline" @click="loadMore">
        Load More
      </Button>
    </div>
     <div v-if="isLoading && tokens.length > 0" class="text-center mt-6 text-muted-foreground">
      Loading...
    </div>
  </div>
</template> 