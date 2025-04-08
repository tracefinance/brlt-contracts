<script setup lang="ts">
import { computed } from 'vue'
import AppHeader from '~/components/AppHeader.vue'
// Import Sidebar components
import {
  Sidebar,
  SidebarHeader,
  SidebarContent,
  SidebarGroup,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuItem,
  SidebarMenuButton
} from '~/components/ui/sidebar' // Assuming path, adjust if needed
import { Button } from '~/components/ui/button' // Import Button

// Route handling
const route = useRoute()

// TODO: Define navigation items based on settings sections
const navigationItems = [
  { name: 'Wallets', href: '/settings/wallets', icon: 'lucide:wallet' } 
  // Add other settings sections here
]

const isCurrentRoute = (href: string) => {
  return route.path.startsWith(href)
}
</script>

<template>
  <AppHeader />
  <div>
    <SidebarProvider>
      <!-- Refactored Sidebar using custom components -->
      <Sidebar class="mt-16">       
        <SidebarContent>
          <SidebarGroup>
             <SidebarGroupLabel>General</SidebarGroupLabel>
            <SidebarMenu>
              <SidebarMenuItem v-for="item in navigationItems" :key="item.name">
                <SidebarMenuButton 
                  :is-active="isCurrentRoute(item.href)"
                  as-child
                >
                  <NuxtLink :to="item.href" class="flex items-center gap-3 w-full">
                     <Icon :name="item.icon" class="h-4 w-4" />
                     <span>{{ item.name }}</span>
                  </NuxtLink>
                </SidebarMenuButton>
              </SidebarMenuItem>
            </SidebarMenu>
          </SidebarGroup>
          <!-- Add more SidebarGroups for other setting categories if needed -->
        </SidebarContent>
        <!-- Optional: Add SidebarFooter if needed -->
      </Sidebar>
      
      <SidebarInset>
        <header class="flex h-16 shrink-0 items-center gap-2 border-b px-4 mt-16">
           <SidebarTrigger class="size-10 -ml-2" />
           <!-- Breadcrumb can be dynamic based on the current settings page -->
           <Breadcrumb>
            <BreadcrumbList>
              <BreadcrumbItem>Settings</BreadcrumbItem>
              <!-- Add more items based on route -->
               <template v-if="route.path.startsWith('/settings/wallets')">
                 <BreadcrumbSeparator/>
                 <BreadcrumbItem>
                   <!-- Make Wallets a link only if not on the specific subpage -->
                   <NuxtLink v-if="route.path !== '/settings/wallets'" to="/settings/wallets">Wallets</NuxtLink>
                   <span v-else>Wallets</span>
                 </BreadcrumbItem>
               </template>
               <!-- Add Breadcrumb for New Wallet page -->
               <template v-if="route.path === '/settings/wallets/new'">
                 <BreadcrumbSeparator/>
                 <BreadcrumbItem>New</BreadcrumbItem>
               </template>
               <!-- Add logic for other settings sections -->
            </BreadcrumbList>
          </Breadcrumb>
          <!-- Add Spacer and Button -->
          <div class="ml-auto flex items-center gap-2">
            <NuxtLink v-if="route.path == '/settings/wallets'" to="/settings/wallets/new">
              <Button size="sm">Create Wallet</Button>
            </NuxtLink>
          </div>
        </header>
        <main class="flex flex-1 flex-col gap-4 p-4">
          <slot />
        </main>
      </SidebarInset>
    </SidebarProvider>
  </div>
</template>

<style scoped>
/* Add any specific styles if needed */
</style> 