<script setup lang="ts">
import { ZERO_ADDRESS } from '~/lib/utils'

definePageMeta({
  layout: 'wallet',
  middleware: async () => {
    const { $api } = useNuxtApp()
    
    const { data: walletsData } = await useAsyncData(
      'walletsRedirect',
      () => $api.wallet.listWallets(1)
    )
    
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