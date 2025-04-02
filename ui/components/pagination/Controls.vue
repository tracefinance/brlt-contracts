<script setup lang="ts">
import { computed } from 'vue';

interface PageControlsProps {
  offset: number;
  limit: number;
  hasMore: boolean;
}

const props = defineProps<PageControlsProps>();

// Define the events emitted by this component
const emit = defineEmits<{ 
  (e: 'previous'): void 
  (e: 'next'): void 
}>();

const isFirstPage = computed(() => props.offset === 0);
const canGoNext = computed(() => props.hasMore);
</script>

<template>
  <div class="flex items-center space-x-2">
    <Button 
      variant="outline" 
      size="icon" 
      :disabled="isFirstPage" 
      @click="!isFirstPage && emit('previous')"
    >
      <span class="sr-only">Previous page</span>
      <Icon name="lucide:chevron-left" class="h-4 w-4" />
    </Button>
    <Button 
      variant="outline" 
      size="icon" 
      :disabled="!canGoNext" 
      @click="canGoNext && emit('next')"
    >
      <span class="sr-only">Next page</span>
      <Icon name="lucide:chevron-right" class="h-4 w-4" />
    </Button>
  </div>
</template> 