import type {
  IAddTokenRequest,
  IPagedResponse,
  IToken,
} from '~/types';
import {
  Token,
  fromJsonArray
} from '~/types';
import type {
  ApiClient
} from './client';
import { API_ENDPOINTS } from './endpoints';

/**
 * Client for interacting with token-related API endpoints
 */
export class TokenClient {
  private client: ApiClient;
  
  /**
   * Creates a new token client
   * @param client API client instance
   */
  constructor(client: ApiClient) {
    this.client = client;
  }
  
  /**
   * Lists all tokens with optional filtering and token-based pagination
   * @param chainType Optional chain type to filter by
   * @param tokenType Optional token type to filter by
   * @param limit Maximum number of tokens to return (default: 10)
   * @param nextToken Token for retrieving the next page of results (default: undefined)
   * @returns Paginated list of tokens
   */
  async listTokens(
    chainType?: string,
    tokenType?: string,
    limit: number = 10,
    nextToken?: string
  ): Promise<IPagedResponse<IToken>> {
    const params: Record<string, any> = { limit };
    
    if (nextToken) {
      params.next_token = nextToken;
    }
    
    if (chainType) {
      params.chain_type = chainType;
    }
    
    if (tokenType) {
      params.token_type = tokenType;
    }
    
    const data = await this.client.get<any>(API_ENDPOINTS.TOKENS.BASE, params);
    return {
      items: fromJsonArray<IToken>(data.items || []),
      limit: data.limit,
      nextToken: data.nextToken
    };
  }
  
  /**
   * Adds a new token
   * @param request Token creation request
   * @returns Created token
   */
  async addToken(request: IAddTokenRequest): Promise<IToken> {
    const data = await this.client.post<any>(API_ENDPOINTS.TOKENS.BASE, request);
    return Token.fromJson(data);
  }
  
  /**
   * Verifies a token by its address and chain type
   * @param address Token address
   * @returns Token details
   */
  async verifyToken(address: string): Promise<IToken> {
    const endpoint = API_ENDPOINTS.TOKENS.VERIFY(address);
    const data = await this.client.get<any>(endpoint);
    return Token.fromJson(data);
  }
  
  /**
   * Gets a token by its address and chain type
   * @param chainType Blockchain network type
   * @param address Token address
   * @returns Token details
   */
  async getToken(chainType: string, address: string): Promise<IToken> {
    const endpoint = API_ENDPOINTS.TOKENS.BY_ADDRESS(chainType, address);
    const data = await this.client.get<any>(endpoint);
    return Token.fromJson(data);
  }
  
  /**
   * Deletes a token
   * @param address Token address
   * @param chainType Blockchain network type
   */
  async deleteToken(address: string): Promise<void> {
    const endpoint = API_ENDPOINTS.TOKENS.DELETE(address);
    await this.client.delete(endpoint);
  }
} 