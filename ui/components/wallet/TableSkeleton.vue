<script setup lang="ts">

// Define props for controlling the number of skeleton rows
defineProps({
  rows: {
    type: Number,
    default: 3
  }
})
</script>

<template>
  <div class="overflow-hidden rounded-lg border">
    <Table>
      <TableHeader>
        <TableRow class="bg-muted hover:bg-muted">
          <TableHead class="w-[10%]">ID</TableHead>
          <TableHead class="w-auto">Name</TableHead>
          <TableHead class="w-[10%]">Chain</TableHead>
          <TableHead class="w-[10%]">Address</TableHead>
          <TableHead class="w-[10%] text-right">Balance</TableHead>
          <TableHead class="w-[15%]">Created</TableHead>
          <TableHead class="w-[80px] text-right">Actions</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <TableRow v-for="i in rows" :key="i">
          <TableCell><Skeleton class="h-4 w-20" /></TableCell>
          <TableCell><Skeleton class="h-4 w-24" /></TableCell>
          <TableCell>
            <div class="flex items-center gap-2">
              <Skeleton class="size-5 rounded-full" />
              <Skeleton class="h-4 w-16" />
            </div>
          </TableCell>
          <TableCell><Skeleton class="h-4 w-24" /></TableCell>
          <TableCell><Skeleton class="h-4 w-16 ml-auto" /></TableCell>
          <TableCell><Skeleton class="h-4 w-24" /></TableCell>
          <TableCell class="text-right">
            <Skeleton class="ml-auto size-8 rounded" />
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