<script setup lang="ts">
import { computed } from 'vue'
import { ZERO_ADDRESS, formatCurrency } from '~/lib/utils'
import type { Wallet, TokenBalanceResponse } from '~/types/wallet'

// Define props
interface Props {
  wallets: Wallet[]
  selectedWallet: Wallet
  onWalletChange: (wallet: Wallet) => void
  balances: TokenBalanceResponse[]
  activeTokenAddress?: string
}

// Default props
const props = withDefaults(defineProps<Props>(), {
  wallets: () => [],
  balances: () => [],
  activeTokenAddress: undefined
})

// Computed address for comparison
const comparisonAddress = computed(() => props.activeTokenAddress?.toLowerCase())
</script>

<template>
  <Sidebar class="mt-16">
    <SidebarHeader>
      <WalletSelector 
        :wallets="wallets" 
        :selected-wallet="selectedWallet" 
        :on-wallet-change="onWalletChange"
      />
    </SidebarHeader>
    <SidebarContent>
      <SidebarGroup>
        <SidebarGroupLabel>Tokens</SidebarGroupLabel>
        <SidebarMenu>
          <SidebarMenuItem v-for="balance in balances" :key="(balance.token?.address || ZERO_ADDRESS).toLowerCase()">
            <SidebarMenuButton 
              :is-active="comparisonAddress === (balance.token?.address || ZERO_ADDRESS).toLowerCase()"
              as-child
            >
              <NuxtLink 
                class="flex items-center w-full" 
                :to="`/wallets/${selectedWallet.address}/${selectedWallet.chainType}/transactions/${(balance.token?.address || ZERO_ADDRESS).toLowerCase()}`"
              >
                <Web3Icon :symbol="balance.token?.symbol || 'N/A'" variant="branded" class="size-6"/>
                <span>{{ balance.token?.symbol || 'N/A' }}</span>
                <span class="ml-auto text-sm text-gray-500">
                  {{ formatCurrency(balance.balance) }}
                </span>
              </NuxtLink>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarGroup>
    </SidebarContent>
  </Sidebar>
</template> 