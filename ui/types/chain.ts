import { fromJson, fromJsonArray } from './model';

/**
 * Interface representing a blockchain chain reference.
 * Based on backend's ChainResponse DTO.
 */
export interface IChain {
  id: number;
  type: string;
  layer: string;
  name: string;
  symbol: string;
  explorerUrl: string;
}

/**
 * Factory functions for IChain
 */
export const Chain = {
  /**
   * Converts a plain JSON object (snake_case) from the API to an IChain (camelCase)
   */
  fromJson(json: any): IChain {
    return fromJson<IChain>(json); // Uses the generic helper which handles snake_case to camelCase
  },

  /**
   * Converts an array of plain JSON objects from the API to IChain objects
   */
  fromJsonArray(jsonArray: any[]): IChain[] {
    return fromJsonArray<IChain>(jsonArray); // Uses the generic helper
  }
}; 