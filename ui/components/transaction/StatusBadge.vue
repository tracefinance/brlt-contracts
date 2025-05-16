<script setup lang="ts">
const props = defineProps<{
  status: string | null | undefined;
}>();

// Determine icon and style based on status
const iconName = computed(() => {
  const status = props.status?.toLowerCase();
  if (status === 'success') return 'lucide:check-circle';
  if (status === 'pending') return 'lucide:loader';
  if (status === 'mined') return 'lucide:package';
  if (status === 'failed') return 'lucide:x-circle';
  if (status === 'dropped') return 'lucide:trash-2';
  if (status === 'unknown') return 'lucide:help-circle';
  return 'lucide:help-circle';
});

const iconClass = computed(() => {
  const status = props.status?.toLowerCase();
  if (status === 'success') return 'text-green-600';
  if (status === 'pending') return 'animate-spin text-muted-foreground';
  if (status === 'mined') return 'text-blue-500';
  if (status === 'failed') return 'text-destructive';
  if (status === 'dropped') return 'text-amber-500';
  if (status === 'unknown') return 'text-muted-foreground';
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