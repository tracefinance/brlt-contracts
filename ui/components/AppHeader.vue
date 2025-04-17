<script setup lang="ts">
const route = useRoute()

// Function to check if a route is active
const isActive = (path: string) => {
  return route.path.startsWith(path)
}

const navLinks = [
  { path: '/wallets', icon: 'lucide:wallet', label: 'Wallets' },
  { path: '/vaults', icon: 'lucide:vault', label: 'Vaults' },
  { path: '/swaps', icon: 'lucide:repeat', label: 'Swap' },
  { path: '/bridges', icon: 'lucide:shuffle', label: 'Bridge' },
  { path: '/settings', icon: 'lucide:settings', label: 'Settings' },
];
</script>

<template>
  <header class="fixed top-0 left-0 right-0 z-50 w-full border-b bg-background h-16 flex items-center justify-between px-4">
      <div class="flex w-full h-full items-center gap-4">
        <Logo class="h-10"/>

        <nav class="flex items-center text-sm font-medium ml-4">
          <NuxtLink 
            v-for="link in navLinks" 
            :key="link.path"
            :to="link.path" 
            :class="{
              'flex items-center gap-2 px-2 py-2': true,
              'text-primary font-semibold bg-muted rounded-md': isActive(link.path),
              'text-muted-foreground hover:text-foreground': !isActive(link.path)
            }">
            <Icon :name="link.icon" class="size-5 flex-shrink-0" />
            <span class="w-16">{{ link.label }}</span>
          </NuxtLink>
        </nav>        
      </div>      
      <ThemeToggle />    
    </header>
</template>