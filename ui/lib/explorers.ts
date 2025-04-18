/**
 * Generates a transaction explorer URL based on the base URL and transaction hash.
 * Handles common patterns for Etherscan/Polygonscan-like explorers.
 * 
 * @param explorerBaseUrl - The base URL of the blockchain explorer (e.g., "https://etherscan.io").
 * @param txHash - The transaction hash.
 * @returns The full transaction explorer URL or undefined if inputs are invalid.
 */
export function getTransactionExplorerUrl(
  explorerBaseUrl: string | undefined | null, 
  txHash: string | undefined | null
): string | undefined {
  if (!explorerBaseUrl || !txHash) return undefined;

  const baseUrl = explorerBaseUrl.replace(/\/$/, ''); // Remove trailing slash
  
  // Common patterns
  if (baseUrl.includes('etherscan') || baseUrl.includes('polygonscan')) {
    return `${baseUrl}/tx/${txHash}`;
  }
  
  // Fallback: Assume hash can be appended directly (might not work for all explorers)
  return `${baseUrl}/${txHash}`;
}

/**
 * Generates an address explorer URL based on the base URL and address.
 * Handles common patterns for Etherscan/Polygonscan-like explorers.
 * 
 * @param explorerBaseUrl - The base URL of the blockchain explorer.
 * @param address - The blockchain address.
 * @returns The full address explorer URL or undefined if inputs are invalid.
 */
export function getAddressExplorerUrl(
  explorerBaseUrl: string | undefined | null, 
  address: string | undefined | null
): string | undefined {
  if (!explorerBaseUrl || !address) return undefined;

  const baseUrl = explorerBaseUrl.replace(/\/$/, ''); // Remove trailing slash

  // Common patterns
  if (baseUrl.includes('etherscan') || baseUrl.includes('polygonscan')) {
    return `${baseUrl}/address/${address}`;
  }

  // Fallback: Assume address can be appended directly
  return `${baseUrl}/address/${address}`;
}

/**
 * Generates a block explorer URL based on the base URL and block number.
 * Handles common patterns for Etherscan/Polygonscan-like explorers.
 * 
 * @param explorerBaseUrl - The base URL of the blockchain explorer.
 * @param blockNumber - The block number.
 * @returns The full block explorer URL or undefined if inputs are invalid.
 */
export function getBlockExplorerUrl(
  explorerBaseUrl: string | undefined | null, 
  blockNumber: number | string | undefined | null
): string | undefined {
  if (!explorerBaseUrl || blockNumber === undefined || blockNumber === null) return undefined;

  const baseUrl = explorerBaseUrl.replace(/\/$/, ''); // Remove trailing slash
  const blockNumStr = typeof blockNumber === 'number' ? blockNumber.toString() : blockNumber;

  // Common patterns
  if (baseUrl.includes('etherscan') || baseUrl.includes('polygonscan')) {
    return `${baseUrl}/block/${blockNumStr}`;
  }

  // Fallback: Assume block number can be appended directly
  // This might need adjustment based on other explorers
  return `${baseUrl}/block/${blockNumStr}`;
} 