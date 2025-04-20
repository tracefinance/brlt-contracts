<script setup lang="ts">
import { ref, reactive, watch, computed } from 'vue'
import { useRouter } from 'vue-router'
import { toast } from 'vue-sonner'
import { getErrorMessage } from '~/lib/utils'
import type { ICreateKeyRequest } from '~/types'
import { KEY_TYPES, CURVE_NAMES } from '~/types'

definePageMeta({
  layout: 'settings'
})

const router = useRouter()
const { 
  createKey: mutateCreateKey,
  isCreating, 
  createError 
} = useKeyMutations()

const formData = reactive<ICreateKeyRequest>({
  name: '',
  type: KEY_TYPES.ECDSA,
  curve: CURVE_NAMES.P256,
  tags: {}
})

const tagsList = ref([{ key: '', value: '' }])

const availableCurves = computed(() => {
  if (formData.type === KEY_TYPES.ECDSA) {
    return Object.values(CURVE_NAMES)
  }
  return []
})

watch(() => formData.type, (newType) => {
  if (newType !== KEY_TYPES.ECDSA) {
    formData.curve = undefined
  } else if (!formData.curve) {
    formData.curve = CURVE_NAMES.P256
  }
})

const addTag = () => {
  tagsList.value.push({ key: '', value: '' })
}

const removeTag = (index: number) => {
  tagsList.value.splice(index, 1)
}

watch(createError, (newError) => {
  if (newError) {
    toast.error(getErrorMessage(newError, 'An unknown error occurred while creating the key.'))
  }
})

const handleSubmit = async () => {
  createError.value = null

  const tags: Record<string, string> = tagsList.value
    .filter(item => item.key.trim() !== '' && item.value.trim() !== '')
    .reduce((acc, item) => {
      acc[item.key.trim()] = item.value.trim()
      return acc
    }, {} as Record<string, string>)

  const payload: ICreateKeyRequest = {
    name: formData.name.trim(),
    type: formData.type,
    curve: formData.type === KEY_TYPES.ECDSA ? formData.curve : undefined,
    tags: Object.keys(tags).length > 0 ? tags : undefined
  }

  if (!payload.name || !payload.type) {
    toast.error('Name and Type are required.')
    return
  }
  if (payload.type === KEY_TYPES.ECDSA && !payload.curve) {
    toast.error('Curve is required for EC key type.')
    return
  }

  const newKey = await mutateCreateKey(payload)

  if (newKey) {
    toast.success('Key created successfully!')
    router.push('/settings/keys')
  }
}

</script>

<template>
  <div class="flex justify-center">
    <Card class="w-full max-w-2xl">
      <CardHeader>
        <CardTitle>Create New Key</CardTitle>
        <CardDescription>Generate a new cryptographic key.</CardDescription>
      </CardHeader>
      <CardContent>
        <form class="space-y-6" @submit.prevent="handleSubmit">
          <!-- Name -->
          <div class="space-y-2">
            <Label for="name">Key Name</Label>
            <Input id="name" v-model="formData.name" required placeholder="My Signing Key" />
          </div>

          <!-- Type -->
          <div class="space-y-2">
            <Label for="type">Key Type</Label>
            <Select id="type" v-model="formData.type">
              <SelectTrigger>
                <SelectValue placeholder="Select key type" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem :value="KEY_TYPES.ECDSA">ECDSA</SelectItem>
                <SelectItem :value="KEY_TYPES.RSA">RSA</SelectItem>
                <SelectItem :value="KEY_TYPES.ED25519">Ed25519</SelectItem>
                <SelectItem :value="KEY_TYPES.SYMMETRIC">Symmetric</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <!-- Curve (Conditional) -->
          <div v-if="formData.type === KEY_TYPES.ECDSA" class="space-y-2">
            <Label for="curve">Curve</Label>
            <Select id="curve" v-model="formData.curve">
              <SelectTrigger>
                <SelectValue placeholder="Select curve" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem v-for="curveName in availableCurves" :key="curveName" :value="curveName">
                  {{ curveName }}
                </SelectItem>
              </SelectContent>
            </Select>
          </div>

          <!-- Tags -->
          <div class="space-y-4">
            <Label>Tags (Optional)</Label>
            <div v-for="(item, index) in tagsList" :key="index" class="flex items-center gap-2">
              <Input v-model="item.key" placeholder="Key" class="flex-1" />
              <Input v-model="item.value" placeholder="Value" class="flex-1" />
              <Button type="button" variant="outline" size="icon" :disabled="tagsList.length <= 1 && item.key === '' && item.value === ''" @click="removeTag(index)">
                <Icon name="lucide:trash-2" class="h-4 w-4" />
              </Button>
            </div>
            <Button type="button" variant="outline" size="sm" @click="addTag">
              <Icon name="lucide:plus" class="h-4 w-4 mr-1" />
              Add Tag
            </Button>
          </div>
        </form>
      </CardContent>
      <CardFooter class="flex justify-end gap-2">
        <NuxtLink to="/settings/keys">
          <Button variant="outline" type="button">Cancel</Button>
        </NuxtLink>
        <Button type="submit" :disabled="isCreating" @click="handleSubmit">
          <Icon v-if="isCreating" name="svg-spinners:3-dots-fade" class="w-4 h-4 mr-2" />
          {{ isCreating ? 'Creating...' : 'Create Key' }}
        </Button>
      </CardFooter>
    </Card>
  </div>
</template>