<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { toast } from 'vue-sonner'
import type { IWallet } from '~/types'

definePageMeta({
  layout: 'settings'
})

const router = useRouter()
const route = useRoute()

const { limit, nextToken, setLimit, previousPage, nextPage } = usePagination(10)

const {
  wallets,
  isLoading: isLoadingWallets,
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
// const isLoading = computed(() => isLoadingWallets.value) // Unused

// Delete confirmation dialog
const isDeleteDialogOpen = ref(false)
const walletToDelete = ref<IWallet | null>(null)

// Helper functions moved to WalletListTable.vue
// const getWalletExplorerBaseUrl = (wallet: IWallet): string | undefined => { ... }
// const getNativeTokenSymbol = (wallet: IWallet): string => { ... }

// --- Handlers for events emitted by WalletListTable ---
const handleEditWallet = (wallet: IWallet) => {
  if (!wallet || !wallet.chainType || !wallet.address) {
    console.error('Invalid wallet data for edit navigation:', wallet)
    toast.error('Invalid wallet data. Cannot navigate to edit page.')
    return
  }

  const chainTypeEncoded = encodeURIComponent(wallet.chainType)
  const addressEncoded = encodeURIComponent(wallet.address)
  router.push(`/settings/wallets/${chainTypeEncoded}/${addressEncoded}/edit`)
}

const handleDeleteWallet = (wallet: IWallet) => {
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

// --- End Handlers ---

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
    <WalletTableSkeleton v-if="isLoadingWallets && !wallets.length" />
    <div v-else-if="error">
      <Alert variant="destructive">
        <Icon name="lucide:alert-triangle" class="w-4 h-4" />
        <AlertTitle>Error Loading Wallets</AlertTitle>
        <AlertDescription>
          {{ error.message || 'Failed to load wallets or chains' }}
        </AlertDescription>
      </Alert>
    </div>

    <div v-else>
      <WalletListTable
        :wallets="wallets"
        :chains="chains"
        :is-loading="isLoadingWallets" 
        :is-loading-chains="isLoadingChains"
        @edit="handleEditWallet"
        @delete="handleDeleteWallet"
      />
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
