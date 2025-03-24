import { plainToInstance, instanceToPlain, Expose } from 'class-transformer';

/**
 * Class representing a token
 */
export class Token {
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
  
  /**
   * Converts a plain JSON object from the API to a Token instance
   */
  static fromJson(json: any): Token {
    return plainToInstance(Token, json, {
      excludeExtraneousValues: true
    });
  }
  
  /**
   * Converts an array of plain JSON objects from the API to Token instances
   */
  static fromJsonArray(jsonArray: any[]): Token[] {
    return jsonArray.map(json => Token.fromJson(json));
  }
  
  /**
   * Converts this Token instance to a plain JSON object for sending to the API
   */
  toJson(): any {
    return instanceToPlain(this, {
      excludeExtraneousValues: true
    });
  }
}

/**
 * Class representing a request to add a token
 */
export class AddTokenRequest {
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
    this.address = address;
    this.chainType = chainType;
    this.symbol = symbol;
    this.decimals = decimals;
    this.type = type;
  }
  
  /**
   * Converts this AddTokenRequest instance to a plain JSON object for sending to the API
   */
  toJson(): any {
    return instanceToPlain(this, {
      excludeExtraneousValues: true
    });
  }
}

/**
 * Class representing a paginated response containing Tokens
 */
export class TokenListResponse {
  @Expose()
  items: Token[];
  
  @Expose()
  total: number;
  
  /**
   * Converts a plain JSON paged response to a TokenListResponse instance
   */
  static fromJson(json: any): TokenListResponse {
    const response = plainToInstance(TokenListResponse, json, {
      excludeExtraneousValues: true
    });
    
    // Convert each item in the items array
    response.items = Token.fromJsonArray(json.items || []);
    
    return response;
  }
} 