<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import { toast } from 'vue-sonner'
import { getAddressExplorerUrl } from '~/lib/explorers'
import type { IWallet } from '~/types'

definePageMeta({
  layout: 'settings'
})

const router = useRouter()

const { limit, offset, setLimit, previousPage, nextPage } = usePagination(10)

const { 
  wallets, 
  isLoading: isLoadingWallets, 
  error: walletsError,       
  hasMore, 
  refresh: refreshWallets
} = useWalletsList(limit, offset)

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

const isLoading = computed(() => isLoadingWallets.value || isLoadingChains.value)

const error = computed(() => walletsError.value || chainsError.value || walletMutationsError.value)

const getWalletExplorerBaseUrl = (wallet: IWallet): string | undefined => {
  if (isLoadingChains.value || chainsError.value) return undefined
  const chain = chains.value.find(c => c.type?.toLowerCase() === wallet.chainType?.toLowerCase())
  return chain?.explorerUrl
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
    toast.success(`Wallet \"${name}\" deleted successfully.`)
    await refreshWallets()
  } else {
    toast.error(`Failed to delete wallet \"${name}\"": ${walletMutationsError.value?.message || 'Unknown error'}`)
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

</script>

<template>
  <div class="flex flex-col">
    <div v-if="isLoading">
      <WalletTableSkeleton />
    </div>
    
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

    <div v-else class="border rounded-lg overflow-hidden">
      <Table>
        <TableHeader class="bg-muted">
          <TableRow>
            <TableHead class="w-[15%]">Name</TableHead>
            <TableHead class="w-[15%]">Chain</TableHead>
            <TableHead class="w-[30%]">Address</TableHead>
            <TableHead class="w-[15%]">Last Sync Block</TableHead>
            <TableHead class="w-[20%]">Tags</TableHead>
            <TableHead class="w-[5%] text-right">Actions</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow v-for="wallet in wallets" :key="wallet.id || `${wallet.chainType}-${wallet.address}`">
            <TableCell class="font-medium">{{ wallet.name }}</TableCell>
            <TableCell>
              <div class="flex items-center gap-2">
                <Web3Icon :symbol="wallet.chainType" class="size-5" variant="branded" />
                <span class="capitalize">{{ wallet.chainType }}</span>
              </div>
            </TableCell>
            <TableCell>
              <a
                :href="getAddressExplorerUrl(getWalletExplorerBaseUrl(wallet), wallet.address)"
                target="_blank"
                rel="noopener noreferrer"
                class="hover:underline"
                v-if="wallet.address && getWalletExplorerBaseUrl(wallet)"
              >
                {{ wallet.address }}
              </a>
              <span v-else-if="wallet.address">{{ wallet.address }}</span>
              <span v-else class="text-muted-foreground">N/A</span>
            </TableCell>
            <TableCell>{{ wallet.lastBlockNumber || 'N/A' }}</TableCell>
            <TableCell>
               <div v-if="wallet.tags && Object.keys(wallet.tags).length > 0" class="flex flex-wrap gap-1">
                 <Badge v-for="(value, key) in wallet.tags" :key="key" variant="secondary" class="whitespace-nowrap">
                   {{ key }}: {{ value }}
                 </Badge>
               </div>
               <span v-else class="text-xs text-muted-foreground">No tags</span>
            </TableCell>
            <TableCell class="text-right">
              <DropdownMenu>
                <DropdownMenuTrigger as-child>
                  <Button variant="ghost" class="h-8 w-8 p-0">
                    <span class="sr-only">Open menu</span>
                    <Icon name="lucide:more-horizontal" class="size-4" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end">
                  <DropdownMenuItem @click="goToEditWallet(wallet)" :disabled="!wallet.chainType || !wallet.address">
                    <Icon name="lucide:pencil" class="mr-2 size-4" />
                    <span>Edit</span>
                  </DropdownMenuItem>
                  <DropdownMenuItem @click="openDeleteDialog(wallet)" class="text-destructive focus:text-destructive focus:bg-destructive/10">
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
        :offset="offset" 
        :limit="limit" 
        :has-more="hasMore" 
        @previous="previousPage"
        @next="nextPage"
      />
    </div>

    <AlertDialog :open="isDeleteDialogOpen" @update:open="isDeleteDialogOpen = $event">
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Are you absolutely sure?</AlertDialogTitle>
          <AlertDialogDescription>
            This action cannot be undone. This will permanently delete the wallet 
            "{{ walletToDelete?.name }}" ({{ walletToDelete?.address?.substring(0,6) }}...). 
            Associated transaction data might also be affected depending on system configuration.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel :disabled="isDeleting" @click="isDeleteDialogOpen = false">Cancel</AlertDialogCancel>
          <AlertDialogAction @click="handleDeleteConfirm" variant="destructive" :disabled="isDeleting">
            <span v-if="isDeleting">Deleting...</span>
            <span v-else>Delete Wallet</span>
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>

  </div>
</template>

<style scoped>
/* Add any specific styles if needed */
</style> 