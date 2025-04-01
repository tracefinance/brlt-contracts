<script setup lang="ts">
import { Sun, MoonStar, MonitorCog } from 'lucide-vue-next'
import { useColorMode } from '#imports'

const colorMode = useColorMode()
</script>

<template>
  <DropdownMenu>
    <DropdownMenuTrigger as-child>
      <Button variant="ghost" size="icon">
        <Sun class="h-[1.2rem] w-[1.2rem] rotate-0 scale-100 transition-all dark:-rotate-90 dark:scale-0" />
        <MoonStar class="absolute h-[1.2rem] w-[1.2rem] rotate-90 scale-0 transition-all dark:rotate-0 dark:scale-100" />
        <span class="sr-only">Toggle theme</span>
      </Button>
    </DropdownMenuTrigger>
    <DropdownMenuContent align="end">
      <DropdownMenuItem @click="colorMode.preference = 'light'">
        <Sun class="mr-2 h-[1.2rem] w-[1.2rem]" />
        <span>Light</span>
      </DropdownMenuItem>
      <DropdownMenuItem @click="colorMode.preference = 'dark'">
        <MoonStar class="mr-2 h-[1.2rem] w-[1.2rem]" />
        <span>Dark</span>
      </DropdownMenuItem>
      <DropdownMenuItem @click="colorMode.preference = 'system'">
        <MonitorCog class="mr-2 h-[1.2rem] w-[1.2rem]" />
        <span>System</span>
      </DropdownMenuItem>
    </DropdownMenuContent>
  </DropdownMenu>
</template> 