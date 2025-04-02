<script setup lang="ts">
import { ZERO_ADDRESS } from '~/lib/utils'

// Define page metadata with server-side middleware for redirection
definePageMeta({
  layout: 'wallet',
  middleware: async (to) => {
    // Get API client
    const { $api } = useNuxtApp()
    
    // Fetch wallets data on the server
    const { data: walletsData } = await useAsyncData(
      'wallets',
      () => $api.wallet.listWallets(10, 0)
    )
    
    // If there are wallets, redirect to the first one's transactions
    if (walletsData.value && walletsData.value.items && walletsData.value.items.length > 0) {
      const wallet = walletsData.value.items[0]
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