<script setup lang="ts">
import { ref, computed } from 'vue'
import type { IAddTokenRequest, ChainType, TokenType } from '~/types'
// Imports from ~/components/ui/ and ~/components/ will be removed
// as Nuxt auto-imports them.

// Define props
defineProps({
  isLoading: {
    type: Boolean,
    default: false,
  },
})

// Define emits
const emit = defineEmits<{
  (e: 'submit', data: IAddTokenRequest): void
  (e: 'cancel'): void
}>()

// Form state - Defaults are suitable for creation
const address = ref('')
const chainType = ref<ChainType | null>(null)
const symbol = ref('')
const decimals = ref<number | undefined>(18)
const type = ref<TokenType>('erc20')

const tokenTypeOptions: TokenType[] = ['erc20', 'erc721', 'erc1155', 'native']

// Computed property for v-model binding with ChainSelect
const chainTypeModel = computed({
  get: () => chainType.value ?? '', 
  set: (value) => {
    chainType.value = value || null
  }
})

function handleSubmit() {
  if (!chainType.value) {
    console.error('Chain Type is required')
    return
  }
  if (decimals.value === undefined) {
    console.error('Decimals field is required')
    return
  }
  const formData: IAddTokenRequest = {
    address: address.value.trim(),
    chainType: chainType.value,
    symbol: symbol.value.trim(),
    decimals: decimals.value,
    type: type.value,
  }
  emit('submit', formData)
}

function handleCancel() {
  emit('cancel')
}
</script>

<template>
  <form class="space-y-4" @submit.prevent="handleSubmit">
    <div class="space-y-2">
      <Label for="address">Token Address</Label>
      <Input id="address" v-model="address" required placeholder="0x..." />
      <!-- TODO: Add address validation pattern -->
    </div>

    <div class="space-y-2">
      <Label for="chainType">Chain Type</Label>
      <ChainSelect id="chainType" v-model="chainTypeModel" required />
    </div>

    <div class="space-y-2">
      <Label for="type">Token Type</Label>
      <Select id="type" v-model="type" required>
        <SelectTrigger>
          <SelectValue as-child>
            <span class="uppercase">{{ type || 'Select token type' }}</span>
          </SelectValue>
        </SelectTrigger>
        <SelectContent>
          <SelectGroup>
            <SelectItem v-for="option in tokenTypeOptions" :key="option" :value="option">
              <span class="uppercase">{{ option }}</span>
            </SelectItem>
          </SelectGroup>
        </SelectContent>
      </Select>
    </div>

    <div class="space-y-2">
      <Label for="symbol">Symbol</Label>
      <Input id="symbol" v-model="symbol" required placeholder="e.g., ETH, USDC" />
    </div>

    <div class="space-y-2">
      <Label for="decimals">Decimals</Label>
      <Input
        id="decimals"
        v-model.number="decimals"
        required
        type="number"
        min="0"
        max="18" 
        placeholder="e.g., 18"
      />
    </div>

    <div class="flex justify-end gap-2 pt-4">
      <Button type="button" variant="outline" :disabled="isLoading" @click="handleCancel">
        Cancel
      </Button>
      <Button type="submit" :disabled="isLoading">
        {{ isLoading ? 'Adding...' : 'Add Token' }}
      </Button>
    </div>
  </form>
</template> 