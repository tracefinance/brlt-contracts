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
    minimumFractionDigits: 2
  });
}
