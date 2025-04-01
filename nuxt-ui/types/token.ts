import { 
  Expose,
  Type
} from 'class-transformer';
import { BaseModel, fromJson, fromJsonArray, toJson } from './model';

/**
 * Token model representing a blockchain token
 */
export class Token extends BaseModel {
  @Expose()
  id!: string;
  
  @Expose()
  address!: string;
  
  @Expose({ name: 'chain_type' })
  chainType!: string;
  
  @Expose({ name: 'token_type' })
  tokenType!: string;
  
  @Expose()
  name!: string;
  
  @Expose()
  symbol!: string;
  
  @Expose()
  decimals!: number;
  
  @Expose()
  logo?: string;
  
  @Expose({ name: 'created_at' })
  createdAt!: string;
  
  @Expose({ name: 'updated_at' })
  updatedAt!: string;
  
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
 * Request model for adding a new token
 */
export class AddTokenRequest extends BaseModel {
  @Expose()
  address!: string;
  
  @Expose({ name: 'chain_type' })
  chainType!: string;
  
  @Expose({ name: 'token_type' })
  tokenType!: string;
  
  @Expose()
  name!: string;
  
  @Expose()
  symbol!: string;
  
  @Expose()
  decimals!: number;
  
  @Expose()
  logo?: string;
  
  constructor(
    address: string,
    chainType: string,
    tokenType: string,
    name: string,
    symbol: string,
    decimals: number,
    logo?: string
  ) {
    super();
    this.address = address;
    this.chainType = chainType;
    this.tokenType = tokenType;
    this.name = name;
    this.symbol = symbol;
    this.decimals = decimals;
    this.logo = logo;
  }
  
  override toJson(): any {
    return toJson(this);
  }
}

/**
 * Response model for paginated token list
 */
export class TokenListResponse extends BaseModel {
  @Expose()
  @Type(() => Token)
  items!: Token[];
  
  @Expose()
  total!: number;
  
  @Expose()
  limit!: number;
  
  @Expose()
  offset!: number;
  
  constructor(data: Partial<TokenListResponse> = {}) {
    super();
    Object.assign(this, data);
  }
  
  static fromJson(json: any): TokenListResponse {
    return fromJson(TokenListResponse, json);
  }
} 