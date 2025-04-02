<script setup lang="ts">
import { ref, watch } from 'vue';

interface PageSizeSelectProps {
  currentLimit: number;
  options?: number[];
}

const props = withDefaults(defineProps<PageSizeSelectProps>(), {
  options: () => [10, 20, 50, 100]
});

// Define the event that this component emits
const emit = defineEmits<{ 
  (e: 'update:limit', value: number): void 
}>();

// Local state for v-model, initialized with the current prop value
const selectedLimit = ref(String(props.currentLimit));

// Watch for changes in the prop to update local state
watch(() => props.currentLimit, (newLimit) => {
  selectedLimit.value = String(newLimit);
});

// Watch for changes in the local state (v-model) to emit the update event
watch(selectedLimit, (newLimitString) => {
  if (typeof newLimitString === 'string') {
    const newLimit = parseInt(newLimitString, 10);
    if (!isNaN(newLimit) && newLimit !== props.currentLimit) {
      emit('update:limit', newLimit);
    }
  }
});

</script>

<template>  
  <Select v-model="selectedLimit"> 
    <SelectTrigger class="w-[90px]">
      <SelectValue placeholder="Limit" />
    </SelectTrigger>
    <SelectContent>
      <SelectItem 
        v-for="option in props.options"
        :key="option" 
        :value="String(option)"
      >
        {{ option }}
      </SelectItem>
    </SelectContent>
  </Select>  
</template> 