<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import type { IUpdateKeyRequest } from '~/types'
import { toast } from 'vue-sonner'
import { getErrorMessage } from '~/lib/utils'

definePageMeta({
  layout: 'settings'
})

const route = useRoute()
const router = useRouter()
const keyId = computed(() => route.params.id as string)

// Fetch existing key details
const { 
  key,
  isLoading: isLoadingKey,
  error: fetchError,
  refresh: refreshKey 
} = useKeyDetails(keyId)

// Use key mutations for update
const { 
  updateKey: mutateUpdateKey,
  isUpdating,
  updateError
} = useKeyMutations()

// Form state
const keyName = ref('')
const tagsList = ref([{ key: '', value: '' }])

// Watch for key data to load and populate the form
watch(key, (newKey) => {
  if (newKey) {
    keyName.value = newKey.name || ''
    const loadedTags = newKey.tags ? Object.entries(newKey.tags) : []
    if (loadedTags.length > 0) {
      tagsList.value = loadedTags.map(([key, value]) => ({ key, value }))
    } else {
      tagsList.value = [{ key: '', value: '' }]
    }
  } else {
    keyName.value = ''
    tagsList.value = [{ key: '', value: '' }]
  }
}, { immediate: true })

// Watch for mutation errors
watch(updateError, (newError) => {
  if (newError) {
    toast.error(getErrorMessage(newError, 'An unknown error occurred while saving the key.'))
  }
})

// Metadata helpers
const addTag = () => {
  tagsList.value.push({ key: '', value: '' })
}

const removeTag = (index: number) => {
  if (tagsList.value.length > 1) {
     tagsList.value.splice(index, 1)
  } else if (tagsList.value.length === 1) {
    tagsList.value = [{ key: '', value: '' }]
  }
}

// Handle form submission
const handleSaveChanges = async () => {
  if (!keyId.value || !key.value) {
    toast.error('Cannot save, key context is invalid or data is missing.')
    return
  }

  if (!keyName.value.trim()) {
    toast.error('Key Name is required.')
    return
  }

  updateError.value = null

  // Convert tagsList back to Record<string, string>
  const tagsPayload: Record<string, string> = tagsList.value
    .map(item => ({ key: item.key.trim(), value: item.value.trim() }))
    .filter(item => item.key !== '')
    .reduce((acc, item) => {
      acc[item.key] = item.value
      return acc
    }, {} as Record<string, string>)

  const payload: IUpdateKeyRequest = {
    name: keyName.value.trim(),
    tags: Object.keys(tagsPayload).length > 0 ? tagsPayload : undefined
  }

  const updatedKey = await mutateUpdateKey(keyId.value, payload)

  if (updatedKey) {
    toast.success('Key updated successfully!')
    router.push(`/settings/keys/${keyId.value}`) // Go back to details view
  }  
}

</script>

<template>
  <div class="flex justify-center">
    <Card class="w-full max-w-2xl">
      <CardHeader>
        <CardTitle>Edit Key</CardTitle>
         <CardDescription v-if="key" class="flex items-center gap-2 pt-1">
           <Icon name="lucide:lock" class="size-4" />
           <span class="font-mono">{{ key.id }}</span>
         </CardDescription>
        <CardDescription v-else-if="isLoadingKey">Loading key details...</CardDescription>
        <CardDescription v-else-if="fetchError">Error loading key.</CardDescription>
        <CardDescription v-else>Key details unavailable.</CardDescription>
      </CardHeader>
      
      <CardContent>
        <div v-if="isLoadingKey" class="flex items-center justify-center p-8">
          <Icon name="svg-spinners:pulse-3" class="w-6 h-6 mr-2" />
          <span>Loading key details...</span>
        </div>

        <div v-else-if="fetchError" class="my-4">
          <Alert variant="destructive">
            <Icon name="lucide:alert-triangle" class="w-4 h-4" />
            <AlertTitle>Error Loading Key</AlertTitle>
            <AlertDescription>
              {{ fetchError.message || 'Failed to load key details.' }}
               <Button variant="link" size="sm" class="p-0 h-auto mt-1" @click="refreshKey">Retry</Button>
            </AlertDescription>
          </Alert>
        </div>
        
        <form v-else-if="key" class="space-y-6" @submit.prevent="handleSaveChanges">
          <!-- Key Name -->
          <div class="space-y-2">
            <Label for="key-name">Key Name</Label>
            <Input 
              id="key-name" 
              v-model="keyName" 
              placeholder="e.g. My Primary Signing Key"
              required 
            />
          </div>
          
          <!-- Tags (Renamed from Metadata) -->
          <div class="space-y-4">
            <Label>Tags (Optional)</Label>
            <div v-for="(item, index) in tagsList" :key="index" class="flex items-center gap-2">
              <Input v-model="item.key" placeholder="Key (e.g. environment)" class="flex-1" />
              <Input v-model="item.value" placeholder="Value (e.g. production)" class="flex-1" />
              <Button 
                type="button" 
                variant="outline" 
                size="icon" 
                :disabled="tagsList.length === 1 && (!item.key && !item.value)"
                aria-label="Remove Tag" 
                @click="removeTag(index)"
              >
                <Icon name="lucide:trash-2" class="h-4 w-4" />
              </Button>
            </div>
            <Button type="button" variant="outline" size="sm" @click="addTag">
              <Icon name="lucide:plus" class="h-4 w-4 mr-1" />
              Add Tag
            </Button>
          </div>        
        </form>
        
        <div v-else-if="!isLoadingKey && !fetchError">
           <p class="text-muted-foreground">Could not load key data.</p>
        </div>

      </CardContent>
      
      <CardFooter v-if="!isLoadingKey && !fetchError && key" class="flex justify-end space-x-2">
        <Button variant="outline" :disabled="isUpdating" type="button" @click="router.back()">Cancel</Button>
        <Button
          type="submit"
          :disabled="isUpdating || !keyName.trim()"
          @click="handleSaveChanges"
        >
          <Icon v-if="isUpdating" name="svg-spinners:3-dots-fade" class="w-4 h-4 mr-2" />
          {{ isUpdating ? 'Saving...' : 'Save Changes' }}
        </Button>
      </CardFooter>
    </Card>
  </div>
</template> 