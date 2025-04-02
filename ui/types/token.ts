import { fromJson, fromJsonArray, toJson } from './model';

/**
 * Interface representing a blockchain token
 */
export interface IToken {
  id: string;
  address: string;
  chainType: string;
  tokenType: string;
  name: string;
  symbol: string;
  decimals: number;
  logo?: string;
  createdAt: string;
  updatedAt: string;
}

/**
 * Factory functions for IToken
 */
export const Token = {
  /**
   * Converts a plain JSON object from the API to an IToken
   */
  fromJson(json: any): IToken {
    return fromJson<IToken>(json);
  },

  /**
   * Converts an array of plain JSON objects from the API to IToken objects
   */
  fromJsonArray(jsonArray: any[]): IToken[] {
    return fromJsonArray<IToken>(jsonArray);
  }
};

/**
 * Interface for adding a new token
 */
export interface IAddTokenRequest {
  address: string;
  chainType: string;
  tokenType: string;
  name: string;
  symbol: string;
  decimals: number;
  logo?: string;
}

/**
 * Factory functions for IAddTokenRequest
 */
export const AddTokenRequest = {
  create(
    address: string,
    chainType: string,
    tokenType: string,
    name: string,
    symbol: string,
    decimals: number,
    logo?: string
  ): IAddTokenRequest {
    return {
      address,
      chainType,
      tokenType,
      name,
      symbol,
      decimals,
      logo
    };
  },
  
  toJson(request: IAddTokenRequest): any {
    return toJson(request);
  }
};

/**
 * Interface for paginated token list
 */
export interface IPagedTokens {
  items: IToken[];
  limit: number;
  offset: number;
  hasMore: boolean;
}

/**
 * Factory functions for IPagedTokens
 */
export const PagedTokens = {
  fromJson(json: any): IPagedTokens {
    const response = fromJson<IPagedTokens>(json);
    
    // Convert each item in the items array
    if (json.items && Array.isArray(json.items)) {
      response.items = Token.fromJsonArray(json.items);
    } else {
      response.items = [];
    }
    
    return response;
  }
}; 