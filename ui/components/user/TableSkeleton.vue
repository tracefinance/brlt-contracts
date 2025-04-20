<script setup lang="ts">
// No props needed anymore
</script>

<template>
  <div class="overflow-hidden rounded-lg border">
    <Table>
      <TableHeader>
        <TableRow class="bg-muted hover:bg-muted">
          <TableHead class="w-[10%]">ID</TableHead>
          <TableHead class="w-auto">Email</TableHead>
          <TableHead class="w-[15%]">Created</TableHead>
          <TableHead class="w-[80px] text-right">Actions</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <TableRow>
          <TableCell><Skeleton class="h-4 w-20" /></TableCell>
          <TableCell><Skeleton class="h-4 w-40" /></TableCell>
          <TableCell><Skeleton class="h-4 w-24" /></TableCell>
          <TableCell class="text-right">
            <Skeleton class="ml-auto size-8 rounded" />
          </TableCell>
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