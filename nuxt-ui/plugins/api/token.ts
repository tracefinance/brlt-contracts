import {
  AddTokenRequest,
  Token,
  TokenListResponse
} from '~/types/token';
import {
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
   * Lists all tokens with optional filtering and pagination
   * @param chainType Optional chain type to filter by
   * @param tokenType Optional token type to filter by
   * @param limit Maximum number of tokens to return (default: 10)
   * @param offset Number of tokens to skip for pagination (default: 0)
   * @returns Paginated list of tokens
   */
  async listTokens(
    chainType?: string,
    tokenType?: string,
    limit: number = 10,
    offset: number = 0
  ): Promise<TokenListResponse> {
    const params: Record<string, string | number | boolean> = {
      limit,
      offset
    };
    
    if (chainType) {
      params.chain_type = chainType;
    }
    
    if (tokenType) {
      params.token_type = tokenType;
    }
    
    const data = await this.client.get<any>(API_ENDPOINTS.TOKENS.BASE, params);
    return TokenListResponse.fromJson(data);
  }
  
  /**
   * Adds a new token
   * @param request Token creation request data
   * @returns Created token
   */
  async addToken(request: AddTokenRequest): Promise<Token> {
    const data = await this.client.post<any>(API_ENDPOINTS.TOKENS.BASE, request.toJson());
    return Token.fromJson(data);
  }
  
  /**
   * Verifies a token by its address and chain type
   * @param address Token address
   * @returns Token details
   */
  async verifyToken(address: string): Promise<Token> {
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
  async getToken(chainType: string, address: string): Promise<Token> {
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
    await this.client.delete<void>(endpoint);
  }
} 