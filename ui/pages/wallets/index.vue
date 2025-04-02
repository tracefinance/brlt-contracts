<script setup lang="ts">
import { ZERO_ADDRESS } from '~/lib/utils'

// Define page metadata with server-side middleware for redirection
definePageMeta({
  layout: 'wallet',
  middleware: async (to) => {
    // Fetch wallets data on the server
    const { wallets, loadWallets } = useWallets()
    await loadWallets()
    
    // If there are wallets, redirect to the first one's transactions
    if (wallets.value.length > 0) {
      const wallet = wallets.value[0]
      return navigateTo(`/wallets/${wallet.chainType}/${wallet.address}/transactions/${ZERO_ADDRESS}`, { 
        redirectCode: 302 
      })
    }
  }
})
</script>

<template>
  <div>
    <h1>Redirecting to first wallet transactions...</h1>
  </div>
</template> 