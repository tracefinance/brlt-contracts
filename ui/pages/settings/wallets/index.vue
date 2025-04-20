<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { toast } from 'vue-sonner'
import { getAddressExplorerUrl } from '~/lib/explorers'
import type { IWallet } from '~/types'
import { formatDateTime, shortenAddress, formatCurrency } from '~/lib/utils'

definePageMeta({
  layout: 'settings'
})

const router = useRouter()
const route = useRoute()

const { limit, nextToken, setLimit, previousPage, nextPage } = usePagination(10)

const {
  wallets,
  isLoading,
  error: walletsError,
  nextPageToken,
  refresh: refreshWallets
} = useWalletsList(limit, nextToken)

const {
  chains,
  isLoading: isLoadingChains,
  error: chainsError
} = useChains()

const {
  deleteWallet,
  isDeleting,
  error: walletMutationsError
} = useWalletMutations()

const error = computed(() => walletsError.value || chainsError.value)

const getWalletExplorerBaseUrl = (wallet: IWallet): string | undefined => {
  if (isLoadingChains.value || chainsError.value) return undefined
  const chain = chains.value.find(c => c.type?.toLowerCase() === wallet.chainType?.toLowerCase())
  return chain?.explorerUrl
}

const getNativeTokenSymbol = (wallet: IWallet): string => {
  if (isLoadingChains.value || chainsError.value) return ''
  const chain = chains.value.find(c => c.type?.toLowerCase() === wallet.chainType?.toLowerCase())
  return chain?.symbol || ''
}

const isDeleteDialogOpen = ref(false)
const walletToDelete = ref<IWallet | null>(null)

const openDeleteDialog = (wallet: IWallet) => {
  walletToDelete.value = wallet
  isDeleteDialogOpen.value = true
}

const handleDeleteConfirm = async () => {
  if (!walletToDelete.value || !walletToDelete.value.chainType || !walletToDelete.value.address) {
    toast.error('Cannot delete wallet: Invalid data provided.')
    isDeleteDialogOpen.value = false
    walletToDelete.value = null
    return
  }

  const { chainType, address, name } = walletToDelete.value

  const success = await deleteWallet(chainType, address)

  if (success) {
    toast.success(`Wallet "${name}" deleted successfully.`)
    await refreshWallets()
  } else {
    toast.error(`Failed to delete wallet "${name}"`, {
      description: walletMutationsError.value?.message || 'Unknown error'
    })
  }

  isDeleteDialogOpen.value = false
  walletToDelete.value = null
}

const goToEditWallet = (wallet: IWallet) => {
  if (!wallet || !wallet.chainType || !wallet.address) {
    console.error('Invalid wallet data for edit navigation:', wallet)
    toast.error('Invalid wallet data. Cannot navigate to edit page.')
    return
  }

  const chainTypeEncoded = encodeURIComponent(wallet.chainType)
  const addressEncoded = encodeURIComponent(wallet.address)
  router.push(`/settings/wallets/${chainTypeEncoded}/${addressEncoded}/edit`)
}

// When we get a new token from the API, update the route
watch(nextPageToken, (newToken) => {
  if (newToken && newToken !== nextToken.value) {
    router.push({ 
      query: { 
        ...route.query, 
        next_token: newToken 
      } 
    })
  }
})

</script>

<template>
  <div>
    <WalletTableSkeleton v-if="isLoading" />
    <div v-else-if="error">
      <Alert variant="destructive">
        <Icon name="lucide:alert-triangle" class="w-4 h-4" />
        <AlertTitle>Error Loading Wallets</AlertTitle>
        <AlertDescription>
          {{ error.message || 'Failed to load wallets or chains' }}
        </AlertDescription>
      </Alert>
    </div>

    <div v-else-if="wallets.length === 0">
      <Alert>
        <Icon name="lucide:inbox" class="w-4 h-4" />
        <AlertTitle>No Wallets Found</AlertTitle>
        <AlertDescription>
          You haven't added any wallets yet. Create or import one!
        </AlertDescription>
      </Alert>
    </div>

    <div v-else>
      <div class="border rounded-lg overflow-hidden">
        <Table>
          <TableHeader class="bg-muted">
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
            <TableRow v-for="wallet in wallets" :key="wallet.id">
              <TableCell>
                <NuxtLink :to="`/settings/wallets/${wallet.chainType}/${wallet.address}/view`" class="hover:underline">
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
                  :href="getAddressExplorerUrl(getWalletExplorerBaseUrl(wallet), wallet.address)"
                  target="_blank" rel="noopener noreferrer" class="hover:underline">
                  {{ shortenAddress(wallet.address, 6, 4) }}
                </a>
                <span v-else-if="wallet.address">{{ shortenAddress(wallet.address, 6, 4) }}</span>
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
                    <DropdownMenuItem :disabled="!wallet.chainType || !wallet.address" @click="goToEditWallet(wallet)">
                      <Icon name="lucide:edit" class="mr-2 size-4" />
                      <span>Edit</span>
                    </DropdownMenuItem>
                    <DropdownMenuItem
                      class="text-destructive focus:text-destructive focus:bg-destructive/10"
                      @click="openDeleteDialog(wallet)">
                      <Icon name="lucide:trash-2" class="mr-2 size-4" />
                      <span>Delete</span>
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </div>
      <div class="flex items-center gap-2 mt-2">
        <PaginationSizeSelect :current-limit="limit" @update:limit="setLimit" />
        <PaginationControls 
          :next-token="nextPageToken" 
          :current-token="nextToken"
          @previous="previousPage"
          @next="nextPage(nextPageToken)" 
        />
      </div>
    </div>

    <AlertDialog :open="isDeleteDialogOpen" @update:open="isDeleteDialogOpen = $event">
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Are you absolutely sure?</AlertDialogTitle>
          <AlertDialogDescription>
            This action cannot be undone. This will permanently delete the wallet
            "{{ walletToDelete?.name }}" ({{ walletToDelete?.address?.substring(0, 6) }}...).
            Associated transaction data might also be affected depending on system configuration.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel :disabled="isDeleting" @click="isDeleteDialogOpen = false">Cancel</AlertDialogCancel>
          <AlertDialogAction :disabled="isDeleting" variant="destructive" @click="handleDeleteConfirm">
            <span v-if="isDeleting">Deleting...</span>
            <span v-else>Delete Wallet</span>
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  </div>
</template>
