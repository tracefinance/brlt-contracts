<script setup lang="ts">
import type { IToken, IWallet, ITokenBalanceResponse, ChainType, TokenType } from '~/types'
import { ref, computed, watch } from 'vue'
import { shortenAddress } from '~/lib/utils'

interface Props {
  selectedWallet: IWallet
  onTokenActivation: (token: IToken) => void
  balances: ITokenBalanceResponse[]
}

const props = defineProps<Props>()

const fetchLimit = ref(500)
const fetchNextToken = ref<string | undefined>(undefined)
const fetchChainTypeFilter = computed<ChainType | null>(() => {
    const type = props.selectedWallet?.chainType;
    if (type === 'ethereum' || type === 'polygon' || type === 'base') {
        return type as ChainType;
    }
    return null;
});
const fetchTokenTypeFilter = ref<TokenType | null>(null)

const selectedTokenAddress = ref<string | null>(null)

const { tokens, isLoading, error, refresh } = useTokensList(
  fetchChainTypeFilter,
  fetchTokenTypeFilter,
  fetchLimit,
  fetchNextToken
)

const availableTokens = computed(() => {
  if (!tokens.value || !props.balances) return []
  const activeTokenAddresses = new Set(props.balances.map(b => b.token.address.toLowerCase()))
  return tokens.value.filter(token => !activeTokenAddresses.has(token.address.toLowerCase()))
})

const selectedTokenSymbol = computed(() => {
  if (!selectedTokenAddress.value || !availableTokens.value) return null
  const token = availableTokens.value.find(t => t.address === selectedTokenAddress.value)
  return token?.symbol ?? null
})

watch(fetchChainTypeFilter, () => {
  selectedTokenAddress.value = null
  refresh()
})

function handleActivateToken() {
  if (selectedTokenAddress.value) {
    const tokenToActivate = tokens.value.find(token => token.address === selectedTokenAddress.value)
    if (tokenToActivate) {
      props.onTokenActivation(tokenToActivate)
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
    <PopoverContent class="w-60 flex flex-col gap-4 p-4">
      <div class="text-sm font-medium mb-2">Select Token to Activate</div>
      <Select v-model="selectedTokenAddress" :disabled="isLoading || !!error">
        <SelectTrigger>
          <div class="flex items-center gap-1 truncate">
            <Web3Icon
              v-if="selectedTokenSymbol"
              :symbol="selectedTokenSymbol"
              variant="branded"
              class="size-5 flex-shrink-0"/>
            <span v-if="selectedTokenSymbol" class="truncate">{{ selectedTokenSymbol }}</span>
            <SelectValue v-else placeholder="Choose a token" />
          </div>
        </SelectTrigger>
        <SelectContent>
          <div v-if="isLoading" class="p-2 text-xs text-muted-foreground">Loading...</div>
          <div v-else-if="error" class="p-2 text-xs text-red-500">Error loading tokens.</div>
          <div v-else-if="availableTokens.length === 0" class="p-2 text-xs text-muted-foreground">
            {{ tokens.length > 0 ? 'All tokens activated' : 'No tokens available' }}
          </div>
          <SelectItem
            v-for="token in availableTokens"
            v-else
            :key="token.address"
            :value="token.address"
            class="font-mono"
          >
            <Web3Icon :symbol="token.symbol" variant="branded" class="size-5 mr-2 inline-block align-middle" />
            <span>{{ token.symbol }}</span>
            <span class="text-xs text-muted-foreground ml-1">({{ shortenAddress(token.address, 4, 4) }})</span>
          </SelectItem>
        </SelectContent>
      </Select>

       <Button v-if="error && !isLoading" variant="outline" size="sm" class="mt-2" @click="refresh">
          Retry
       </Button>

      <Button
        variant="default"
        :disabled="!selectedTokenAddress || isLoading || !!error"
        class="mt-2"
        @click="handleActivateToken"
      >
        Activate
      </Button>
    </PopoverContent>
  </Popover>
</template>