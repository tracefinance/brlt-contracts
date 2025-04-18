<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { toast } from 'vue-sonner'
import type { ISigner } from '~/types'
// Import the new date formatting utility
import { formatDateTime } from '~/lib/utils'

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

// Navigate to edit signer page
const editSigner = (signer: ISigner) => {
  router.push(`/settings/signers/${signer.id}/edit`)
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
</script>

<template>
  <div>
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
                <TableHead class="w-[10%]">Type</TableHead>
                <TableHead class="w-[15%]">User ID</TableHead>
                <TableHead class="w-[15%]">Addresses</TableHead>
                <TableHead class="w-[25%]">Created</TableHead> 
                <TableHead class="w-[10%] text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow v-for="signer in signers" :key="signer.id">
                <TableCell class="font-medium">{{ signer.name }}</TableCell>
                <TableCell>
                  <SignerTypeBadge :type="signer.type" />
                </TableCell>
                <TableCell>
                  <span v-if="signer.userId">{{ signer.userId }}</span>
                  <span v-else class="text-xs text-muted-foreground">Not assigned</span>
                </TableCell>                
                <TableCell>
                  <Badge v-if="signer.addresses?.length" variant="outline">
                    {{ signer.addresses.length }}
                  </Badge>
                  <span v-else class="text-xs text-muted-foreground">None</span>
                </TableCell>
                <TableCell>{{ formatDateTime(signer.createdAt) }}</TableCell>
                <TableCell class="text-right">
                  <DropdownMenu>
                    <DropdownMenuTrigger as-child>
                      <Button variant="ghost" class="h-8 w-8 p-0">
                        <span class="sr-only">Open menu</span>
                        <Icon name="lucide:more-horizontal" class="h-4 w-4" />
                      </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end">
                      <DropdownMenuItem @click="editSigner(signer)">
                        <Icon name="lucide:pencil" class="mr-2 size-4" />
                        <span>Edit</span>
                      </DropdownMenuItem>
                      <DropdownMenuItem 
                        class="text-destructive focus:text-destructive focus:bg-destructive/10"
                        @click="openDeleteDialog(signer)">
                        <Icon name="lucide:trash-2" class="mr-2 h-4 w-4" />
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
          :offset="offset" :limit="limit" :has-more="hasMore" @previous="previousPage"
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
          <AlertDialogAction variant="destructive" :disabled="isDeleting" @click="handleDeleteConfirm">
            <span v-if="isDeleting">Deleting...</span>
            <span v-else>Delete Signer</span>
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  </div>
</template>