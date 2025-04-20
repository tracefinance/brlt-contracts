<script setup lang="ts">
import type { IToken, IWallet, ITokenBalanceResponse } from '~/types'
import { ref, computed, watch } from 'vue'

interface Props {
  selectedWallet: IWallet
  onTokenActivation: (token: IToken) => void
  balances: ITokenBalanceResponse[]
}

const props = defineProps<Props>()
const limit = ref(100)
const nextToken = ref<string | undefined>(undefined)
const chainType = computed(() => props.selectedWallet?.chainType)

const selectedToken = ref<string | null>(null)

const { tokens, isLoading, error, refresh } = useTokensList(chainType, limit, nextToken)

// Filter out tokens that are already active in the wallet
const availableTokens = computed(() => {
  if (!tokens.value || !props.balances) return []
  
  // Get the addresses of tokens that already have balances
  const activeTokenAddresses = new Set(props.balances.map(balance => balance.token.address.toLowerCase()))
  
  // Filter out tokens that are already active
  return tokens.value.filter(token => !activeTokenAddresses.has(token.address.toLowerCase()))
})

// Computed property to find the symbol of the selected token
const selectedTokenSymbol = computed(() => {
  if (!selectedToken.value || !availableTokens.value) return null
  const token = availableTokens.value.find(t => t.address === selectedToken.value)
  return token?.symbol ?? null
})

watch(chainType, () => {
  selectedToken.value = null
  refresh()
})

function handleActivateToken() {
  if (selectedToken.value) {
    const token = tokens.value.find(token => token.address === selectedToken.value)
    if (token) {
      props.onTokenActivation(token)
    }
  }
}
</script>

<template>
  <Popover>
    <PopoverTrigger as-child>
      <SidebarMenuItem>
        <SidebarMenuButton :is-active="false" class="flex items-center">
          <Icon name="lucide:circle-plus" class="size-4" />
          <span>Activate Token</span>
        </SidebarMenuButton>
      </SidebarMenuItem>
    </PopoverTrigger>
    <PopoverContent class="w-60 flex flex-col gap-4">
      <Select v-model="selectedToken" class="w-full">
        <SelectTrigger>
          <div class="flex items-center gap-1">
            <Web3Icon 
              v-if="selectedTokenSymbol" 
              :symbol="selectedTokenSymbol" 
              variant="branded" 
              class="size-5"/> 
            <SelectValue placeholder="Choose a token" />
          </div>
        </SelectTrigger>
        <SelectContent>
          <SelectItem
            v-for="token in availableTokens"
            :key="token.address"
            :value="token.address"
          >
            <Web3Icon :symbol="token.symbol" variant="branded" class="size-5 mr-2 inline-block align-middle" />
            <span class="font-mono">{{ token.symbol }}</span>
          </SelectItem>
        </SelectContent>
      </Select>
      <div v-if="isLoading" class="text-xs text-muted mt-2">Loading tokens...</div>
      <div v-else-if="error" class="text-xs text-red-500 mt-2">Failed to load tokens. <button class="underline" @click="() => refresh()">Retry</button></div>
      <div v-else-if="availableTokens.length === 0 && tokens.length > 0" class="text-xs text-muted mt-2">All available tokens are already activated.</div>
      <div v-else-if="availableTokens.length === 0" class="text-xs text-muted mt-2">No tokens found for this chain.</div>
      <Button
        variant="default"
        :disabled="!selectedToken"
        @click="handleActivateToken"
      >
        Activate Token
      </Button>
    </PopoverContent>
  </Popover>
</template>