<script setup lang="ts">
import { getAddressExplorerUrl } from '~/lib/explorers';
import type { ChainType, IToken } from '~/types';

defineProps<{
  tokens: IToken[]
  isLoading: boolean
  getExplorerUrl: (chainType: ChainType) => string | undefined
}>()
</script>

<template>
  <div class="border rounded-lg overflow-hidden">
    <Table>
      <TableHeader class="bg-muted">
        <TableRow>
          <TableHead class="w-auto">Address</TableHead>
          <TableHead class="w-[130px]">Chain</TableHead>
          <TableHead class="w-[130px]">Symbol</TableHead>
          <TableHead class="w-[80px]">Type</TableHead>
          <TableHead class="w-[80px] text-right">Decimals</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <template v-if="isLoading">
          <TableRow>
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
          </TableRow>
        </template>
        <TableRow v-else-if="tokens.length === 0">
          <TableCell colSpan="5" class="text-center pt-3 pb-4">
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
        </TableRow>
      </TableBody>
    </Table>
  </div>
</template> 