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