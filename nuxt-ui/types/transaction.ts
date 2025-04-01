import { 
  Expose 
} from 'class-transformer';
import { BaseModel, fromJson, fromJsonArray } from './model';

export class Transaction extends BaseModel {
  @Expose()
  id!: number;
  
  @Expose({ name: 'wallet_id' })
  walletId!: number;
  
  @Expose({ name: 'chain_type' })
  chainType!: string;
  
  @Expose()
  hash!: string;
  
  @Expose({ name: 'from_address' })
  fromAddress!: string;
  
  @Expose({ name: 'to_address' })
  toAddress!: string;
  
  @Expose()
  value!: string;
  
  @Expose()
  data?: string;
  
  @Expose()
  nonce!: number;
  
  @Expose({ name: 'gas_price' })
  gasPrice?: string;
  
  @Expose({ name: 'gas_limit' })
  gasLimit?: number;
  
  @Expose()
  type!: string;
  
  @Expose({ name: 'token_address' })
  tokenAddress?: string;
  
  @Expose({ name: 'token_symbol' })
  tokenSymbol?: string;
  
  @Expose()
  status!: string;
  
  @Expose()
  timestamp!: number;
  
  @Expose({ name: 'created_at' })
  createdAt!: string;
  
  @Expose({ name: 'updated_at' })
  updatedAt!: string;

  constructor(data: Partial<Transaction> = {}) {
    super();
    Object.assign(this, data);
  }

  /**
   * Converts a plain JSON object from the API to a Transaction instance
   */
  static fromJson(json: any): Transaction {
    return fromJson(Transaction, json);
  }

  /**
   * Converts an array of plain JSON objects from the API to Transaction instances
   */
  static fromJsonArray(jsonArray: any[]): Transaction[] {
    return fromJsonArray(Transaction, jsonArray);
  }
}

/**
 * Class representing a paginated response containing Transactions
 */
export class PagedTransactions extends BaseModel {
  @Expose()
  items!: Transaction[];

  @Expose()
  limit!: number;

  @Expose()
  offset!: number;

  @Expose({ name: 'has_more' })
  hasMore!: boolean;

  constructor(data: Partial<PagedTransactions> = {}) {
    super();
    Object.assign(this, data);
    this.items = data.items || [];
  }

  /**
   * Converts a plain JSON paged response to a PagedTransactions instance
   */
  static fromJson(json: any): PagedTransactions {
    const response = fromJson(PagedTransactions, json);

    // Convert each item in the items array
    response.items = Transaction.fromJsonArray(json.items || []);

    return response;
  }
}

/**
 * Class representing a transaction sync response
 */
export class SyncTransactionsResponse extends BaseModel {
  @Expose()
  count!: number;

  constructor(data: Partial<SyncTransactionsResponse> = {}) {
    super();
    Object.assign(this, data);
  }

  /**
   * Converts a plain JSON sync response to a SyncTransactionsResponse instance
   */
  static fromJson(json: any): SyncTransactionsResponse {
    return fromJson(SyncTransactionsResponse, json);
  }
} 