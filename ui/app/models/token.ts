import { Expose } from 'class-transformer';
import { BaseModel, fromJson, fromJsonArray } from './model';

/**
 * Class representing a token
 */
export class Token extends BaseModel {
  @Expose()
  address!: string;
  
  @Expose({ name: 'chain_type' })
  chainType!: string;
  
  @Expose()
  symbol!: string;
  
  @Expose()
  decimals!: number;
  
  @Expose()
  type!: string;
  
  constructor(data: Partial<Token> = {}) {
    super();
    Object.assign(this, data);
  }
  
  /**
   * Converts a plain JSON object from the API to a Token instance
   */
  static fromJson(json: any): Token {
    return fromJson(Token, json);
  }
  
  /**
   * Converts an array of plain JSON objects from the API to Token instances
   */
  static fromJsonArray(jsonArray: any[]): Token[] {
    return fromJsonArray(Token, jsonArray);
  }
}

/**
 * Class representing a request to add a token
 */
export class AddTokenRequest extends BaseModel {
  @Expose()
  address: string;
  
  @Expose({ name: 'chain_type' })
  chainType: string;
  
  @Expose()
  symbol: string;
  
  @Expose()
  decimals: number;
  
  @Expose()
  type: string;
  
  constructor(address: string, chainType: string, symbol: string, decimals: number, type: string) {
    super();
    this.address = address;
    this.chainType = chainType;
    this.symbol = symbol;
    this.decimals = decimals;
    this.type = type;
  }
}

/**
 * Class representing a paginated response containing Tokens
 */
export class TokenListResponse extends BaseModel {
  @Expose()
  items!: Token[];
  
  @Expose()
  total!: number;
  
  constructor(data: Partial<TokenListResponse> = {}) {
    super();
    Object.assign(this, data);
    this.items = data.items || [];
  }
  
  /**
   * Converts a plain JSON paged response to a TokenListResponse instance
   */
  static fromJson(json: any): TokenListResponse {
    const response = fromJson(TokenListResponse, json);
    
    // Convert each item in the items array
    response.items = Token.fromJsonArray(json.items || []);
    
    return response;
  }
} 