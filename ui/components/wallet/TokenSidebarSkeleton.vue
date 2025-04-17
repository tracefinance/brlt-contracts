<script setup lang="ts">


defineProps({
  tokenCount: {
    type: Number,
    default: 5
  }
})
</script>

<template>
  <Sidebar class="mt-16">
    <SidebarHeader>
      <div class="p-2">
        <Skeleton class="h-10 w-full rounded-md" />
      </div>
    </SidebarHeader>
    <SidebarContent>
      <SidebarGroup>
        <SidebarGroupLabel>Tokens</SidebarGroupLabel>
        <SidebarMenu>
          <SidebarMenuItem v-for="i in tokenCount" :key="i">
            <SidebarMenuButton as-child>
              <div class="flex items-center w-full p-2">
                <Skeleton class="size-6 rounded-full mr-2" />
                <Skeleton class="h-4 w-16" />
                <Skeleton class="h-4 w-14 ml-auto" />
              </div>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarGroup>
    </SidebarContent>
  </Sidebar>
</template> 