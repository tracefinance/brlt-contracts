<script setup lang="ts">
import { getAddressExplorerUrl } from '~/lib/explorers';
import type { ChainType, IToken } from '~/types';

defineProps<{
  tokens: IToken[]
  isLoading: boolean
  getExplorerUrl: (chainType: ChainType) => string | undefined
}>()

// Define Emits
const emit = defineEmits<{
  (e: 'edit', token: IToken): void
}>()

const handleEdit = (token: IToken) => {
  emit('edit', token)
}

</script>

<template>
  <div class="border rounded-lg overflow-hidden">
    <Table>
      <TableHeader class="bg-muted">
        <TableRow>
          <TableHead class="w-auto">Address</TableHead>
          <TableHead class="w-[12%]">Chain</TableHead>
          <TableHead class="w-[12%]">Symbol</TableHead>
          <TableHead class="w-[80px]">Type</TableHead>
          <TableHead class="w-[80px] text-right">Decimals</TableHead>
          <TableHead class="w-[80px] text-right">Actions</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <template v-if="isLoading">
          <TableRow v-for="n in 1" :key="`skeleton-${n}`">
            <TableCell>
              <Skeleton class="h-4 w-full" />
            </TableCell>
            <TableCell>
              <Skeleton class="h-4 w-[110px]" />
            </TableCell>
            <TableCell>
              <Skeleton class="h-4 w-[110px]" />
            </TableCell>
            <TableCell>
              <Skeleton class="h-4 w-[80px]" />
            </TableCell>
            <TableCell class="text-right">
              <Skeleton class="h-4 w-[60px] inline-block" />
            </TableCell>
            <TableCell class="text-right">
              <Skeleton class="ml-auto size-8 rounded" />
            </TableCell>
          </TableRow>
        </template>
        <TableRow v-else-if="tokens.length === 0">
          <TableCell colSpan="6" class="text-center pt-3 pb-4">
            <div class="flex items-center justify-center gap-1.5">
              <Icon name="lucide:inbox" class="size-5 text-primary" />
              <span>No tokens found.</span>
            </div>
          </TableCell>
        </TableRow>
        <TableRow v-for="token in tokens" v-else :key="token.address + token.chainType">
          <TableCell>
            <a
              v-if="token.address && getExplorerUrl(token.chainType)"
              :href="getAddressExplorerUrl(getExplorerUrl(token.chainType)!, token.address)"
              target="_blank" rel="noopener noreferrer"
              class="font-mono hover:underline"
            >
              {{ token.address }}
            </a>
            <span v-else class="font-mono">
              {{ token.address }}
            </span>
          </TableCell>
          <TableCell>
            <div class="capitalize flex items-center gap-1.5">
              <Web3Icon :symbol="token.chainType" variant="branded" class="size-5" />
              <span>{{ token.chainType }}</span>
            </div>
          </TableCell>
          <TableCell>
            <div class="flex items-center gap-1.5">
              <Web3Icon :symbol="token.symbol" variant="branded" class="size-5" />
              <span>{{ token.symbol }}</span>
            </div>
          </TableCell>
          <TableCell class="uppercase">{{ token.type }}</TableCell>
          <TableCell class="text-right">{{ token.decimals }}</TableCell>
          <TableCell class="text-right">
            <DropdownMenu>
              <DropdownMenuTrigger as-child>
                <Button variant="ghost" class="h-8 w-8 p-0">
                  <span class="sr-only">Open menu</span>
                  <Icon name="lucide:more-horizontal" class="size-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuItem @click="handleEdit(token)">
                  <Icon name="lucide:edit" class="mr-2 size-4" />
                  <span>Edit</span>
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </TableCell>
        </TableRow>
      </TableBody>
    </Table>
  </div>
</template> 