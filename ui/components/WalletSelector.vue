<script setup lang="ts">
import { computed } from 'vue'
import type { Wallet } from '~/types/wallet'
import { useSidebar } from "@/components/ui/sidebar";

interface Props {
  selectedWallet?: Wallet
  wallets?: Wallet[]
  onWalletChange?: (wallet: Wallet) => void
}

const props = withDefaults(defineProps<Props>(), {
  selectedWallet: undefined,
  wallets: () => [],
  onWalletChange: () => {}
})

// Get sidebar state
const sidebar = useSidebar()
const isMobile = computed(() => sidebar.isMobile.value)

// Computed properties for display
const displayChainType = computed(() => props.selectedWallet?.chainType || 'ethereum')
const displayName = computed(() => props.selectedWallet?.name || 'Select Wallet')

// Handle wallet selection
const handleWalletSelect = (wallet: Wallet) => {
  props.onWalletChange(wallet)
}
</script>

<template>
  <SidebarMenu>
    <SidebarMenuItem>
      <DropdownMenu>
        <DropdownMenuTrigger as-child>
          <div>
            <SidebarMenuButton
              size="lg"
              :class="{ 'bg-sidebar-accent text-sidebar-accent-foreground': $attrs['data-state'] === 'open' }"
            >
              <div class="flex aspect-square size-8 items-center justify-center rounded-md bg-sidebar-primary text-sidebar-primary-foreground">
                <Web3Icon :symbol="displayChainType" size="24" variant="mono" />
              </div>
              <div class="grid flex-1 text-left text-sm leading-tight">
                <span class="truncate font-semibold">
                  {{ displayName }}
                </span>
              </div>
              <Icon name="lucide:chevrons-up-down" class="ml-auto" />
            </SidebarMenuButton>
          </div>
        </DropdownMenuTrigger>
        <DropdownMenuContent
          class="w-[--radix-dropdown-menu-trigger-width] min-w-[200px]"
          :align="isMobile ? 'start' : 'start'"
          :side="isMobile ? 'bottom' : 'right'"
          :side-offset="4"
        >
          <template v-if="wallets && wallets.length > 0">
            <DropdownMenuItem
              v-for="wallet in wallets"
              :key="wallet.address"
              @click="handleWalletSelect(wallet)"
            >
              <div class="flex size-6 items-center justify-center rounded-md border bg-background">
                <Web3Icon :symbol="wallet.chainType" size="24" variant="mono" />
              </div>
              <span class="ml-2">{{ wallet.name }}</span>
            </DropdownMenuItem>
          </template>
          <div v-else class="p-3 text-sm text-muted-foreground">
            No wallets found
          </div>
        </DropdownMenuContent>
      </DropdownMenu>
    </SidebarMenuItem>
  </SidebarMenu>
</template> 