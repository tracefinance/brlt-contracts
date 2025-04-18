<script setup lang="ts">
import type { PropType } from 'vue'

// Define props for the component
defineProps({
  type: {
    type: String as PropType<'internal' | 'external'>,
    required: true
  }
})
</script>

<template>
  <Badge variant="outline" class="rounded-full capitalize py-1 px-2">
    <Icon v-if="type === 'internal'" name="lucide:shield" class="h-3 w-3 mr-1" />
    <Icon v-else-if="type === 'external'" name="lucide:key" class="h-3 w-3 mr-1" />
    {{ type }}
  </Badge>
</template> 