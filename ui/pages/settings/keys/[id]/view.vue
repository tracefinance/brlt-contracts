<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { formatDateTime, shortenAddress } from '~/lib/utils'
import { toast } from 'vue-sonner'

definePageMeta({
  layout: 'settings'
})

const router = useRouter()
const route = useRoute()

const keyId = computed(() => route.params.id as string)

// Fetch Key Details
const {
  key,
  error: fetchError,
  isLoading,
  refresh
} = useKeyDetails(keyId)

// Key Mutations (for delete)
const {
  deleteKey,
  isDeleting,
  deleteError
} = useKeyMutations()

const isDeleteDialogOpen = ref(false)

const goToEditPage = () => {
  if (keyId.value) {
    router.push(`/settings/keys/${keyId.value}/edit`)
  }
}

const openDeleteDialog = () => {
  isDeleteDialogOpen.value = true
}

const handleDeleteConfirm = async () => {
  if (!key.value) {
    toast.error('Cannot delete key: Key data not available.')
    isDeleteDialogOpen.value = false
    return
  }

  const { id, name } = key.value
  const success = await deleteKey(id)

  if (success) {
    toast.success(`Key "${name}" (ID: ${shortenAddress(id)}) deleted successfully.`)
    router.push('/settings/keys') // Redirect to list after delete
  } else {
    toast.error(`Failed to delete key "${name}"`, {
      description: deleteError.value?.message || 'Unknown error'
    })
  }

  isDeleteDialogOpen.value = false
}

const copyToClipboard = (text: string | undefined) => {
  if (text) {
    navigator.clipboard.writeText(text)
    toast.success('Copied to clipboard')
  }
}

</script>

<template>
  <div class="flex justify-center">
    <Card class="w-full max-w-2xl">
      <CardHeader>
        <CardTitle>Key Details</CardTitle>
        <CardDescription>Read-only information for the selected cryptographic key.</CardDescription>
      </CardHeader>

      <div v-if="isLoading" class="flex justify-center p-6">
        <Icon name="svg-spinners:180-ring-with-bg" class="h-8 w-8" />
      </div>

      <div v-else-if="fetchError" class="p-6">
        <Alert variant="destructive">
          <Icon name="lucide:alert-triangle" class="h-4 w-4" />
          <AlertTitle>Error Loading Key Data</AlertTitle>
          <AlertDescription>
            {{ fetchError?.message || 'Failed to load key data' }}
            <Button variant="link" class="p-0 h-auto ml-1" @click="refresh">Retry</Button>
          </AlertDescription>
        </Alert>
      </div>

      <template v-else-if="key">
        <CardContent class="space-y-4">
          <div class="space-y-1">
            <Label>Key ID</Label>
            <div class="flex items-center gap-2 text-sm">
              <p class="font-mono">{{ key.id }}</p>
              <Button variant="ghost" size="icon" @click="copyToClipboard(key.id)">
                 <Icon name="lucide:copy" class="size-4"/>
              </Button>
            </div>
          </div>
           <div class="space-y-1">
            <Label>Name</Label>
            <p class="text-sm">{{ key.name }}</p>
          </div>
          <div class="space-y-1">
            <Label>Type</Label>
            <p class="text-sm uppercase">{{ key.type }}</p>
          </div>
          <div v-if="key.curve" class="space-y-1">
            <Label>Curve</Label>
            <p class="text-sm">{{ key.curve }}</p>
          </div>
          <div class="space-y-1">
            <Label>Tags</Label>
            <div v-if="key.tags && Object.keys(key.tags).length > 0" class="flex flex-wrap gap-1">
              <Badge v-for="(value, tagKey) in key.tags" :key="tagKey" variant="secondary">
                {{ tagKey }}: {{ value }}
              </Badge>
            </div>
            <p v-else class="text-sm text-muted-foreground">No tags</p>
          </div>
          <div v-if="key.publicKey" class="space-y-1">
            <Label>Public Key</Label>
             <div class="flex items-center gap-2 text-sm">
               <p class="font-mono break-all">{{ key.publicKey }}</p>
               <Button variant="ghost" size="icon" @click="copyToClipboard(key.publicKey)">
                 <Icon name="lucide:copy" class="size-4"/>
               </Button>
             </div>
           </div>
          <div class="space-y-1">
            <Label>Created At</Label>
            <p class="text-sm text-muted-foreground">{{ formatDateTime(key.createdAt) }}</p>
          </div>
        </CardContent>

        <CardFooter class="flex justify-between">
           <Button variant="destructive" @click="openDeleteDialog">
             <Icon name="lucide:trash-2" class="w-4 h-4 mr-2" />
             Delete Key
           </Button>
           <div class="flex gap-2">
              <Button variant="outline" @click="router.back()">Back</Button>
              <Button @click="goToEditPage">
                <Icon name="lucide:edit" class="w-4 h-4 mr-2" />
                Edit
              </Button>
           </div>
        </CardFooter>
      </template>

      <div v-else class="p-6">
         <Alert>
          <Icon name="lucide:search-x" class="h-4 w-4" />
          <AlertTitle>Key Not Found</AlertTitle>
          <AlertDescription>
            The requested key (ID: {{ keyId }}) could not be found.
          </AlertDescription>
        </Alert>
      </div>
    </Card>

    <!-- Delete Confirmation Dialog -->
    <AlertDialog :open="isDeleteDialogOpen" @update:open="isDeleteDialogOpen = $event">
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Are you absolutely sure?</AlertDialogTitle>
          <AlertDialogDescription>
            This action cannot be undone. This will permanently delete the key
            "{{ key?.name }}" (ID: {{ shortenAddress(key?.id || '') }}).
            This key will no longer be usable for signing or other cryptographic operations.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel :disabled="isDeleting" @click="isDeleteDialogOpen = false">Cancel</AlertDialogCancel>
          <AlertDialogAction :disabled="isDeleting" variant="destructive" @click="handleDeleteConfirm">
            <Icon v-if="isDeleting" name="svg-spinners:3-dots-fade" class="w-4 h-4 mr-2" />
            <span v-if="isDeleting">Deleting...</span>
            <span v-else>Delete Key</span>
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>

  </div>
</template> 