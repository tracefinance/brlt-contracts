<script setup lang="ts">
import { toast } from 'vue-sonner'
import type { IWallet } from '~/types'
import QrcodeVue from 'qrcode.vue'

// Define props
interface Props {
  currentWallet: IWallet | null | undefined
}

const props = defineProps<Props>()

// Function to copy address to clipboard
const copyAddress = async () => {
  if (!props.currentWallet) return
  try {
    await navigator.clipboard.writeText(props.currentWallet.address)
    toast.success('Address copied to clipboard!')
  } catch (err) {
    console.error('Failed to copy address: ', err)
    toast.error('Failed to copy address.')
  }
}
</script>

<template>
  <Dialog v-if="currentWallet">
    <DialogTrigger as-child>
      <Button>Receive</Button>
    </DialogTrigger>
    <DialogContent class="sm:max-w-lg">
      <DialogHeader>
        <DialogTitle>Receive Funds</DialogTitle>
        <DialogDescription>
          Share your address or QR code to receive funds.
        </DialogDescription>
      </DialogHeader>
      <div class="flex items-center justify-center space-x-2 py-4">
        <!-- QR Code -->
        <div class="bg-white rounded-md p-2">
          <qrcode-vue :value="currentWallet.address" :size="128" level="H" />
        </div>
      </div>
      <DialogFooter class="sm:justify-center items-center gap-2 border border-gray rounded-md p-2">
        <p class="text-sm font-medium text-center break-all flex-1">
          {{ currentWallet.address }}
        </p>
        <Button variant="ghost" size="icon" @click="copyAddress" class="shrink-0">
          <Icon name="lucide:copy" class="h-4 w-4" />
          <span class="sr-only">Copy Address</span>
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>
