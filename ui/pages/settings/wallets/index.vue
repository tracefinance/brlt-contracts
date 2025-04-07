<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import type { IWallet, IPagedWallets } from '~/types/wallet'

definePageMeta({
  layout: 'settings'
})

// API client
const { $api } = useNuxtApp()

// Get router and route
const router = useRouter()
const route = useRoute()

// State
const walletsData = ref<IPagedWallets | null>(null)
const isLoading = ref(true)
const error = ref<string | null>(null)

// Get pagination parameters from route query
const limit = computed(() => {
  const queryLimit = route.query.limit ? Number(route.query.limit) : 10
  return isNaN(queryLimit) ? 10 : queryLimit
})

const offset = computed(() => {
  const queryOffset = route.query.offset ? Number(route.query.offset) : 0
  return isNaN(queryOffset) ? 0 : queryOffset
})

// Fetch wallets
const fetchWallets = async () => {
  isLoading.value = true
  error.value = null
  try {
    // Use reactive limit and offset
    walletsData.value = await $api.wallet.listWallets(limit.value, offset.value) 
  } catch (err) {
    console.error("Error fetching wallets:", err)
    error.value = err instanceof Error ? err.message : 'Failed to load wallets'
  } finally {
    isLoading.value = false
  }
}

// Initial fetch and watch for changes
watch([limit, offset], fetchWallets, { immediate: true })

// Computed wallets list
const wallets = computed<IWallet[]>(() => walletsData.value?.items || [])

// Compute hasMore based on API response
const hasMoreWallets = computed(() => walletsData.value?.hasMore || false)

// Handle page size change
function handleLimitChange(newLimit: number) {
  router.push({ 
    query: { 
      ...route.query, 
      limit: newLimit,
      offset: 0 // Reset offset when changing page size
    } 
  })
}

// Handle pagination events
function handlePreviousPage() {
  const newOffset = Math.max(0, offset.value - limit.value);
  router.push({ 
    query: { 
      ...route.query, 
      offset: newOffset 
    } 
  });
}

function handleNextPage() {
  const newOffset = offset.value + limit.value;
  router.push({ 
    query: { 
      ...route.query, 
      offset: newOffset 
    } 
  });
}

</script>

<template>
  <div class="flex flex-col">
    <div v-if="isLoading" class="text-center py-4">
      Loading wallets...
    </div>
    <div v-else-if="error" class="text-red-500 text-center py-4">
      Error: {{ error }}
    </div>
    <div v-else-if="wallets.length === 0" class="text-center py-4">
      No wallets found. Create one!
    </div>
    <div v-else class="border rounded-lg overflow-hidden">
      <Table>
        <TableHeader class="bg-muted">
          <TableRow>
            <TableHead>Name</TableHead>
            <TableHead>Chain</TableHead>
            <TableHead>Address</TableHead>
            <TableHead>Last Sync Block</TableHead>
            <TableHead>Tags</TableHead>
            <!-- <TableHead>Actions</TableHead> -->
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow v-for="wallet in wallets" :key="`${wallet.chainType}-${wallet.address}`">
            <TableCell class="font-medium">{{ wallet.name }}</TableCell>
            <TableCell class="flex items-center gap-2">
              <Web3Icon :symbol="wallet.chainType" class="size-5" variant="branded" />
              <span class="capitalize">{{ wallet.chainType }}</span>
            </TableCell>
            <TableCell class="font-mono text-xs">{{ wallet.address }}</TableCell>
            <TableCell>{{ wallet.lastBlockNumber || 'N/A' }}</TableCell>
            <TableCell>
               <div v-if="wallet.tags && Object.keys(wallet.tags).length > 0" class="flex flex-wrap gap-1">
                 <Badge v-for="(value, key) in wallet.tags" :key="key" variant="secondary">
                   {{ key }}: {{ value }}
                 </Badge>
               </div>
               <span v-else class="text-xs text-muted-foreground">No tags</span>
            </TableCell>
            <!-- <TableCell class="text-right">
              <Button variant="ghost" size="icon" @click="// TODO: Implement edit">
                <Icon name="lucide:pencil" class="h-4 w-4" />
              </Button>
              <Button variant="ghost" size="icon" @click="// TODO: Implement delete">
                 <Icon name="lucide:trash-2" class="h-4 w-4 text-red-500" />
              </Button>
            </TableCell> -->
          </TableRow>
        </TableBody>
      </Table>
      <!-- Add Pagination Controls -->      
    </div>
    <div class="flex items-center gap-2 mt-2">
        <PaginationSizeSelect :current-limit="limit" @update:limit="handleLimitChange" />
        <PaginationControls 
          :offset="offset" 
          :limit="limit" 
          :has-more="hasMoreWallets" 
          @previous="handlePreviousPage"
          @next="handleNextPage"
        />
      </div>
  </div>
</template>

<style scoped>
/* Add any specific styles if needed */
</style> 