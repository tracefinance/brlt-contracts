<script setup lang="ts">
const props = defineProps<{
  status: string | null | undefined;
}>();

// Determine icon and style based on status
const iconName = computed(() => {
  const status = props.status?.toLowerCase();
  if (status === 'success') return 'lucide:check-circle';
  if (status === 'pending') return 'lucide:loader';
  if (status === 'failed') return 'lucide:x-circle';
  return 'lucide:help-circle';
});

const iconClass = computed(() => {
  const status = props.status?.toLowerCase();
  if (status === 'success') return 'text-green-600';
  if (status === 'pending') return 'animate-spin text-muted-foreground';
  if (status === 'failed') return 'text-destructive';
  return 'text-muted-foreground';
});
</script>

<template>
  <Badge variant="outline" class="rounded-full px-2 py-1 capitalize">
    <Icon 
      :name="iconName" 
      class="mr-1 h-4 w-4" 
      :class="iconClass" 
    />
    {{ props.status || 'Unknown' }}
  </Badge>
</template> 