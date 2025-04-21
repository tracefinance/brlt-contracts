<script setup lang="ts">
import type { ISigner } from '~/types'
import { formatDateTime, shortenAddress } from '~/lib/utils'

// Define Props
defineProps<{
  signers: ISigner[]
  isLoading: boolean
}>()

// Define Emits for actions - Combined signature
const emit = defineEmits<{
  (e: 'edit' | 'delete', signer: ISigner): void
}>()

const handleEdit = (signer: ISigner) => {
  emit('edit', signer)
}

const handleDelete = (signer: ISigner) => {
  emit('delete', signer)
}
</script>

<template>
  <div class="border rounded-lg overflow-hidden">
    <Table>
      <TableHeader class="bg-muted">
        <TableRow>
          <TableHead class="w-[10%]">ID</TableHead>
          <TableHead class="w-auto">Name</TableHead>
          <TableHead class="w-[10%]">Type</TableHead>
          <TableHead class="w-[10%]">User ID</TableHead>
          <TableHead class="w-[10%]">Addresses</TableHead>
          <TableHead class="w-[15%]">Created</TableHead>
          <TableHead class="w-[80px] text-right">Actions</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <!-- Skeleton rows -->
        <template v-if="isLoading">
          <TableRow v-for="n in 1" :key="`skeleton-${n}`">
            <TableCell><Skeleton class="h-4 w-20" /></TableCell>
            <TableCell><Skeleton class="h-4 w-32" /></TableCell>
            <TableCell><Skeleton class="h-6 w-20 rounded-full" /></TableCell>
            <TableCell><Skeleton class="h-4 w-20" /></TableCell>
            <TableCell><Skeleton class="h-6 w-6 rounded-md" /></TableCell>
            <TableCell><Skeleton class="h-4 w-24" /></TableCell>
            <TableCell class="text-right">
              <Skeleton class="ml-auto size-8 rounded" />
            </TableCell>
          </TableRow>
        </template>
        <!-- Empty state -->
        <TableRow v-else-if="signers.length === 0">
          <TableCell colSpan="7" class="text-center pt-3 pb-4">
            <div class="flex items-center justify-center gap-1.5">
              <Icon name="lucide:inbox" class="size-5 text-primary" />
              <span>No signers found. Create one to get started!</span>
            </div>
          </TableCell>
        </TableRow>
        <!-- Populated rows -->
        <template v-else>
          <TableRow v-for="signer in signers" :key="signer.id">
            <TableCell>
              <NuxtLink :to="`/settings/signers/${signer.id}/view`" class="font-mono hover:underline">
                {{ shortenAddress(signer.id, 4, 4) }}
              </NuxtLink>
            </TableCell>
            <TableCell>{{ signer.name }}</TableCell>
            <TableCell>
              <SignerTypeBadge :type="signer.type" />
            </TableCell>
            <TableCell>
              <NuxtLink
                v-if="signer.userId"
                :to="`/settings/users/${signer.userId}/view`"
                class="font-mono hover:underline"
              >
                {{ shortenAddress(signer.userId, 4, 4) }}
              </NuxtLink>
              <span v-else class="text-xs text-muted-foreground">Not assigned</span>
            </TableCell>
            <TableCell>
              <Badge v-if="signer.addresses?.length" variant="secondary" class="px-2 py-1">
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
                  <DropdownMenuItem @click="handleEdit(signer)">
                    <Icon name="lucide:edit" class="mr-2 size-4" />
                    <span>Edit</span>
                  </DropdownMenuItem>
                  <DropdownMenuItem
                    class="text-destructive focus:text-destructive focus:bg-destructive/10"
                    @click="handleDelete(signer)">
                    <Icon name="lucide:trash-2" class="mr-2 h-4 w-4" />
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