<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  symbol: string
  size?: string
  class?: string
  variant?: 'mono' | 'branded'
}

const props = withDefaults(defineProps<Props>(), {
  size: '1em',
  class: '',
  variant: 'mono'
})

// Compute the icon name by normalizing symbol: convert to lowercase, remove 'w' prefix for wrapped tokens, and add token: prefix
const iconName = computed(() => {
  let normalizedSymbol = props.symbol.toLowerCase()
  
  // Remove 'w' prefix for wrapped tokens (e.g., wbtc → btc, weth → eth)
  if (normalizedSymbol.startsWith('w') && normalizedSymbol.length > 1) {
    normalizedSymbol = normalizedSymbol.substring(1)
  }
  
  // Use different prefix based on variant
  const prefix = props.variant === 'branded' ? 'token-branded:' : 'token:'
  
  return `${prefix}${normalizedSymbol}`
})
</script>

<template>
  <Icon :name="iconName" :size="props.size" :class="props.class" />
</template>