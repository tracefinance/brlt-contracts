<script setup lang="ts">
import { computed } from 'vue'
import { ZERO_ADDRESS, formatCurrency } from '~/lib/utils'
import type { IWallet, ITokenBalanceResponse, IToken } from '~/types'

interface Props {
  wallets?: IWallet[]
  selectedWallet: IWallet
  balances?: ITokenBalanceResponse[]
  activeTokenAddress?: string
  onWalletChange: (wallet: IWallet) => void
  onTokenActivation: (token: IToken) => void  
}

const props = withDefaults(defineProps<Props>(), {
  wallets: () => [],
  balances: () => [],
  activeTokenAddress: undefined
})

const comparisonAddress = computed(() => props.activeTokenAddress?.toLowerCase())
</script>

<template>
  <Sidebar>
    <SidebarHeader class="mt-16">
      <WalletSelect
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
                :to="`/wallets/${selectedWallet.chainType}/${selectedWallet.address}/transactions/${(balance.token?.address || ZERO_ADDRESS).toLowerCase()}`"
              >
                <Web3Icon :symbol="balance.token?.symbol || 'N/A'" variant="branded" class="size-6"/>
                <span>{{ balance.token?.symbol || 'N/A' }}</span>
                <span class="ml-auto text-sm text-gray-500 font-mono">
                  {{ formatCurrency(balance.balance) }}
                </span>
              </NuxtLink>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarGroup>
    </SidebarContent>
    <SidebarFooter>
      <SidebarMenu>
        <WalletActivateTokenPopover 
          :selected-wallet="selectedWallet" 
          :on-token-activation="onTokenActivation" 
          :balances="balances" 
        />
      </SidebarMenu>
    </SidebarFooter>
  </Sidebar>
</template> 