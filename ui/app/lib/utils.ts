import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";


export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

/**
 * Formats a number as currency with a maximum of 6 decimal places
 * and minimum of 2 decimal places
 */
export function formatCurrency(value: string | number): string {
  if (!value) return '$0';
  
  const num = typeof value === 'string' ? parseFloat(value) : value;
  
  // Check if the number is valid
  if (isNaN(num)) return '$0';
  
  return '$' + num.toLocaleString(undefined, {
    maximumFractionDigits: 6,
    minimumFractionDigits: 4
  });
}

/**
 * Shortens an Ethereum address for display
 * @param address The full Ethereum address
 * @param startChars Number of characters to show at the beginning (default: 6)
 * @param endChars Number of characters to show at the end (default: 4)
 * @returns Shortened address (e.g., "0x1234...abcd")
 */
export function shortenAddress(address: string, startChars = 6, endChars = 4): string {
    if (!address) return '';
    const prefix = address.startsWith('0x') ? '0x' : '';
    const body = address.startsWith('0x') ? address.substring(2) : address;
    
    if (body.length <= startChars + endChars) {
        return address; // Address is too short to shorten
    }
    
    return `${prefix}${body.substring(0, startChars)}...${body.substring(body.length - endChars)}`;
}
