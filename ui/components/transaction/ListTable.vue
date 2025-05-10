<script setup lang="ts">
import { formatDistanceToNow } from 'date-fns'
import { formatCurrency, shortenAddress } from '~/lib/utils'
import { getTransactionExplorerUrl, getAddressExplorerUrl } from '~/lib/explorers'
import type { ITransaction } from '~/types'

defineProps<{
  transactions: ITransaction[]
  isLoading: boolean
  walletAddress: string
  explorerBaseUrl?: string
  nativeTokenSymbol?: string
  hasInitiallyLoaded?: boolean
  rows?: number
}>()

</script>

<template>
  <div class="overflow-auto rounded-lg border">
    <Table>
      <TableHeader>
        <TableRow class="bg-muted hover:bg-muted">
          <TableHead class="w-auto">Hash</TableHead>
          <TableHead class="w-[10%]">Type</TableHead>
          <TableHead class="w-[10%]">From</TableHead>
          <TableHead class="w-[10%]">To</TableHead>
          <TableHead class="w-[8%]">Token</TableHead>
          <TableHead class="w-[10%] text-right">Value</TableHead>
          <TableHead class="w-[15%]">Age</TableHead>
          <TableHead class="w-[110px]">Status</TableHead>
        </TableRow>
      </TableHeader>
      
      <!-- Skeleton loading state -->
      <TableBody v-if="isLoading && !hasInitiallyLoaded">
        <TableRow v-for="n in (rows || 3)" :key="`skeleton-${n}`">
          <TableCell><Skeleton class="h-4 w-24" /></TableCell>
          <TableCell><Skeleton class="h-[1.6rem] w-20 rounded-full" /></TableCell>
          <TableCell><Skeleton class="h-4 w-20" /></TableCell>
          <TableCell><Skeleton class="h-4 w-20" /></TableCell>
          <TableCell>
            <div class="flex items-center">
              <Skeleton class="h-5 w-5 mr-2 rounded-full" />
              <Skeleton class="h-4 w-12" />
            </div>
          </TableCell>
          <TableCell class="text-right"><Skeleton class="h-4 w-16 ml-auto" /></TableCell>
          <TableCell><Skeleton class="h-4 w-16" /></TableCell>
          <TableCell><Skeleton class="h-[1.6rem] w-20 rounded-full" /></TableCell>
        </TableRow>
      </TableBody>
      
      <!-- Empty State Table Body -->
      <TableBody v-else-if="transactions.length === 0">
        <TableRow>
          <TableCell colSpan="8" class="text-center py-3">
            <div class="flex items-center justify-center gap-1.5">
              <Icon name="lucide:inbox" class="size-5 text-primary" />
              <span>No transactions found for this token.</span>
            </div>
          </TableCell>
        </TableRow>
      </TableBody>
      
      <!-- Populated Table Body -->
      <TableBody v-else>
        <TableRow v-for="tx in transactions" :key="tx.hash">
          <TableCell>
            <a :href="getTransactionExplorerUrl(explorerBaseUrl, tx.hash)" target="_blank" rel="noopener noreferrer" class="hover:underline">
              {{ shortenAddress(tx.hash) }}
            </a>
          </TableCell>
          <TableCell>
            <TransactionTypeBadge :wallet-address="walletAddress" :from-address="tx.fromAddress" />
          </TableCell>
          <TableCell>
            <a :href="getAddressExplorerUrl(explorerBaseUrl, tx.fromAddress)" target="_blank" rel="noopener noreferrer" class="hover:underline">
              {{ shortenAddress(tx.fromAddress) }}
            </a>
          </TableCell>
          <TableCell>
            <a :href="getAddressExplorerUrl(explorerBaseUrl, tx.toAddress)" target="_blank" rel="noopener noreferrer" class="hover:underline">
              {{ shortenAddress(tx.toAddress) }}
            </a>
          </TableCell>
          <TableCell class="flex items-center">
            <div class="flex items-center gap-2">
              <!-- Show token symbol if available, otherwise use native token if it's a native transaction -->
              <Web3Icon v-if="tx.tokenSymbol" :symbol="tx.tokenSymbol" class="size-5" />
              <Web3Icon v-else-if="nativeTokenSymbol" :symbol="nativeTokenSymbol" class="size-5" />
              <Icon v-else name="lucide:help-circle" class="size-5 text-muted-foreground" />
              {{ tx.tokenSymbol || nativeTokenSymbol || 'N/A' }}
            </div>
          </TableCell>
          <TableCell class="text-right font-mono">{{ formatCurrency(tx.value) }}</TableCell>
          <TableCell>
            {{ formatDistanceToNow(new Date(tx.timestamp * 1000), { addSuffix: true }) }}
          </TableCell>
          <TableCell>
            <TransactionStatusBadge :status="tx.status" />
          </TableCell>
        </TableRow>
      </TableBody>
    </Table>
  </div>
  
  <!-- Pagination skeleton when loading -->
  <div v-if="isLoading && !hasInitiallyLoaded" class="flex items-center gap-2 mt-2">
    <Skeleton class="h-9 w-24" />
    <Skeleton class="size-9" />
    <Skeleton class="size-9" />
  </div>
</template> 