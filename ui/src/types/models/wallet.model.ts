import { 
  plainToInstance, 
  instanceToPlain, 
  Expose 
} from 'class-transformer';

export class Wallet {
  @Expose()
  id: number;
  
  @Expose({ name: 'key_id' })
  keyId: string;
  
  @Expose({ name: 'chain_type' })
  chainType: string;
  
  @Expose()
  address: string;
  
  @Expose()
  name: string;
  
  @Expose()
  tags?: Record<string, string>;
  
  @Expose({ name: 'created_at' })
  createdAt: string;
  
  @Expose({ name: 'updated_at' })
  updatedAt: string;

  /**
   * Converts a plain JSON object from the API to a Wallet instance
   */
  static fromJson(json: any): Wallet {
    return plainToInstance(Wallet, json, {
      excludeExtraneousValues: true
    });
  }

  /**
   * Converts an array of plain JSON objects from the API to Wallet instances
   */
  static fromJsonArray(jsonArray: any[]): Wallet[] {
    return jsonArray.map(json => Wallet.fromJson(json));
  }

  /**
   * Converts this Wallet instance to a plain JSON object for sending to the API
   */
  toJson(): any {
    return instanceToPlain(this, {
      excludeExtraneousValues: true
    });
  }
}

/**
 * Class representing a request to create a wallet
 */
export class CreateWalletRequest {
  @Expose({ name: 'chain_type' })
  chainType: string;
  
  @Expose()
  name: string;
  
  @Expose()
  tags?: Record<string, string>;

  constructor(chainType: string, name: string, tags?: Record<string, string>) {
    this.chainType = chainType;
    this.name = name;
    this.tags = tags;
  }

  /**
   * Converts this CreateWalletRequest instance to a plain JSON object for sending to the API
   */
  toJson(): any {
    return instanceToPlain(this, {
      excludeExtraneousValues: true
    });
  }
}

/**
 * Class representing a request to update a wallet
 */
export class UpdateWalletRequest {
  @Expose()
  name: string;
  
  @Expose()
  tags?: Record<string, string>;

  constructor(name: string, tags?: Record<string, string>) {
    this.name = name;
    this.tags = tags;
  }

  /**
   * Converts this UpdateWalletRequest instance to a plain JSON object for sending to the API
   */
  toJson(): any {
    return instanceToPlain(this, {
      excludeExtraneousValues: true
    });
  }
}

/**
 * Class representing a paginated response containing Wallets
 */
export class PagedWallets {
  @Expose()
  items: Wallet[];

  @Expose()
  limit: number;

  @Expose()
  offset: number;

  @Expose({ name: 'has_more' })
  hasMore: boolean;

  /**
   * Converts a plain JSON paged response to a PagedWallets instance
   */
  static fromJson(json: any): PagedWallets {
    const response = plainToInstance(PagedWallets, json, {
      excludeExtraneousValues: true
    });

    // Convert each item in the items array
    response.items = Wallet.fromJsonArray(json.items || []);

    return response;
  }
} 