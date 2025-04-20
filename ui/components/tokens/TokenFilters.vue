<script setup lang="ts">
import { ref, watch } from 'vue'
import type { Ref } from 'vue'
import type { ChainType, TokenType } from '~/types'
import { Select, SelectContent, SelectGroup, SelectItem, SelectLabel, SelectTrigger, SelectValue } from '~/components/ui/select'
import { Button } from '~/components/ui/button'

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

// Local refs to manage select values, syncing with props
// Use empty string locally to represent the "All" (null) state for the select component
const selectedChainType: Ref<ChainType | '' > = ref(props.chainTypeFilter ?? '')
const selectedTokenType: Ref<TokenType | '' > = ref(props.tokenTypeFilter ?? '')

// Watch props to update local state if changed externally (e.g., URL change)
watch(() => props.chainTypeFilter, (newValue) => {
  selectedChainType.value = newValue ?? ''
})
watch(() => props.tokenTypeFilter, (newValue) => {
  selectedTokenType.value = newValue ?? ''
})

// Emit updates when local refs change
watch(selectedChainType, (newValue) => {
  // Emit null if the empty string (placeholder) is selected, otherwise emit the value
  emit('update:chainTypeFilter', newValue === '' ? null : newValue)
})
watch(selectedTokenType, (newValue) => {
  // Emit null if the empty string (placeholder) is selected, otherwise emit the value
  emit('update:tokenTypeFilter', newValue === '' ? null : newValue)
})

// TODO: Fetch these from a reference data source eventually
const chainTypeOptions: ChainType[] = ['ethereum', 'polygon', 'base']
const tokenTypeOptions: TokenType[] = ['erc20', 'erc721', 'erc1155', 'native']

function handleClearFilters() {
  emit('clearFilters')
}
</script>

<template>
  <div class="flex items-center gap-4 mb-4">
    <!-- Use v-model with the local ref (which uses '' for null) -->
    <Select v-model="selectedChainType">
      <SelectTrigger class="w-[180px]">
        <SelectValue placeholder="Filter by Chain..." />
      </SelectTrigger>
      <SelectContent>
        <SelectGroup>
          <SelectLabel>Chain Type</SelectLabel>
          <!-- The empty string value corresponds to the null state -->
          <SelectItem value="">All Chains</SelectItem>
          <SelectItem v-for="option in chainTypeOptions" :key="option" :value="option">
            {{ option }}
          </SelectItem>
        </SelectGroup>
      </SelectContent>
    </Select>

    <!-- Use v-model with the local ref (which uses '' for null) -->
    <Select v-model="selectedTokenType">
      <SelectTrigger class="w-[180px]">
        <SelectValue placeholder="Filter by Type..." />
      </SelectTrigger>
      <SelectContent>
        <SelectGroup>
          <SelectLabel>Token Type</SelectLabel>
          <!-- The empty string value corresponds to the null state -->
          <SelectItem value="">All Types</SelectItem>
          <SelectItem v-for="option in tokenTypeOptions" :key="option" :value="option">
            {{ option }}
          </SelectItem>
        </SelectGroup>
      </SelectContent>
    </Select>

    <Button variant="outline" @click="handleClearFilters">Clear Filters</Button>
  </div>
</template> 