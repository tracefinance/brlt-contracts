<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import { toast } from 'vue-sonner'
import type { IUser } from '~/types'

definePageMeta({
  layout: 'settings'
})

const router = useRouter()

const { limit, nextToken, setLimit, previousPage, nextPage } = usePagination(10)

const {
  users,
  isLoading,
  error: usersError,
  nextPageToken,
  refresh: refreshUsers
} = useUsersList(limit, nextToken)

const {
  deleteUser,
  isDeleting,
  error: userMutationsError
} = useUserMutations()

const error = computed(() => usersError.value)

const isDeleteDialogOpen = ref(false)
const userToDelete = ref<IUser | null>(null)

const handleEditUser = (user: IUser) => {
  if (!user || !user.id) {
    console.error('Invalid user data for edit navigation:', user)
    toast.error('Invalid user data. Cannot navigate to edit page.')
    return
  }
  router.push(`/settings/users/${user.id}/edit`)
}

const handleDeleteUser = (user: IUser) => {
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
</script>

<template>
  <div>
    <UserTableSkeleton v-if="isLoading && !users.length" />
    <div v-else-if="error">
      <Alert variant="destructive">
        <Icon name="lucide:alert-triangle" class="w-4 h-4" />
        <AlertTitle>Error Loading Users</AlertTitle>
        <AlertDescription>
          {{ error.message || 'Failed to load users' }}
        </AlertDescription>
      </Alert>
    </div>

    <div v-else>
      <UserListTable 
        :users="users"
        :is-loading="isLoading"
        @edit="handleEditUser"
        @delete="handleDeleteUser"
      />
      <div class="flex items-center gap-2 mt-2">
        <PaginationSizeSelect :current-limit="limit" @update:limit="setLimit" />
        <PaginationControls 
          :next-token="nextPageToken"
          :current-token="nextToken"
          @previous="previousPage"
          @next="nextPage(nextPageToken)" />
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