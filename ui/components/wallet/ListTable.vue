<script setup lang="ts">
import type { IWallet, IChain } from '~/types'
import { getAddressExplorerUrl } from '~/lib/explorers'
import { formatDateTime, shortenAddress, formatCurrency } from '~/lib/utils'

// Define Props
const props = defineProps<{
  wallets: IWallet[]
  chains: IChain[]
  isLoading: boolean
  isLoadingChains: boolean
}>()

// Define Emits - Combined Signature
const emit = defineEmits<{
  (e: 'edit' | 'delete', wallet: IWallet): void
}>()

// Helper functions moved from the page (or kept similar)
const getWalletExplorerBaseUrl = (wallet: IWallet): string | undefined => {
  if (props.isLoadingChains) return undefined
  const chain = props.chains.find(c => c.type?.toLowerCase() === wallet.chainType?.toLowerCase())
  return chain?.explorerUrl
}

const getNativeTokenSymbol = (wallet: IWallet): string => {
  if (props.isLoadingChains) return ''
  const chain = props.chains.find(c => c.type?.toLowerCase() === wallet.chainType?.toLowerCase())
  return chain?.symbol || ''
}

const handleEdit = (wallet: IWallet) => {
  emit('edit', wallet)
}

const handleDelete = (wallet: IWallet) => {
  emit('delete', wallet)
}
</script>

<template>
  <div class="border rounded-lg overflow-hidden">
    <Table>
      <TableHeader class="bg-muted">
        <!-- Header from TableSkeleton -->
        <TableRow>
          <TableHead class="w-[10%]">ID</TableHead>
          <TableHead class="w-auto">Name</TableHead>
          <TableHead class="w-[10%]">Chain</TableHead>
          <TableHead class="w-[10%]">Address</TableHead>
          <TableHead class="w-[10%] text-right">Balance</TableHead>
          <TableHead class="w-[15%]">Created</TableHead>
          <TableHead class="w-[80px] text-right">Actions</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <!-- Skeleton Rows from TableSkeleton -->
        <template v-if="isLoading">
          <TableRow v-for="n in 1" :key="`skeleton-${n}`">
            <TableCell><Skeleton class="h-4 w-20" /></TableCell>
            <TableCell><Skeleton class="h-4 w-24" /></TableCell>
            <TableCell>
              <div class="flex items-center gap-2">
                <Skeleton class="size-5 rounded-full" />
                <Skeleton class="h-4 w-16" />
              </div>
            </TableCell>
            <TableCell><Skeleton class="h-4 w-24" /></TableCell>
            <TableCell><Skeleton class="h-4 w-16 ml-auto" /></TableCell>
            <TableCell><Skeleton class="h-4 w-24" /></TableCell>
            <TableCell class="text-right">
              <Skeleton class="ml-auto size-8 rounded" />
            </TableCell>
          </TableRow>
        </template>
        <!-- Empty State from index.vue -->
        <TableRow v-else-if="wallets.length === 0">
          <TableCell colSpan="7" class="text-center pt-3 pb-4">
            <div class="flex items-center justify-center gap-1.5">
              <Icon name="lucide:inbox" class="size-5 text-primary" />
              <span>No wallets found. Create one to get started!</span>
            </div>
          </TableCell>
        </TableRow>
        <!-- Populated Rows from index.vue -->
        <template v-else>
          <TableRow v-for="wallet in wallets" :key="wallet.id">
            <TableCell>
              <NuxtLink :to="`/settings/wallets/${wallet.chainType}/${wallet.address}/view`" class="font-mono hover:underline">
                {{ shortenAddress(wallet.id, 4, 4) }}
              </NuxtLink>
            </TableCell>
            <TableCell>{{ wallet.name }}</TableCell>
            <TableCell>
              <div class="flex items-center gap-2">
                <Web3Icon :symbol="wallet.chainType" class="size-5" variant="branded" />
                <span class="capitalize">{{ wallet.chainType }}</span>
              </div>
            </TableCell>
            <TableCell>
              <a
                v-if="wallet.address && getWalletExplorerBaseUrl(wallet)"
                :href="getAddressExplorerUrl(getWalletExplorerBaseUrl(wallet)!, wallet.address)"
                target="_blank" rel="noopener noreferrer" class="font-mono hover:underline">
                {{ shortenAddress(wallet.address, 6, 4) }}
              </a>
              <span v-else-if="wallet.address" class="font-mono">{{ shortenAddress(wallet.address, 6, 4) }}</span>
              <span v-else class="text-muted-foreground">N/A</span>
            </TableCell>
            <TableCell class="text-right">
              <span v-if="wallet.balance !== undefined && wallet.balance !== null" class="font-mono">
                {{ formatCurrency(wallet.balance) }} {{ getNativeTokenSymbol(wallet) }}
              </span>
              <span v-else-if="isLoadingChains">
                <Icon name="svg-spinners:3-dots-fade" class="size-4 text-muted-foreground" />
              </span>
              <span v-else class="text-muted-foreground">N/A</span>
            </TableCell>
            <TableCell>{{ wallet.createdAt ? formatDateTime(wallet.createdAt) : 'N/A' }}</TableCell>
            <TableCell class="text-right">
              <DropdownMenu>
                <DropdownMenuTrigger as-child>
                  <Button variant="ghost" class="h-8 w-8 p-0">
                    <span class="sr-only">Open menu</span>
                    <Icon name="lucide:more-horizontal" class="size-4" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end">
                  <DropdownMenuItem :disabled="!wallet.chainType || !wallet.address" @click="handleEdit(wallet)">
                    <Icon name="lucide:edit" class="mr-2 size-4" />
                    <span>Edit</span>
                  </DropdownMenuItem>
                  <DropdownMenuItem
                    class="text-destructive focus:text-destructive focus:bg-destructive/10"
                    @click="handleDelete(wallet)">
                    <Icon name="lucide:trash-2" class="mr-2 size-4" />
                    <span>Delete</span>
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            </TableCell>
          </TableRow>
        </template>
      </TableBody>
    </Table>
  </div>
</template> 