<script setup lang="ts">

// Define props for controlling the number of skeleton rows
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
          <TableHead class="w-[15%]">Name</TableHead>
          <TableHead class="w-[15%]">Chain</TableHead>
          <TableHead class="w-[30%]">Address</TableHead>
          <TableHead class="w-[15%]">Last Sync Block</TableHead>
          <TableHead class="w-[20%]">Tags</TableHead>
          <TableHead class="w-[5%] text-right">Actions</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <TableRow v-for="i in rows" :key="i">
          <TableCell><Skeleton class="h-4 w-24" /></TableCell>
          <TableCell>
            <div class="flex items-center gap-2">
              <Skeleton class="size-5 rounded-full" />
              <Skeleton class="h-4 w-10" />
            </div>
          </TableCell>
          <TableCell><Skeleton class="h-4 w-40" /></TableCell>
          <TableCell><Skeleton class="h-4 w-20" /></TableCell>
          <TableCell>
            <div class="flex flex-wrap gap-1">
              <Skeleton class="h-5 w-16 rounded" />
              <Skeleton class="h-5 w-20 rounded" />
            </div>
          </TableCell>
          <TableCell class="text-right">
            <Skeleton class="inline-block size-8" />
          </TableCell>
        </TableRow>
      </TableBody>
    </Table>              
  </div>
  <div class="flex items-center gap-2 mt-2">
    <Skeleton class="h-9 w-24" /> <!-- PaginationSizeSelect skeleton -->
    <Skeleton class="size-9" /> <!-- Previous button skeleton -->
    <Skeleton class="size-9" /> <!-- Next button skeleton -->
  </div>
</template> 