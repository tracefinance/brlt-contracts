<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import { toast } from 'vue-sonner'
import type { ISigner } from '~/types'

// Define page metadata
definePageMeta({
  layout: 'settings'
})

const router = useRouter()

// Use the pagination composable
const { limit, offset, setLimit, previousPage, nextPage } = usePagination(10)

// Use the signers list composable
const { signers, hasMore, isLoading, error, refresh } = useSignersList(limit, offset)
const { deleteSigner, isDeleting, error: signerMutationsError } = useSignerMutations()

// Delete confirmation dialog
const isDeleteDialogOpen = ref(false)
const signerToDelete = ref<ISigner | null>(null)

const openDeleteDialog = (signer: ISigner) => {
  signerToDelete.value = signer
  isDeleteDialogOpen.value = true
}

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

// Format date helper
const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleString()
}
</script>

<template>
  <div>
    <!-- Loading/Error/Empty States -->
    <div v-if="error">
      <Alert variant="destructive">
        <Icon name="lucide:alert-triangle" class="w-4 h-4" />
        <AlertTitle>Error Loading Signers</AlertTitle>
        <AlertDescription>
          {{ error.message || 'Failed to load signers' }}
        </AlertDescription>
      </Alert>
    </div>
    <div v-else-if="isLoading" class="flex justify-center p-6">
      <Spinner class="h-6 w-6" />
    </div>
    <div v-else-if="signers.length === 0">
      <Alert>
        <Icon name="lucide:inbox" class="w-4 h-4" />
        <AlertTitle>No Signers Found</AlertTitle>
        <AlertDescription>
          You haven't added any signers yet. Create one to get started!
        </AlertDescription>
      </Alert>
    </div>

    <!-- Signers Table -->
    <div v-else>
      <div class="border rounded-lg overflow-hidden">
        <Table>
          <TableHeader class="bg-muted">
            <TableRow>
              <TableHead class="w-[25%]">Name</TableHead>
              <TableHead class="w-[15%]">Type</TableHead>
              <TableHead class="w-[20%]">Created</TableHead>
              <TableHead class="w-[15%]">Addresses</TableHead>
              <TableHead class="w-[20%]">User ID</TableHead>
              <TableHead class="w-[5%] text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="signer in signers" :key="signer.id">
              <TableCell class="font-medium">{{ signer.name }}</TableCell>
              <TableCell>
                <Badge variant="secondary">{{ signer.type }}</Badge>
              </TableCell>
              <TableCell>{{ formatDate(signer.createdAt) }}</TableCell>
              <TableCell>
                <Badge variant="outline" v-if="signer.addresses?.length">
                  {{ signer.addresses.length }}
                </Badge>
                <span v-else class="text-xs text-muted-foreground">None</span>
              </TableCell>
              <TableCell>
                <span v-if="signer.userId">{{ signer.userId }}</span>
                <span v-else class="text-xs text-muted-foreground">Not assigned</span>
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
                    <DropdownMenuItem @click="router.push(`/settings/signers/${signer.id}`)">
                      <Icon name="lucide:eye" class="mr-2 size-4" />
                      <span>View</span>
                    </DropdownMenuItem>
                    <DropdownMenuItem @click="openDeleteDialog(signer)"
                      class="text-destructive focus:text-destructive focus:bg-destructive/10">
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
        <PaginationControls :offset="offset" :limit="limit" :has-more="hasMore" @previous="previousPage"
          @next="nextPage" />
      </div>
    </div>
  </div>

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
        <AlertDialogAction @click="handleDeleteConfirm" variant="destructive" :disabled="isDeleting">
          <span v-if="isDeleting">Deleting...</span>
          <span v-else>Delete Signer</span>
        </AlertDialogAction>
      </AlertDialogFooter>
    </AlertDialogContent>
  </AlertDialog>
</template> 