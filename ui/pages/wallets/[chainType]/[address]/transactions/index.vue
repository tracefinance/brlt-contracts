<script setup lang="ts">
import { ZERO_ADDRESS } from '~/lib/utils'

// Define page metadata with server-side middleware
definePageMeta({
  layout: 'wallet',
  middleware: (to) => {
    // Get route params
    const chainType = to.params.chainType
    const address = to.params.address
    
    // Check if we're at exactly the transactions path (no token specified)
    if (to.fullPath === `/wallets/${chainType}/${address}/transactions`) {
      // Redirect to the native token transactions page
      return navigateTo(`/wallets/${chainType}/${address}/transactions/${ZERO_ADDRESS}`, {
        redirectCode: 302
      })
    }
  }
})
</script>

<template>
  <div>
    <h1>Redirecting to native token transactions...</h1>
  </div>
</template> 