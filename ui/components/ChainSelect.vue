<script setup lang="ts">
import { computed, watch } from 'vue'

// Define props and emits for v-model
const props = defineProps({
  modelValue: {
    type: String,
    required: true
  }
})

const emit = defineEmits(['update:modelValue'])

// Use the composable to load chains
const { chains, isLoading: isLoadingChains, error: chainsError, refresh: refreshChains } = useChains()

// Computed chain options for the Select component
const chainOptions = computed(() => {
  return chains.value.map(chain => ({
    label: chain.name,
    value: chain.type,
    type: chain.type
  }))
})

// Set a default chain type when chains load, only if the modelValue is currently empty
watch(chains, (newChains) => {
  if (props.modelValue === '' && newChains && newChains.length > 0) {
    // Emit update to set the default value in the parent
    emit('update:modelValue', newChains[0].type)
  }
}, { immediate: true })

// Computed property to handle the v-model binding with the Select component
const selectedChain = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

</script>

<template>
  <div>
    <Skeleton v-if="isLoadingChains" class="h-9 w-full rounded-md" />
    <div v-else-if="chainsError" class="text-red-500 text-sm">
      <span>Error loading chains: {{ chainsError.message }}.</span>
      <Button variant="link" size="sm" class="p-0 h-auto ml-1" @click="refreshChains">Retry</Button>
    </div>
    <Select v-else v-model="selectedChain">
      <SelectTrigger class="bg-background">
        <div class="flex items-center">
          <Web3Icon v-if="selectedChain" :symbol="selectedChain" variant="branded" class="size-5 mr-2"/> 
          <SelectValue placeholder="Select blockchain" />
        </div>
      </SelectTrigger>
      <SelectContent>
        <SelectItem 
          v-for="option in chainOptions" 
          :key="option.value" 
          :value="option.value"
        >
          <Web3Icon :symbol="option.type" variant="branded" class="size-5 mr-2 inline-block align-middle"/> 
          <span class="capitalize font-mono">{{ option.label }}</span>
        </SelectItem>
      </SelectContent>
    </Select>
  </div>
</template> 