import type { Updater } from '@tanstack/vue-table'
import type { Ref } from 'vue'
import { type ClassValue, clsx } from 'clsx'
import { twMerge } from 'tailwind-merge'

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export function valueUpdater<T extends Updater<any>>(updaterOrValue: T, ref: Ref) {
  ref.value
    = typeof updaterOrValue === 'function'
      ? updaterOrValue(ref.value)
      : updaterOrValue
}

/**
 * Ethereum zero address constant
 */
export const ZERO_ADDRESS = '0x0000000000000000000000000000000000000000'

/**
 * Format currency with appropriate decimal places
 */
export function formatCurrency(value: string | number): string {
  if (!value) return '$0';
  
  const num = typeof value === 'string' ? parseFloat(value) : value;
  
  // Check if the number is valid
  if (isNaN(num)) return '$0';
  
  return '$' + num.toLocaleString(undefined, {
    maximumFractionDigits: 6,
    minimumFractionDigits: 2
  });
}

/**
 * Shorten an Ethereum address or hash to a displayable format
 * @param address The address or hash to shorten
 * @param prefixLength Number of characters to keep at the start
 * @param suffixLength Number of characters to keep at the end
 * @returns Shortened string with ellipsis
 */
export function shortenAddress(address: string, prefixLength = 4, suffixLength = 4): string {
  if (!address) return ''
  if (address.length < (prefixLength + suffixLength + 3)) return address
  return `${address.slice(0, prefixLength)}...${address.slice(-suffixLength)}`
}