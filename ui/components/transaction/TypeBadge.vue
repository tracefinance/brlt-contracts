<script setup lang="ts">
const props = defineProps<{
  walletAddress: string;
  fromAddress: string;
}>();

// Determine if this is a send or receive transaction
const isSend = computed(() => {
  return props.fromAddress.toLowerCase() === props.walletAddress.toLowerCase();
});

// Get icon name based on transaction type
const iconName = computed(() => {
  return isSend.value ? 'lucide:arrow-up-right' : 'lucide:arrow-down-left';
});
</script>

<template>
  <Badge variant="outline" class="rounded-full px-2 py-1">
    <Icon 
      :name="iconName" 
      class="mr-1 h-4 w-4" 
    />
    {{ isSend ? 'Send' : 'Receive' }}
  </Badge>
</template> 