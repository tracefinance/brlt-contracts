import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

/**
 * Truncates a string in the middle, keeping a specified number of characters
 * at the beginning and end, and adds an ellipsis in the middle.
 * 
 * @param str The string to truncate
 * @param startChars Number of characters to keep at the beginning
 * @param endChars Number of characters to keep at the end
 * @returns The truncated string
 */
export function truncateMiddle(str: string, startChars: number = 6, endChars: number = 4): string {
  if (!str) return '';
  
  // If the string is shorter than the total characters we want to keep, return it as is
  if (str.length <= startChars + endChars) {
    return str;
  }
  
  const start = str.substring(0, startChars);
  const end = str.substring(str.length - endChars);
  
  return `${start}...${end}`;
}
