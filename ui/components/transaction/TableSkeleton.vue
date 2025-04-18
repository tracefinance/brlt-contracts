<script setup lang="ts">
defineProps({
  rows: {
    type: Number,
    default: 5
  }
})
</script>

<template>
  <div class="overflow-hidden rounded-lg border">
    <Table>
      <TableHeader>
        <TableRow class="bg-muted hover:bg-muted">
          <TableHead class="w-auto">Hash</TableHead>
            <TableHead class="w-[10%]">Type</TableHead>
            <TableHead class="w-[10%]">From</TableHead>
            <TableHead class="w-[10%]">To</TableHead>
            <TableHead class="w-[8%]">Token</TableHead>
            <TableHead class="w-[10%] text-right">Value</TableHead>
            <TableHead class="w-[15%]">Age</TableHead>
            <TableHead class="w-[10%] text-right">Status</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <TableRow v-for="i in rows" :key="i">
          <TableCell><Skeleton class="h-4 w-24" /></TableCell>
          <TableCell><Skeleton class="h-6 w-20 rounded-full" /></TableCell>
          <TableCell><Skeleton class="h-4 w-20" /></TableCell>
          <TableCell><Skeleton class="h-4 w-20" /></TableCell>
          <TableCell>
            <div class="flex items-center">
              <Skeleton class="h-5 w-5 mr-2 rounded-full" />
              <Skeleton class="h-4 w-12" />
            </div>
          </TableCell>
          <TableCell class="text-right"><Skeleton class="h-4 w-16 ml-auto" /></TableCell>
          <TableCell><Skeleton class="h-4 w-16" /></TableCell>
          <TableCell><Skeleton class="h-6 w-20 rounded-full" /></TableCell>
        </TableRow>
      </TableBody>
    </Table>              
  </div>
  <div class="flex items-center gap-2 mt-2">
    <Skeleton class="h-9 w-24" />
    <Skeleton class="size-9" />
    <Skeleton class="size-9" />
  </div>
</template> 