<script setup lang="ts">
import { ref, watch } from 'vue'
import type { Ref } from 'vue'
import type { ChainType, TokenType } from '~/types'

// Define props using defineProps generic
const props = defineProps<{
  chainTypeFilter: ChainType | null
  tokenTypeFilter: TokenType | null
}>()

// Define emits
const emit = defineEmits<{
  (e: 'update:chainTypeFilter', value: ChainType | null): void
  (e: 'update:tokenTypeFilter', value: TokenType | null): void
  (e: 'clearFilters'): void
}>()

// Use a unique placeholder value locally instead of empty string for the "All" state
const ALL_FILTER_VALUE = '__ALL__'
const selectedChainType: Ref<ChainType | typeof ALL_FILTER_VALUE > = ref(props.chainTypeFilter ?? ALL_FILTER_VALUE)
const selectedTokenType: Ref<TokenType | typeof ALL_FILTER_VALUE > = ref(props.tokenTypeFilter ?? ALL_FILTER_VALUE)

// Watch props to update local state if changed externally (e.g., URL change)
watch(() => props.chainTypeFilter, (newValue) => {
  selectedChainType.value = newValue ?? ALL_FILTER_VALUE
})
watch(() => props.tokenTypeFilter, (newValue) => {
  selectedTokenType.value = newValue ?? ALL_FILTER_VALUE
})

// Emit updates when local refs change
watch(selectedChainType, (newValue) => {
  // Emit null if the placeholder value is selected, otherwise emit the actual value
  emit('update:chainTypeFilter', newValue === ALL_FILTER_VALUE ? null : newValue)
})
watch(selectedTokenType, (newValue) => {
  // Emit null if the placeholder value is selected, otherwise emit the actual value
  emit('update:tokenTypeFilter', newValue === ALL_FILTER_VALUE ? null : newValue)
})

// TODO: Fetch these from a reference data source eventually
const chainTypeOptions: ChainType[] = ['ethereum', 'polygon', 'base']
const tokenTypeOptions: TokenType[] = ['erc20', 'erc721', 'erc1155']

function handleClearFilters() {
  emit('clearFilters')
}
</script>

<template>
  <div class="flex items-center gap-4">
    <!-- Use v-model with the local ref (which uses ALL_FILTER_VALUE for null) -->
    <Select v-model="selectedChainType">
      <SelectTrigger class="w-[180px]">
        <SelectValue as-child>
          <div class="flex items-center gap-2">
            <template v-if="selectedChainType !== ALL_FILTER_VALUE">
              <Web3Icon :symbol="selectedChainType" class="size-4 flex-shrink-0" />
              <span class="capitalize">{{ selectedChainType }}</span>
            </template>
            <template v-else>
              <span class="text-muted-foreground">All Chains</span>
            </template>
          </div>
        </SelectValue>
      </SelectTrigger>
      <SelectContent>          
          <SelectItem :value="ALL_FILTER_VALUE">All Chains</SelectItem>
          <SelectItem v-for="option in chainTypeOptions" :key="option" :value="option">
            <div class="flex items-center gap-2">
              <Web3Icon :symbol="option" class="size-4" />
              <span class="capitalize">{{ option }}</span>
            </div>
          </SelectItem>
      </SelectContent>
    </Select>

    <!-- Use v-model with the local ref (which uses ALL_FILTER_VALUE for null) -->
    <Select v-model="selectedTokenType">
      <SelectTrigger class="w-[180px]">
        <SelectValue as-child>
          <div class="flex items-center gap-2">
            <template v-if="selectedTokenType !== ALL_FILTER_VALUE">
              <span class="uppercase">{{ selectedTokenType }}</span>
            </template>
            <template v-else>
              <span class="text-muted-foreground">All types</span>
            </template>
          </div>
        </SelectValue>
      </SelectTrigger>
      <SelectContent>
          <SelectItem :value="ALL_FILTER_VALUE">All Types</SelectItem>
          <SelectItem v-for="option in tokenTypeOptions" :key="option" :value="option" class="uppercase">
            {{ option }}
          </SelectItem>
      </SelectContent>
    </Select>

    <Button variant="outline" @click="handleClearFilters">Clear Filters</Button>
  </div>
</template> 