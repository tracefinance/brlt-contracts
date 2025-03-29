import { 
  Expose 
} from 'class-transformer';
import { BaseModel, fromJson, fromJsonArray } from './model';
import { Token } from './token';

export class Wallet extends BaseModel {
  @Expose()
  id!: number;
  
  @Expose({ name: 'key_id' })
  keyId!: string;
  
  @Expose({ name: 'chain_type' })
  chainType!: string;
  
  @Expose()
  address!: string;
  
  @Expose()
  name!: string;
  
  @Expose()
  tags?: Record<string, string>;
  
  @Expose({ name: 'created_at' })
  createdAt!: string;
  
  @Expose({ name: 'updated_at' })
  updatedAt!: string;

  constructor(data: Partial<Wallet> = {}) {
    super();
    Object.assign(this, data);
  }

  /**
   * Converts a plain JSON object from the API to a Wallet instance
   */
  static fromJson(json: any): Wallet {
    return fromJson(Wallet, json);
  }

  /**
   * Converts an array of plain JSON objects from the API to Wallet instances
   */
  static fromJsonArray(jsonArray: any[]): Wallet[] {
    return fromJsonArray(Wallet, jsonArray);
  }
}

/**
 * Response type for token balance endpoints
 */
export class TokenBalanceResponse extends BaseModel {
  @Expose()
  token!: Token;
  
  @Expose()
  balance!: string;
  
  @Expose({ name: 'updated_at' })
  updatedAt!: string;

  constructor(data: Partial<TokenBalanceResponse> = {}) {
    super();
    Object.assign(this, data);
  }

  static fromJson(json: any): TokenBalanceResponse {
    return fromJson(TokenBalanceResponse, json);
  }

  static fromJsonArray(jsonArray: any[]): TokenBalanceResponse[] {
    return fromJsonArray(TokenBalanceResponse, jsonArray);
  }
}

/**
 * Class representing a request to create a wallet
 */
export class CreateWalletRequest extends BaseModel {
  @Expose({ name: 'chain_type' })
  chainType: string;
  
  @Expose()
  name: string;
  
  @Expose()
  tags?: Record<string, string>;

  constructor(chainType: string, name: string, tags?: Record<string, string>) {
    super();
    this.chainType = chainType;
    this.name = name;
    this.tags = tags;
  }
}

/**
 * Class representing a request to update a wallet
 */
export class UpdateWalletRequest extends BaseModel {
  @Expose()
  name: string;
  
  @Expose()
  tags?: Record<string, string>;

  constructor(name: string, tags?: Record<string, string>) {
    super();
    this.name = name;
    this.tags = tags;
  }
}

/**
 * Class representing a paginated response containing Wallets
 */
export class PagedWallets extends BaseModel {
  @Expose()
  items!: Wallet[];

  @Expose()
  limit!: number;

  @Expose()
  offset!: number;

  @Expose({ name: 'has_more' })
  hasMore!: boolean;

  constructor(data: Partial<PagedWallets> = {}) {
    super();
    Object.assign(this, data);
    this.items = data.items || [];
  }

  /**
   * Converts a plain JSON paged response to a PagedWallets instance
   */
  static fromJson(json: any): PagedWallets {
    const response = fromJson(PagedWallets, json);
    
    // Convert each item in the items array
    response.items = Wallet.fromJsonArray(json.items || []);

    return response;
  }
} 