<script setup>
// Redirect to the wallets page
definePageMeta({
  middleware: () => {
    return navigateTo('/wallets')
  }
})
</script>

<template>
  <div>
    <!-- This content won't be shown as the page will redirect -->
  </div>
</template>