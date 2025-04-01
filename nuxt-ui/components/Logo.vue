<script setup lang="ts">
import { ref, onMounted, watch, onUnmounted, computed } from 'vue'

// Import SVGs directly
import logoBlack from '~/assets/img/logo-black.svg'
import logoWhite from '~/assets/img/logo-white.svg'

interface Props {
  width?: number
  height?: number
  class?: string
}

const props = withDefaults(defineProps<Props>(), {
  width: 150,
  height: 50,
  class: 'h-14 w-auto'
})

const colorMode = useColorMode()
const logoSrc = ref(logoBlack)
const mounted = ref(false)

// Prevent hydration mismatch with computed prop
const currentLogo = computed(() => {
  if (!process.client || !mounted.value) {
    // Return null during SSR to delay rendering until client-side
    return null
  }
  
  return logoSrc.value
})

// Set mounted state
onMounted(() => {
  mounted.value = true
  updateLogoSrc()
  
  // For system preference, also set up the event listener
  if (colorMode.preference === 'system') {
    mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
    handleChange = (e: MediaQueryListEvent) => {
      logoSrc.value = e.matches ? logoWhite : logoBlack
    }
    
    mediaQuery.addEventListener('change', handleChange)
  }
})

// Update logo based on theme
function updateLogoSrc() {
  // For explicit light/dark themes
  if (colorMode.preference === 'light') {
    logoSrc.value = logoBlack
    return
  }
  
  if (colorMode.preference === 'dark') {
    logoSrc.value = logoWhite
    return
  }
  
  // For system theme, check OS preference
  if (colorMode.preference === 'system' && process.client) {
    const isDarkMode = window.matchMedia('(prefers-color-scheme: dark)').matches
    logoSrc.value = isDarkMode ? logoWhite : logoBlack
  }
}

// Watch for theme changes, but only after mount
watch(() => colorMode.preference, updateLogoSrc)

// Handle system preference changes
let mediaQuery: MediaQueryList | null = null
let handleChange: ((e: MediaQueryListEvent) => void) | null = null

onUnmounted(() => {
  if (process.client && mediaQuery && handleChange) {
    mediaQuery.removeEventListener('change', handleChange)
  }
})
</script>

<template>
  <ClientOnly>
    <img 
      v-if="currentLogo"
      :src="currentLogo" 
      alt="Vault0 Logo" 
      :width="width" 
      :height="height" 
      :class="class"
    />
    
    <template #fallback>
      <img 
        :src="logoBlack" 
        alt="Vault0 Logo" 
        :width="width" 
        :height="height" 
        :class="class"
      />
    </template>
  </ClientOnly>
</template>
