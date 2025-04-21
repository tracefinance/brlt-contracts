<script setup lang="ts">
import { ref, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { toast } from 'vue-sonner'
import type { IKey } from '~/types'
import { shortenAddress } from '~/lib/utils'
// Removed unused imports (formatDateTime)

definePageMeta({
  layout: 'settings'
})

const router = useRouter()
const route = useRoute()

// Pagination
const { limit, nextToken, setLimit, previousPage, nextPage } = usePagination(10)

// Data Fetching
const {
  keys,
  isLoading,
  error: keysError,
  nextPageToken,
  refresh: refreshKeys
} = useKeysList(limit, nextToken)

// Mutations
const {
  deleteKey,
  isDeleting,
  deleteError
} = useKeyMutations()

const isDeleteDialogOpen = ref(false)
const keyToDelete = ref<IKey | null>(null)

// --- Handlers for events emitted by KeyListTable ---
const handleEditKey = (key: IKey) => {
  router.push(`/settings/keys/${key.id}/edit`)
}

const handleDeleteKey = (key: IKey) => {
  keyToDelete.value = key
  isDeleteDialogOpen.value = true
}
// --- End Handlers ---

const handleDeleteConfirm = async () => {
  if (!keyToDelete.value) {
    toast.error('Cannot delete key: Invalid data provided.')
    isDeleteDialogOpen.value = false
    keyToDelete.value = null
    return
  }

  const { id, name } = keyToDelete.value
  // Need to use shortenAddress here for the toast message
  const shortId = shortenAddress(id)
  const success = await deleteKey(id)

  if (success) {
    toast.success(`Key "${name}" (ID: ${shortId}) deleted successfully.`)
    await refreshKeys()
  } else {
    toast.error(`Failed to delete key "${name}"`, {
      description: deleteError.value?.message || 'Unknown error'
    })
  }

  isDeleteDialogOpen.value = false
  keyToDelete.value = null
}

// Update route query when next page token changes
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
    <!-- Skeleton Loader - show only on initial load -->
    <KeyTableSkeleton v-if="isLoading && !keys.length" />

    <!-- Error Alert -->
    <div v-else-if="keysError">
      <Alert variant="destructive">
        <Icon name="lucide:alert-triangle" class="w-4 h-4" />
        <AlertTitle>Error Loading Keys</AlertTitle>
        <AlertDescription>
          {{ keysError.message || 'Failed to load keys' }}
        </AlertDescription>
      </Alert>
    </div>

    <!-- Key Table -->
    <div v-else>
      <!-- Use the KeyListTable component -->
      <KeyListTable
        :keys="keys"
        :is-loading="isLoading"
        @edit="handleEditKey"
        @delete="handleDeleteKey"
      />

      <!-- Pagination Controls -->
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

    <!-- Delete Confirmation Dialog -->
    <AlertDialog :open="isDeleteDialogOpen" @update:open="isDeleteDialogOpen = $event">
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Are you absolutely sure?</AlertDialogTitle>
          <AlertDialogDescription>
            This action cannot be undone. This will permanently delete the key
            "{{ keyToDelete?.name }}" (ID: {{ shortenAddress(keyToDelete?.id || '') }}).
            This key will no longer be usable for signing or other cryptographic operations.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel :disabled="isDeleting" @click="isDeleteDialogOpen = false">Cancel</AlertDialogCancel>
          <AlertDialogAction :disabled="isDeleting" variant="destructive" @click="handleDeleteConfirm">
            <span v-if="isDeleting">Deleting...</span>
            <span v-else>Delete Key</span>
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  </div>
</template> 