<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { toast } from 'vue-sonner'
import { getErrorMessage } from '~/lib/utils'
import type { ChainType, IUpdateTokenRequest, TokenType } from '~/types'

// Settings
definePageMeta({
  layout: 'settings',
})

// State & Composables
const route = useRoute()
const router = useRouter()

// Read chainType and address from route params
const addressParam = computed(() => route.params.address as string | undefined)

const { 
  currentToken: fetchedToken,
  isLoading: isLoadingToken, 
  error: loadingError
} = useTokenDetails(addressParam)

const { 
  updateToken: mutateUpdateToken, 
  isUpdating, 
  error: mutationError 
} = useTokenMutations()

// Form state - Initialize potentially from fetchedToken later
const formAddress = ref('')
const formChainType = ref<ChainType | null>(null)
const formSymbol = ref('')
const formDecimals = ref<number | undefined>(undefined)
const formType = ref<TokenType>('erc20')

const tokenTypeOptions: TokenType[] = ['erc20', 'erc721', 'erc1155']

// Populate form when token data loads
watch(fetchedToken, (newTokenData) => {
  if (newTokenData) {
    formAddress.value = newTokenData.address
    formChainType.value = newTokenData.chainType
    formSymbol.value = newTokenData.symbol
    formDecimals.value = newTokenData.decimals
    formType.value = newTokenData.type
  }
}, { immediate: true })

// Watch for loading errors
watch(loadingError, (newError) => {
  if (newError) {
    toast.error('Failed to load token data', {
      description: getErrorMessage(newError, 'Could not fetch token details.'),
    })
  }
})

// Watch for mutation errors
watch(mutationError, (newError) => {
  if (newError) {
    toast.error('Failed to update token', {
      description: getErrorMessage(newError, 'An unexpected error occurred during the update.'),
    })
  }
})

// Handle form submission
async function handleUpdateToken() {
  mutationError.value = null

  // Ensure address is not null/undefined before calling
  const currentAddress = addressParam.value;
  
  if (!currentAddress) {
    toast.error('Address is missing from the route parameters.');
    return;
  }
  
  // Validation
  if (!formSymbol.value || formDecimals.value === undefined || formDecimals.value === null || !formType.value) {
    toast.error('Symbol, Decimals, and Type fields are required.');
    return;
  }

  // Construct payload
  const payload: IUpdateTokenRequest = {
    symbol: formSymbol.value.trim(),
    decimals: formDecimals.value,
    type: formType.value,
  }
  
  const updatedToken = await mutateUpdateToken(currentAddress, payload)

  if (updatedToken) {
    toast.success(`Token ${updatedToken.symbol} updated successfully!`)
    router.push('/settings/tokens') 
  }
}

// Handle Cancel
function handleCancel() {
  router.back()
}

</script>

<template>
  <div class="flex justify-center">
    <Card class="w-full max-w-2xl">
      <CardHeader>
        <CardTitle>Edit Token</CardTitle>
        <CardDescription>Update the details for the selected token.</CardDescription>
      </CardHeader>
      <CardContent>
        <div v-if="isLoadingToken && !fetchedToken" class="flex justify-center items-center p-8">
           <Icon name="svg-spinners:3-dots-fade" class="h-8 w-8" />
        </div>
        <div v-else-if="loadingError && !fetchedToken" class="text-destructive text-center p-8">
          Failed to load token data. Please try again or go back.
        </div>
        <form v-else class="space-y-4" @submit.prevent="handleUpdateToken">
          
           <div class="space-y-2">
            <Label for="address">Token Address (Read-only)</Label>
            <Input id="address" :model-value="formAddress" readonly disabled />
          </div>
          
           <div class="space-y-2">
            <Label for="chainType">Chain Type (Read-only)</Label>
            <div class="flex items-center gap-2 px-3 py-2 text-sm border rounded-md text-muted-foreground">
              <Web3Icon v-if="formChainType" :symbol="formChainType" variant="branded" class="size-5" />
              <span class="capitalize">{{ formChainType || 'N/A' }}</span>
            </div>
          </div>

          <div class="space-y-2">
            <Label for="type">Token Type</Label>
            <Select id="type" v-model="formType" required>
              <SelectTrigger>
                <SelectValue as-child>
                  <span class="uppercase">{{ formType || 'Select token type' }}</span>
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
            <Input id="symbol" v-model="formSymbol" required placeholder="e.g., ETH, USDC" />
          </div>

          <div class="space-y-2">
            <Label for="decimals">Decimals</Label>
            <Input
              id="decimals"
              v-model.number="formDecimals"
              required
              type="number"
              min="0"
              max="18" 
              placeholder="e.g., 18"
            />
          </div>

          <div class="flex justify-end gap-2 pt-4">
            <Button type="button" variant="outline" :disabled="isUpdating || isLoadingToken" @click="handleCancel">
              Cancel
            </Button>
            <Button type="submit" :disabled="isUpdating || isLoadingToken">
              <span v-if="isUpdating">Saving...</span>
              <span v-else>Save Changes</span>
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  </div>
</template> 