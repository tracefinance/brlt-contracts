<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useRouter } from 'vue-router'
import { toast } from 'vue-sonner'
import type { IAddTokenRequest, ChainType, TokenType } from '~/types'
import { getErrorMessage } from '~/lib/utils'

// Settings
definePageMeta({
  layout: 'settings',
})

// State & Composables
const router = useRouter()
const { 
  addToken: mutateAddToken, 
  isCreating, 
  error: mutationError 
} = useTokenMutations()

// Form state moved from component
const address = ref('')
const chainType = ref<ChainType | null>(null)
const symbol = ref('')
const decimals = ref<number | undefined>(18)
const type = ref<TokenType>('erc20')

const tokenTypeOptions: TokenType[] = ['erc20', 'erc721', 'erc1155']

// Computed property moved from component
const chainTypeModel = computed({
  get: () => chainType.value ?? '', 
  set: (value) => {
    chainType.value = value || null
  }
})

// Watch for errors from the composable
watch(mutationError, (newError) => {
  if (newError) {
    toast.error('Failed to add token', {
      description: getErrorMessage(newError, 'An unexpected error occurred.'),
    })
  }
})

// Handle form submission - logic merged from component
async function handleAddToken() {
  mutationError.value = null
  
  // Validation using local refs
  if (!chainType.value) {
    toast.error('Chain Type is required.');
    return;
  }
  if (!address.value || !symbol.value || decimals.value === undefined || decimals.value === null || !type.value) {
    toast.error('All fields are required.');
    return;
  }

  // Construct payload from local refs
  const payload: IAddTokenRequest = {
    address: address.value.trim(),
    chainType: chainType.value,
    symbol: symbol.value.trim(),
    decimals: decimals.value,
    type: type.value,
  }
  
  const newToken = await mutateAddToken(payload)

  if (newToken) {
    toast.success(`Token ${newToken.symbol} added successfully!`)
    router.push('/settings/tokens') 
  }
}

// Use router directly, no need for emit
function handleCancel() {
  router.back()
}
</script>

<template>
  <div class="flex justify-center">
    <Card class="w-full max-w-2xl">
      <CardHeader>
        <CardTitle>Add New Token</CardTitle>
        <CardDescription>Register a new token in the system.</CardDescription>
      </CardHeader>
      <CardContent>
        <form class="space-y-4" @submit.prevent="handleAddToken">
          <div class="space-y-2">
            <Label for="address">Token Address</Label>
            <Input id="address" v-model="address" required placeholder="0x..." />
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
            <Button type="button" variant="outline" :disabled="isCreating" @click="handleCancel">
              Cancel
            </Button>
            <Button type="submit" :disabled="isCreating">
              {{ isCreating ? 'Adding...' : 'Add Token' }}
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  </div>
</template> 