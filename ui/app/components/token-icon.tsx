import { TokenIcon as Web3TokenIcon } from "@web3icons/react";

/**
 * Common prefixes used for wrapped tokens
 */
const WRAPPED_TOKEN_PREFIXES = ['W', 's', 'c', 'a', 'v', 't'];

interface TokenIconProps {
  symbol: string;
  size?: number;
  className?: string;
}

/**
 * TokenIcon component that handles wrapped token symbols by removing common prefixes
 * like 'W' (WBTC -> BTC), 's' (sETH -> ETH), etc.
 */
export function TokenIcon({ symbol, ...props }: TokenIconProps) {
  const normalizedSymbol = normalizeTokenSymbol(symbol);
  return <Web3TokenIcon symbol={normalizedSymbol} {...props} />;
}

/**
 * Normalizes token symbols by removing common wrapped token prefixes
 * 
 * Examples:
 * - WBTC -> BTC
 * - sETH -> ETH
 * - cUSDC -> USDC
 * - aDAI -> DAI
 */
function normalizeTokenSymbol(symbol: string): string {
  if (!symbol) return symbol;
  
  // Convert to uppercase for consistent handling
  const upperSymbol = symbol.toUpperCase();
  
  // Check for common wrapped token prefixes
  for (const prefix of WRAPPED_TOKEN_PREFIXES) {
    const upperPrefix = prefix.toUpperCase();
    if (upperSymbol.startsWith(upperPrefix) && upperSymbol.length > upperPrefix.length) {
      // Make sure we're not removing part of the actual token name
      // For example, we should not convert WAVES to AVES
      const potentialSymbol = upperSymbol.slice(upperPrefix.length);
      if (isCommonTokenSymbol(potentialSymbol)) {
        return potentialSymbol;
      }
    }
  }
  
  // Handle tokens that start with an "x" prefix (like xSUSHI)
  if (upperSymbol.startsWith('X') && upperSymbol.length > 1) {
    const potentialSymbol = upperSymbol.slice(1);
    if (isCommonTokenSymbol(potentialSymbol)) {
      return potentialSymbol;
    }
  }
  
  // Return original symbol if no patterns match
  return symbol;
}

/**
 * Check if a symbol is a common token symbol
 * This helps prevent incorrect prefix removal
 */
function isCommonTokenSymbol(symbol: string): boolean {
  // List of common token symbols
  const commonTokens = [
    'BTC', 'ETH', 'USDC', 'USDT', 'DAI', 'LINK', 'UNI', 'AAVE',
    'SUSHI', 'YFI', 'SNX', 'COMP', 'MKR', 'BAT', 'LTC', 'DOT',
    'SOL', 'AVAX', 'MATIC', 'ADA', 'XRP', 'DOGE', 'SHIB', 'LUNA',
    'ATOM', 'FTM', 'NEAR', 'ALGO', 'XTZ', 'EOS', 'XLM'
  ];
  
  return commonTokens.includes(symbol);
} 