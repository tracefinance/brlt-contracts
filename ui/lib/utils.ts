import type { Updater } from '@tanstack/vue-table'
import type { Ref } from 'vue'
import { type ClassValue, clsx } from 'clsx'
import { twMerge } from 'tailwind-merge'

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
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
export function shortenAddress(address: string, prefixLength = 6, suffixLength = 4): string {
  if (!address) return ''
  if (address.length < (prefixLength + suffixLength + 3)) return address
  return `${address.slice(0, prefixLength)}...${address.slice(-suffixLength)}`
}

/**
 * Extract a readable error message from various error formats
 * @param error The error object or string
 * @param defaultMessage Default message to show if no specific error message is found
 * @returns A formatted error message string
 */
export function getErrorMessage(error: unknown, defaultMessage = 'An unknown error occurred'): string {
  if (!error) return defaultMessage
  
  if (typeof error === 'string') {
    return error
  } else if (error instanceof Error || (error && typeof error === 'object' && 'message' in error)) {
    return String(error.message)
  }
  
  return defaultMessage
}

/**
 * Format a date string or Date object into a readable date and time format.
 * Example: "April 15, 2025 04:08:28 PM"
 * @param dateInput The date string or Date object to format.
 * @param locale Optional locale string (defaults to 'en-US').
 * @returns Formatted date string or an empty string if the input is invalid.
 */
export function formatDateTime(dateInput: string | Date, locale: string = 'en-US'): string {
  try {
    const date = typeof dateInput === 'string' ? new Date(dateInput) : dateInput;
    
    // Check if the date is valid
    if (isNaN(date.getTime())) {
      return 'Invalid Date';
    }
    
    const options: Intl.DateTimeFormatOptions = {
      month: 'long',
      day: 'numeric',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
      hour12: true
    };
    
    return date.toLocaleString(locale, options);
  } catch (e) {
    console.error("Error formatting date:", e);
    return 'Invalid Date'; // Return a default string or empty string on error
  }
}