<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import { toast } from 'vue-sonner'
import type { IUser } from '~/types'
import { formatDateTime, shortenAddress } from '~/lib/utils'

definePageMeta({
  layout: 'settings'
})

const router = useRouter()

const { limit, offset, setLimit, previousPage, nextPage } = usePagination(10)

const {
  users,
  isLoading,
  error: usersError,
  hasMore,
  refresh: refreshUsers
} = useUsersList(limit, offset)

const {
  deleteUser,
  isDeleting,
  error: userMutationsError
} = useUserMutations()

const error = computed(() => usersError.value)

const isDeleteDialogOpen = ref(false)
const userToDelete = ref<IUser | null>(null)

const openDeleteDialog = (user: IUser) => {
  userToDelete.value = user
  isDeleteDialogOpen.value = true
}

const handleDeleteConfirm = async () => {
  if (!userToDelete.value || !userToDelete.value.id) {
    toast.error('Cannot delete user: Invalid data provided.')
    isDeleteDialogOpen.value = false
    userToDelete.value = null
    return
  }

  const { id, email } = userToDelete.value

  const success = await deleteUser(id.toString())

  if (success) {
    toast.success(`User "${email}" deleted successfully.`)
    await refreshUsers()
  } else {
    toast.error(`Failed to delete user "${email}"`, {
      description: userMutationsError.value?.message || 'Unknown error'
    })
  }

  isDeleteDialogOpen.value = false
  userToDelete.value = null
}

const goToEditUser = (user: IUser) => {
  if (!user || !user.id) {
    console.error('Invalid user data for edit navigation:', user)
    toast.error('Invalid user data. Cannot navigate to edit page.')
    return
  }

  router.push(`/settings/users/${user.id}/edit`)
}
</script>

<template>
  <div>
    <UserTableSkeleton v-if="isLoading" />
    <div v-else-if="error">
      <Alert variant="destructive">
        <Icon name="lucide:alert-triangle" class="w-4 h-4" />
        <AlertTitle>Error Loading Users</AlertTitle>
        <AlertDescription>
          {{ error.message || 'Failed to load users' }}
        </AlertDescription>
      </Alert>
    </div>

    <div v-else-if="users.length === 0">
      <Alert>
        <Icon name="lucide:inbox" class="w-4 h-4" />
        <AlertTitle>No Users Found</AlertTitle>
        <AlertDescription>
          No users have been added yet.
        </AlertDescription>
      </Alert>
    </div>

    <div v-else>
      <div class="border rounded-lg overflow-hidden">
        <Table>
          <TableHeader class="bg-muted">
            <TableRow>
              <TableHead class="w-[10%]">ID</TableHead>
              <TableHead class="w-auto">Email</TableHead>
              <TableHead class="w-[15%]">Created</TableHead>
              <TableHead class="w-[80px] text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="user in users" :key="user.id">
              <TableCell>
                <NuxtLink :to="`/settings/users/${user.id}/view`" class="hover:underline">
                  {{ shortenAddress(user.id, 4, 4) }}
                </NuxtLink>
              </TableCell>
              <TableCell>{{ user.email }}</TableCell>
              <TableCell>{{ formatDateTime(user.createdAt) }}</TableCell>
              <TableCell class="text-right">
                <DropdownMenu>
                  <DropdownMenuTrigger as-child>
                    <Button variant="ghost" class="h-8 w-8 p-0">
                      <span class="sr-only">Open menu</span>
                      <Icon name="lucide:more-horizontal" class="size-4" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end">
                    <DropdownMenuItem :disabled="!user.id" @click="goToEditUser(user)">
                      <Icon name="lucide:edit" class="mr-2 size-4" />
                      <span>Edit</span>
                    </DropdownMenuItem>
                    <DropdownMenuItem
                      class="text-destructive focus:text-destructive focus:bg-destructive/10"
                      @click="openDeleteDialog(user)">
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
          :offset="offset" :limit="limit" :has-more="hasMore" @previous="previousPage"
          @next="nextPage" />
      </div>
    </div>

    <AlertDialog :open="isDeleteDialogOpen" @update:open="isDeleteDialogOpen = $event">
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Are you absolutely sure?</AlertDialogTitle>
          <AlertDialogDescription>
            This action cannot be undone. This will permanently delete the user
            "{{ userToDelete?.email }}".
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel :disabled="isDeleting" @click="isDeleteDialogOpen = false">Cancel</AlertDialogCancel>
          <AlertDialogAction :disabled="isDeleting" variant="destructive" @click="handleDeleteConfirm">
            <span v-if="isDeleting">Deleting...</span>
            <span v-else>Delete User</span>
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  </div>
</template> 