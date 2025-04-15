<script setup lang="ts">
import type { IToken, IWallet } from '~/types'
import { ref, computed, watch } from 'vue'

interface Props {
  selectedWallet: IWallet
  onTokenActivation: (token: IToken) => void
}

const props = defineProps<Props>()
const limit = ref(100)
const offset = ref(0)
const chainType = computed(() => props.selectedWallet?.chainType)

const selectedToken = ref<string | null>(null)

const { tokens, isLoading, error, refresh } = useTokensList(chainType, limit, offset)

watch(chainType, () => {
  selectedToken.value = null
  refresh()
})

function handleActivateToken() {
  if (selectedToken.value) {
    props.onTokenActivation(tokens.value.find(token => token.address === selectedToken.value)!)
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
          <SelectValue placeholder="Choose a token" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem
            v-for="token in tokens"
            :key="token.address"
            :value="token.address"
          >
            <Web3Icon :symbol="token.symbol" variant="branded" class="size-4 mr-2 inline-block align-middle" />
            <span class="font-mono">{{ token.symbol }}</span>
          </SelectItem>
        </SelectContent>
      </Select>
      <div v-if="isLoading" class="text-xs text-muted mt-2">Loading tokens...</div>
      <div v-else-if="error" class="text-xs text-red-500 mt-2">Failed to load tokens. <button class="underline" @click="() => refresh()">Retry</button></div>
      <div v-else-if="tokens.length === 0" class="text-xs text-muted mt-2">No tokens found for this chain.</div>
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