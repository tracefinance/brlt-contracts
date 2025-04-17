<script setup lang="ts">
import { reactive } from 'vue'
import { useRouter } from 'vue-router'
import type { ICreateUserRequest } from '~/types'
import { toast } from 'vue-sonner'

definePageMeta({
  layout: 'settings'
})

const router = useRouter()
const { 
  createUser: mutateCreateUser,
  isCreating, 
  error: mutationError 
} = useUserMutations()

const formData = reactive<ICreateUserRequest>({
  email: '',
  password: ''
})

const handleSubmit = async () => {
  mutationError.value = null

  const payload: ICreateUserRequest = {
    email: formData.email.trim(),
    password: formData.password
  }

  // Basic validation
  if (!payload.email || !payload.password) {
    toast.error('Email and Password are required.')
    return
  }

  const newUser = await mutateCreateUser(payload)

  if (newUser) {
    toast.success('User created successfully!')
    router.push('/settings/users')
  }
}
</script>

<template>
  <div class="flex justify-center">
    <Card class="w-full max-w-2xl">
      <CardHeader>
        <CardTitle>Create New User</CardTitle>
      </CardHeader>
      <CardContent>
        <form class="space-y-6" @submit.prevent="handleSubmit">
          <div class="space-y-2">
            <Label for="email">Email</Label>
            <Input id="email" v-model="formData.email" type="email" required placeholder="user@example.com" />
          </div>

          <div class="space-y-2">
            <Label for="password">Password</Label>
            <Input id="password" v-model="formData.password" type="password" required placeholder="••••••••" />
          </div>
        </form>
      </CardContent>
      <CardFooter class="flex justify-end gap-2">
         <NuxtLink to="/settings/users">
            <Button variant="outline">Cancel</Button>
          </NuxtLink>
        <Button type="submit" :disabled="isCreating" @click="handleSubmit">
          <Icon v-if="isCreating" name="svg-spinners:3-dots-fade" class="w-4 h-4 mr-2" />
          {{ isCreating ? 'Creating...' : 'Create User' }}
        </Button>
      </CardFooter>
    </Card>
  </div>
</template> 