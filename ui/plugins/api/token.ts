import type {
  IAddTokenRequest,
  IToken,
  IListTokensRequest,
  ChainType,
  IPagedResponse,
  IUpdateTokenRequest,
} from '~/types'
import { Token } from '~/types'
import type { ApiClient } from './client'
import { API_ENDPOINTS } from './endpoints'

/**
 * Client for interacting with token-related API endpoints
 */
export class TokenClient {
  private client: ApiClient
  
  /**
   * Creates a new token client
   * @param client API client instance
   */
  constructor(client: ApiClient) {
    this.client = client
  }
  
  /**
   * Fetch a paginated list of tokens, optionally filtered.
   * @param filters - Filtering parameters (chainType, tokenType).
   * @param limit - The maximum number of items to return.
   * @param nextToken - Pagination token for the next page.
   * @returns A paged response of tokens.
   */
  async listTokens(
    filters: IListTokensRequest = {},
    limit?: number,
    nextToken?: string,
  ): Promise<IPagedResponse<IToken>> {
    const params: Record<string, any> = { ...filters }

    if (limit) params.limit = limit
    if (nextToken) params.nextToken = nextToken

    const data = await this.client.get<any>(
      API_ENDPOINTS.TOKENS.BASE,
      params,
    )
    
    return {
      items: Token.fromJsonArray(data.items || []),
      limit: data.limit,
      nextToken: data.nextToken,
    }
  }
  
  /**
   * Add a new token to the system.
   * @param request - The token details to add.
   * @returns The newly created token details.
   */
  async addToken(request: IAddTokenRequest): Promise<IToken> {
    // Base client handles toJson conversion from camelCase request to snake_case body
    const data = await this.client.post<any>(
      API_ENDPOINTS.TOKENS.BASE,
      request,
    )
    return Token.fromJson(data)
  }
  
  /**
   * Verify a token by its address.
   * This likely involves checking on-chain data or a verification source.
   * @param address - The address of the token to verify.
   * @returns The verified token details.
   */
  async verifyToken(address: string): Promise<IToken> {
    const endpoint = API_ENDPOINTS.TOKENS.VERIFY(address)
    const data = await this.client.get<any>(endpoint)
    return Token.fromJson(data)
  }
  
  /**
   * Get a specific token by its chain type and address.
   * @param chainType - The chain type of the token.
   * @param address - The address of the token.
   * @returns The token details.
   */
  async getToken(
    chainType: ChainType,
    address: string,
  ): Promise<IToken> {
    // Assuming BY_ADDRESS endpoint needs chainType and address
    const endpoint = API_ENDPOINTS.TOKENS.BY_ADDRESS(chainType, address)
    const data = await this.client.get<any>(endpoint)
    return Token.fromJson(data)
  }
  
  /**
   * Delete a token by its address.
   * @param address - The address of the token to delete.
   */
  async deleteToken(address: string): Promise<void> {
    const endpoint = API_ENDPOINTS.TOKENS.DELETE(address)
    await this.client.delete(endpoint)
  }
  
  /**
   * Update an existing token.
   * @param address - The address of the token to update.
   * @param request - The token update details.
   * @returns The updated token details.
   */
  async updateToken(address: string, request: IUpdateTokenRequest): Promise<IToken> {
    const endpoint = API_ENDPOINTS.TOKENS.UPDATE(address)
    // Base client handles toJson conversion from camelCase request to snake_case body
    const data = await this.client.put<any>(
      endpoint,
      request,
    )
    return Token.fromJson(data)
  }
} 