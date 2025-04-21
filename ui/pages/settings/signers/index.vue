<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { toast } from 'vue-sonner'
import type { ISigner } from '~/types'
// Import the new date formatting utility and shortenAddress
// Removed unused imports (formatDateTime, shortenAddress)

// Define page metadata
definePageMeta({
  layout: 'settings'
})

const router = useRouter()

// Use the pagination composable with token-based pagination
const { limit, nextToken, setLimit, previousPage, nextPage } = usePagination(10)

// Use the signers list composable with token-based pagination
const { 
  signers, 
  nextPageToken, 
  isLoading, 
  error, 
  refresh 
} = useSignersList(limit, nextToken)
const { deleteSigner, isDeleting, error: signerMutationsError } = useSignerMutations()

// Delete confirmation dialog
const isDeleteDialogOpen = ref(false)
const signerToDelete = ref<ISigner | null>(null)

// --- Handler functions for events emitted by SignerListTable ---
const handleEditSigner = (signer: ISigner) => {
  router.push(`/settings/signers/${signer.id}/edit`)
}

const handleDeleteSigner = (signer: ISigner) => {
  signerToDelete.value = signer
  isDeleteDialogOpen.value = true
}
// --- End handler functions ---

// Navigate to edit signer page (renamed to avoid conflict, handled by handleEditSigner)
// const editSigner = (signer: ISigner) => {
//   router.push(`/settings/signers/${signer.id}/edit`)
// }

// Open delete dialog (renamed to avoid conflict, handled by handleDeleteSigner)
// const openDeleteDialog = (signer: ISigner) => {
//   signerToDelete.value = signer
//   isDeleteDialogOpen.value = true
// }

const handleDeleteConfirm = async () => {
  if (!signerToDelete.value || !signerToDelete.value.id) {
    toast.error('Cannot delete signer: Invalid data provided.')
    isDeleteDialogOpen.value = false
    signerToDelete.value = null
    return
  }

  const { id, name } = signerToDelete.value

  const success = await deleteSigner(id.toString())

  if (success) {
    toast.success(`Signer "${name}" deleted successfully.`)
    await refresh()
  } else {
    toast.error(`Failed to delete signer "${name}"`, {
      description: signerMutationsError.value?.message || 'Unknown error'
    })
  }

  isDeleteDialogOpen.value = false
  signerToDelete.value = null
}
</script>

<template>
  <div>
    <div>
      <SignerTableSkeleton v-if="isLoading && !signers.length"/>
      <!-- Error States -->
      <div v-else-if="error">
        <Alert variant="destructive">
          <Icon name="lucide:alert-triangle" class="w-4 h-4" />
          <AlertTitle>Error Loading Signers</AlertTitle>
          <AlertDescription>
            {{ error.message || 'Failed to load signers' }}
          </AlertDescription>
        </Alert>
      </div>
      <!-- Use the SignerListTable component -->
      <div v-else>
        <SignerListTable 
          :signers="signers"
          :is-loading="isLoading"
          @edit="handleEditSigner"
          @delete="handleDeleteSigner"
        />
        <!-- Pagination controls remain here -->
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
    </div>

    <!-- Delete confirmation dialog remains here -->
    <AlertDialog :open="isDeleteDialogOpen" @update:open="isDeleteDialogOpen = $event">
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Are you absolutely sure?</AlertDialogTitle>
          <AlertDialogDescription>
            This action cannot be undone. This will permanently delete the signer
            "{{ signerToDelete?.name }}". Associated addresses will also be deleted.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel :disabled="isDeleting" @click="isDeleteDialogOpen = false">Cancel</AlertDialogCancel>
          <AlertDialogAction variant="destructive" :disabled="isDeleting" @click="handleDeleteConfirm">
            <span v-if="isDeleting">Deleting...</span>
            <span v-else>Delete Signer</span>
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  </div>
</template>