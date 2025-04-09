<script setup lang="ts">
import { Wallet, Vault, Repeat, Shuffle, Settings, MoonStar, Sun } from 'lucide-vue-next'
const colorMode = useColorMode()
const route = useRoute()

// Function to check if a route is active
const isActive = (path: string) => {
  return route.path.startsWith(path)
}
</script>

<template>
  <header class="fixed top-0 left-0 right-0 z-50 w-full border-b bg-background h-16 flex items-center justify-between px-4">
      <div class="flex w-full h-full items-center gap-4">
        <Logo class="h-10"/>

        <nav class="flex items-center space-x-6 text-sm font-medium ml-4">
          <NuxtLink to="/wallets" :class="{
            'flex items-center gap-2': true,
            'text-primary font-semibold': isActive('/wallets'),
            'text-muted-foreground hover:text-foreground': !isActive('/wallets')
          }">
            <Wallet class="h-[1.2rem] w-[1.2rem]" stroke-width="1.5" />
            <span>Wallets</span>
          </NuxtLink>

          <NuxtLink to="/" :class="{
            'flex items-center gap-2': true,
            'text-primary font-semibold': isActive('/vaults'),
            'text-muted-foreground hover:text-foreground': !isActive('/vaults')
          }">
            <Vault class="h-[1.2rem] w-[1.2rem]" stroke-width="1.5" />
            <span>Vaults</span>
          </NuxtLink>
          
          <NuxtLink to="/" :class="{
            'flex items-center gap-2': true,
            'text-primary font-semibold': isActive('/swaps'),
            'text-muted-foreground hover:text-foreground': !isActive('/swaps')
          }">
            <Repeat class="h-[1.2rem] w-[1.2rem]" stroke-width="1.5" />
            <span>Swap</span>
          </NuxtLink>

          <NuxtLink to="/" :class="{
            'flex items-center gap-2': true,
            'text-primary font-semibold': isActive('/bridges'),
            'text-muted-foreground hover:text-foreground': !isActive('/bridges')
          }">
            <Shuffle class="h-[1.2rem] w-[1.2rem]" stroke-width="1.5" />
            <span>Bridge</span>
          </NuxtLink>

          <NuxtLink to="/settings" :class="{
            'flex items-center gap-2': true,
            'text-primary font-semibold': isActive('/settings'),
            'text-muted-foreground hover:text-foreground': !isActive('/settings')
          }">
            <Settings class="h-[1.2rem] w-[1.2rem]" stroke-width="1.5" />
            <span>Settings</span>
          </NuxtLink>
        </nav>        
      </div>      
      <ThemeToggle />    
    </header>
</template>