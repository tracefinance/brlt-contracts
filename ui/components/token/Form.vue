<script setup lang="ts">
import { ref, watch } from 'vue'
import type { PropType } from 'vue'
import type { IAddTokenRequest, ChainType, TokenType, IToken } from '~/types'
import { Button } from '~/components/ui/button'
import { Input } from '~/components/ui/input'
import { Label } from '~/components/ui/label'
import { Select, SelectContent, SelectGroup, SelectItem, SelectTrigger, SelectValue } from '~/components/ui/select'

// Define props
const props = defineProps({
  initialData: {
    type: Object as PropType<IToken | null>,
    default: null,
  },
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

// Form state
const address = ref('')
const chainType = ref<ChainType>('ethereum') // Default value
const symbol = ref('')
const decimals = ref<number | undefined>(18) // Default value
const type = ref<TokenType>('erc20') // Default value

// TODO: Fetch these from a reference data source eventually
const chainTypeOptions: ChainType[] = ['ethereum', 'polygon', 'base']
const tokenTypeOptions: TokenType[] = ['erc20', 'erc721', 'erc1155', 'native']

// Populate form if initialData is provided (for editing)
watch(
  () => props.initialData,
  (newData) => {
    if (newData) {
      address.value = newData.address
      chainType.value = newData.chainType
      symbol.value = newData.symbol
      decimals.value = newData.decimals
      type.value = newData.type
    } else {
      // Reset form if initialData becomes null
      address.value = ''
      chainType.value = 'ethereum'
      symbol.value = ''
      decimals.value = 18
      type.value = 'erc20'
    }
  },
  { immediate: true },
)

function handleSubmit() {
  if (decimals.value === undefined) {
    // Handle error: decimals is required
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
    <div>
      <Label for="address">Token Address</Label>
      <Input id="address" v-model="address" required placeholder="0x..." />
      <!-- TODO: Add address validation pattern -->
    </div>

    <div class="grid grid-cols-2 gap-4">
      <div>
        <Label for="chainType">Chain Type</Label>
        <Select id="chainType" v-model="chainType" required>
          <SelectTrigger>
            <SelectValue placeholder="Select chain type" />
          </SelectTrigger>
          <SelectContent>
            <SelectGroup>
              <SelectItem v-for="option in chainTypeOptions" :key="option" :value="option">
                {{ option }}
              </SelectItem>
            </SelectGroup>
          </SelectContent>
        </Select>
      </div>

      <div>
        <Label for="type">Token Type</Label>
        <Select id="type" v-model="type" required>
          <SelectTrigger>
            <SelectValue placeholder="Select token type" />
          </SelectTrigger>
          <SelectContent>
            <SelectGroup>
              <SelectItem v-for="option in tokenTypeOptions" :key="option" :value="option">
                {{ option }}
              </SelectItem>
            </SelectGroup>
          </SelectContent>
        </Select>
      </div>
    </div>

    <div class="grid grid-cols-2 gap-4">
      <div>
        <Label for="symbol">Symbol</Label>
        <Input id="symbol" v-model="symbol" required placeholder="e.g., ETH, USDC" />
      </div>

      <div>
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
    </div>

    <div class="flex justify-end gap-2 pt-4">
      <Button type="button" variant="outline" @click="handleCancel" :disabled="isLoading">
        Cancel
      </Button>
      <Button type="submit" :disabled="isLoading">
        {{ isLoading ? 'Saving...' : 'Save Token' }}
      </Button>
    </div>
  </form>
</template> 