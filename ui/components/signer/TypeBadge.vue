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
  <Badge variant="outline" class="rounded-full capitalize py-1 px-2 gap-1">
    <Icon v-if="type === 'internal'" name="lucide:shield" class="size-4" />
    <Icon v-else-if="type === 'external'" name="lucide:key" class="size-4" />
    {{ type }}
  </Badge>
</template> 