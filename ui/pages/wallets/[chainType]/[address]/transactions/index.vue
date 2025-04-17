<script setup lang="ts">
import { ZERO_ADDRESS } from '~/lib/utils'

definePageMeta({
  layout: 'wallet',
  middleware: (to) => {
    const chainType = to.params.chainType
    const address = to.params.address

    if (to.fullPath === `/wallets/${chainType}/${address}/transactions`) {
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