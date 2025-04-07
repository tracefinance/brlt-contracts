<script setup lang="ts">
// Redirect immediately to the default settings page
await navigateTo('/settings/wallets', { replace: true, redirectCode: 301 })
</script>

<template>
  <div>Redirecting to wallet settings...</div>
</template> 