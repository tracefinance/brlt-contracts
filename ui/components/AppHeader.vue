<script setup lang="ts">
import { Wallet, Vault, Repeat, Shuffle, Settings, MoonStar, Sun } from 'lucide-vue-next'
const colorMode = useColorMode()
</script>

<template>
  <header class="fixed top-0 left-0 right-0 z-50 w-full border-b bg-background h-16 flex items-center justify-between px-4">
      <div class="flex w-full h-full items-center gap-4">
        <Logo class="h-10"/>

        <nav class="flex items-center space-x-6 text-sm font-medium ml-4">
          <NuxtLink to="/wallets" class="flex items-center gap-2">
            <Wallet class="h-[1.2rem] w-[1.2rem]" stroke-width="1.5" />
            <span>Wallets</span>
          </NuxtLink>

          <NuxtLink to="/vaults" class="flex items-center gap-2">
            <Vault class="h-[1.2rem] w-[1.2rem]" stroke-width="1.5" />
            <span>Vaults</span>
          </NuxtLink>
          
          <NuxtLink to="/swaps" class="flex items-center gap-2">
            <Repeat class="h-[1.2rem] w-[1.2rem]" stroke-width="1.5" />
            <span>Swap</span>
          </NuxtLink>

          <NuxtLink to="/bridges" class="flex items-center gap-2">
            <Shuffle class="h-[1.2rem] w-[1.2rem]" stroke-width="1.5" />
            <span>Bridge</span>
          </NuxtLink>

          <NuxtLink to="/settings" class="flex items-center gap-2">
            <Settings class="h-[1.2rem] w-[1.2rem]" stroke-width="1.5" />
            <span>Settings</span>
          </NuxtLink>
        </nav>        
      </div>      
      <ThemeToggle />    
    </header>
</template>