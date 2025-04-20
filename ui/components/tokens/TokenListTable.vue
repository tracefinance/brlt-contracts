<script setup lang="ts">
import type { IToken } from '~/types'
import {
  Table,
  TableBody,
  TableCaption,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '~/components/ui/table'
import { Skeleton } from '~/components/ui/skeleton'

// Define props
defineProps<{
  tokens: IToken[]
  isLoading: boolean
}>()

// Function to truncate address for display
function truncateAddress(address: string): string {
  if (!address || address.length < 10) return address
  return `${address.substring(0, 6)}...${address.substring(address.length - 4)}`
}
</script>

<template>
  <Table>
    <TableCaption v-if="!isLoading && tokens.length === 0">
      No tokens found.
    </TableCaption>
    <TableHeader>
      <TableRow>
        <TableHead class="w-[250px]">Address</TableHead>
        <TableHead>Chain</TableHead>
        <TableHead>Symbol</TableHead>
        <TableHead>Type</TableHead>
        <TableHead class="text-right">Decimals</TableHead>
        <!-- <TableHead class="text-right">Actions</TableHead> -->
      </TableRow>
    </TableHeader>
    <TableBody>
      <template v-if="isLoading">
        <!-- Skeleton loader rows -->
        <TableRow v-for="n in 5" :key="`skel-${n}`">
          <TableCell>
            <Skeleton class="h-4 w-full" />
          </TableCell>
          <TableCell>
            <Skeleton class="h-4 w-[80px]" />
          </TableCell>
          <TableCell>
            <Skeleton class="h-4 w-[60px]" />
          </TableCell>
          <TableCell>
            <Skeleton class="h-4 w-[60px]" />
          </TableCell>
          <TableCell class="text-right">
            <Skeleton class="h-4 w-[30px] inline-block" />
          </TableCell>
          <!-- <TableCell class="text-right">
            <Skeleton class="h-8 w-[80px] inline-block" />
          </TableCell> -->
        </TableRow>
      </template>
      <template v-else>
        <TableRow v-for="token in tokens" :key="token.address + token.chainType">
          <TableCell class="font-mono">{{ truncateAddress(token.address) }}</TableCell>
          <TableCell class="capitalize">{{ token.chainType }}</TableCell>
          <TableCell>{{ token.symbol }}</TableCell>
          <TableCell class="uppercase">{{ token.type }}</TableCell>
          <TableCell class="text-right">{{ token.decimals }}</TableCell>
          <!-- <TableCell class="text-right">
             <Button variant="outline" size="sm">View</Button>
          </TableCell> -->
        </TableRow>
      </template>
    </TableBody>
  </Table>
</template> 