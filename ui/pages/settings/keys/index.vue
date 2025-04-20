<script setup lang="ts">
import { ref, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { toast } from 'vue-sonner'
import type { IKey } from '~/types'
import { formatDateTime, shortenAddress } from '~/lib/utils'

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

const openDeleteDialog = (key: IKey) => {
  keyToDelete.value = key
  isDeleteDialogOpen.value = true
}

const handleDeleteConfirm = async () => {
  if (!keyToDelete.value) {
    toast.error('Cannot delete key: Invalid data provided.')
    isDeleteDialogOpen.value = false
    keyToDelete.value = null
    return
  }

  const { id, name } = keyToDelete.value
  const success = await deleteKey(id)

  if (success) {
    toast.success(`Key "${name}" (ID: ${shortenAddress(id)}) deleted successfully.`)
    await refreshKeys()
  } else {
    toast.error(`Failed to delete key "${name}"`, {
      description: deleteError.value?.message || 'Unknown error'
    })
  }

  isDeleteDialogOpen.value = false
  keyToDelete.value = null
}

const goToEditKey = (key: IKey) => {
  router.push(`/settings/keys/${key.id}/edit`)
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
    <!-- Skeleton Loader -->
    <TableSkeleton v-if="isLoading" :columns="['ID', 'Name', 'Type', 'Curve', 'Created', 'Actions']" />

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
      <div class="border rounded-lg overflow-hidden">
        <Table>
          <TableHeader class="bg-muted">
            <TableRow>
              <TableHead class="w-[15%]">ID</TableHead>
              <TableHead class="w-auto">Name</TableHead>
              <TableHead class="w-[10%]">Type</TableHead>
              <TableHead class="w-[15%]">Curve</TableHead>
              <TableHead class="w-[15%]">Created</TableHead>
              <TableHead class="w-[80px] text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody v-if="keys.length === 0">
            <TableRow>
              <TableCell colSpan="6" class="text-center py-4">
                <div class="flex items-center justify-center gap-1.5">
                  <Icon name="lucide:key" class="size-5 text-primary" />
                  <span>No keys found. Create one to get started!</span>
                </div>
              </TableCell>
            </TableRow>
          </TableBody>
          <TableBody v-else>
            <TableRow v-for="key in keys" :key="key.id">
              <TableCell>
                <NuxtLink :to="`/settings/keys/${key.id}/view`" class="hover:underline font-mono">
                  {{ shortenAddress(key.id, 4, 4) }}
                </NuxtLink>
              </TableCell>
              <TableCell>{{ key.name }}</TableCell>
              <TableCell class="uppercase">{{ key.type }}</TableCell>
              <TableCell>{{ key.curve || 'N/A' }}</TableCell>
              <TableCell>{{ key.createdAt ? formatDateTime(key.createdAt) : 'N/A' }}</TableCell>
              <TableCell class="text-right">
                <DropdownMenu>
                  <DropdownMenuTrigger as-child>
                    <Button variant="ghost" class="h-8 w-8 p-0">
                      <span class="sr-only">Open menu</span>
                      <Icon name="lucide:more-horizontal" class="size-4" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end">
                    <DropdownMenuItem @click="goToEditKey(key)">
                      <Icon name="lucide:edit" class="mr-2 size-4" />
                      <span>Edit</span>
                    </DropdownMenuItem>
                    <DropdownMenuItem
                      class="text-destructive focus:text-destructive focus:bg-destructive/10"
                      @click="openDeleteDialog(key)">
                      <Icon name="lucide:trash-2" class="mr-2 size-4" />
                      <span>Delete Key</span>
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </div>

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