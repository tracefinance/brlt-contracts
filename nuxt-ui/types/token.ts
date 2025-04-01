import { 
  Expose,
  Type
} from 'class-transformer';
import { BaseModel, fromJson, fromJsonArray } from './model';

export class Token extends BaseModel {
  @Expose()
  id!: number;
  
  @Expose({ name: 'chain_type' })
  chainType!: string;
  
  @Expose()
  address!: string;
  
  @Expose()
  name!: string;
  
  @Expose()
  symbol!: string;
  
  @Expose()
  decimals!: number;
  
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