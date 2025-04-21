<script setup lang="ts">
import type { IUser } from '~/types'
import { formatDateTime, shortenAddress } from '~/lib/utils'

// Define Props
defineProps<{
  users: IUser[]
  isLoading: boolean
}>()

// Define Emits
const emit = defineEmits<{
  (e: 'edit' | 'delete', user: IUser): void
}>()

const handleEdit = (user: IUser) => {
  emit('edit', user)
}

const handleDelete = (user: IUser) => {
  emit('delete', user)
}
</script>

<template>
  <div class="border rounded-lg overflow-hidden">
    <Table>
      <TableHeader class="bg-muted">
        <!-- Header from TableSkeleton -->
        <TableRow>
          <TableHead class="w-[10%]">ID</TableHead>
          <TableHead class="w-auto">Email</TableHead>
          <TableHead class="w-[15%]">Created</TableHead>
          <TableHead class="w-[80px] text-right">Actions</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <!-- Skeleton Rows from TableSkeleton -->
        <template v-if="isLoading">
          <TableRow v-for="n in 1" :key="`skeleton-${n}`">
            <TableCell><Skeleton class="h-4 w-20" /></TableCell>
            <TableCell><Skeleton class="h-4 w-40" /></TableCell>
            <TableCell><Skeleton class="h-4 w-24" /></TableCell>
            <TableCell class="text-right">
              <Skeleton class="ml-auto size-8 rounded" />
            </TableCell>
          </TableRow>
        </template>
        <!-- Empty State from index.vue -->
        <TableRow v-else-if="users.length === 0">
          <TableCell colSpan="4" class="text-center pt-3 pb-4">
            <div class="flex items-center justify-center gap-1.5">
              <Icon name="lucide:inbox" class="size-5 text-primary" />
              <span>No users found. Create one to get started!</span>
            </div>
          </TableCell>
        </TableRow>
        <!-- Populated Rows from index.vue -->
        <template v-else>
          <TableRow v-for="user in users" :key="user.id">
            <TableCell>
              <NuxtLink :to="`/settings/users/${user.id}/view`" class="font-mono hover:underline">
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
                  <DropdownMenuItem :disabled="!user.id" @click="handleEdit(user)">
                    <Icon name="lucide:edit" class="mr-2 size-4" />
                    <span>Edit</span>
                  </DropdownMenuItem>
                  <DropdownMenuItem
                    class="text-destructive focus:text-destructive focus:bg-destructive/10"
                    @click="handleDelete(user)">
                    <Icon name="lucide:trash-2" class="mr-2 size-4" />
                    <span>Delete</span>
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            </TableCell>
          </TableRow>
        </template>
      </TableBody>
    </Table>
  </div>
</template> 