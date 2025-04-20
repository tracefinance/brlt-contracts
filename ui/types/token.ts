import type { ChainType, TokenType } from '~/types';
import { fromJson, fromJsonArray, toJson } from './model';

/**
 * Represents a token entity as used in the frontend.
 * Matches backend structure but uses camelCase.
 */
export interface IToken {
  address: string;
  chainType: ChainType;
  symbol: string;
  decimals: number;
  type: TokenType;
  name?: string; 
  logoURI?: string;
  verified?: boolean;
  createdAt?: string;
  updatedAt?: string;
}

/**
 * Frontend representation for the AddToken request body.
 */
export interface IAddTokenRequest {
  address: string;
  chainType: ChainType;
  symbol: string;
  decimals: number;
  type: TokenType;
}

/**
 * Represents the query parameters for the list tokens request.
 */
export interface IListTokensRequestParams {
  chainType?: ChainType;
  tokenType?: TokenType;
  nextToken?: string;
  limit?: number;
}

/**
 * Factory functions for IToken
 */
export const Token = {
  /**
   * Converts JSON (typically snake_case from API) to IToken (camelCase).
   */
  fromJson(json: any): IToken {
    const token = fromJson<IToken>(json);
    return token;
  },

  /**
   * Converts an array of JSON objects to an array of IToken.
   */
  fromJsonArray(jsonArray: any[]): IToken[] {
    return fromJsonArray<IToken>(jsonArray);
  }
};

/**
 * Factory functions for IAddTokenRequest
 */
export const AddTokenRequest = {
  /**
   * Creates an IAddTokenRequest object.
   */
  create(
    address: string,
    chainType: ChainType,
    symbol: string,
    decimals: number,
    type: TokenType,
  ): IAddTokenRequest {
    return { address, chainType, symbol, decimals, type };
  },

  /**
   * Converts IAddTokenRequest (camelCase) to JSON (snake_case for API).
   */
  toJson(request: IAddTokenRequest): any {
    return toJson(request);
  }
};
