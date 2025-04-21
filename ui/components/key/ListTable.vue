<script setup lang="ts">
import type { IKey } from '~/types'
import { formatDateTime, shortenAddress } from '~/lib/utils'

// Define Props
defineProps<{
  keys: IKey[]
  isLoading: boolean
}>()

// Define Emits
const emit = defineEmits<{
  (e: 'edit' | 'delete', key: IKey): void
}>()

const handleEdit = (key: IKey) => {
  emit('edit', key)
}

const handleDelete = (key: IKey) => {
  emit('delete', key)
}
</script>

<template>
  <div class="border rounded-lg overflow-hidden">
    <Table>
      <TableHeader class="bg-muted">
        <!-- Header from TableSkeleton -->
        <TableRow>
          <TableHead class="w-[10%]">ID</TableHead>
          <TableHead class="w-auto">Name</TableHead>
          <TableHead class="w-[10%]">Type</TableHead>
          <TableHead class="w-[15%]">Curve</TableHead>
          <TableHead class="w-[15%]">Created</TableHead>
          <TableHead class="w-[80px] text-right">Actions</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <!-- Skeleton Rows from TableSkeleton -->
        <template v-if="isLoading">
          <TableRow v-for="n in 1" :key="`skeleton-${n}`">
            <TableCell><Skeleton class="h-4 w-16" /></TableCell>
            <TableCell><Skeleton class="h-4 w-32" /></TableCell>
            <TableCell><Skeleton class="h-4 w-12" /></TableCell>
            <TableCell><Skeleton class="h-4 w-20" /></TableCell>
            <TableCell><Skeleton class="h-4 w-24" /></TableCell>
            <TableCell class="text-right">
              <Skeleton class="ml-auto size-8 rounded" />
            </TableCell>
          </TableRow>
        </template>
        <!-- Empty State from index.vue -->
        <TableRow v-else-if="keys.length === 0">
          <TableCell colSpan="6" class="text-center py-4">
            <div class="flex items-center justify-center gap-1.5">
              <Icon name="lucide:key" class="size-5 text-primary" />
              <span>No keys found. Create one to get started!</span>
            </div>
          </TableCell>
        </TableRow>
        <!-- Populated Rows from index.vue -->
        <template v-else>
          <TableRow v-for="key in keys" :key="key.id">
            <TableCell>
              <NuxtLink :to="`/settings/keys/${key.id}/view`" class="font-mono hover:underline">
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
                  <DropdownMenuItem @click="handleEdit(key)">
                    <Icon name="lucide:edit" class="mr-2 size-4" />
                    <span>Edit</span>
                  </DropdownMenuItem>
                  <DropdownMenuItem
                    class="text-destructive focus:text-destructive focus:bg-destructive/10"
                    @click="handleDelete(key)">
                    <Icon name="lucide:trash-2" class="mr-2 size-4" />
                    <span>Delete Key</span>
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