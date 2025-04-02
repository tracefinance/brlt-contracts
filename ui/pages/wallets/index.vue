<script setup lang="ts">
// Define page metadata
definePageMeta({
  layout: 'wallet'
})
</script>

<template>
  <div class="container p-8">
    <h1 class="text-3xl font-bold mb-6">Wallet Dashboard</h1>
    <p class="text-lg mb-4">
      Welcome to your wallet dashboard. Select a wallet from the sidebar to view your tokens and transactions.
    </p>
    
    <div class="grid gap-6 mt-8">
      <div class="bg-card rounded-lg shadow-sm p-6">
        <h2 class="text-xl font-semibold mb-4">Getting Started</h2>
        <ul class="list-disc pl-5 space-y-2">
          <li>Select a wallet from the dropdown in the sidebar</li>
          <li>View your token balances in the sidebar</li>
          <li>Click on a token to view its transaction history</li>
        </ul>
      </div>
    </div>
  </div>
</template> 