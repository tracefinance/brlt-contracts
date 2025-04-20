import type {
  IAddTokenRequest,
  IToken,
  IListTokensRequestParams,
  ChainType,
  IPagedResponse,
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
   * @param params - Filtering, pagination, and limit parameters.
   * @returns A paged response of tokens.
   */
  async listTokens(
    params: IListTokensRequestParams = {},
  ): Promise<IPagedResponse<IToken>> {
    // Convert camelCase params to snake_case for the backend API
    const snakeCaseParams: Record<string, any> = {}
    if (params.limit !== undefined) {
      snakeCaseParams.limit = params.limit
    }
    if (params.nextToken !== undefined) {
      snakeCaseParams.next_token = params.nextToken
    }
    if (params.chainType !== undefined) {
      snakeCaseParams.chain_type = params.chainType
    }
    if (params.tokenType !== undefined) {
      snakeCaseParams.token_type = params.tokenType
    }

    const data = await this.client.get<any>(
      API_ENDPOINTS.TOKENS.BASE,
      snakeCaseParams,
    )

    // Use Token factory
    return {
      items: Token.fromJsonArray(data.items || []),
      limit: data.limit,
      nextToken: data.next_token, // Use snake_case from response
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
   * Delete a token by its address and chain type.
   * @param chainType - The chain type of the token.
   * @param address - The address of the token to delete.
   */
  async deleteToken(chainType: ChainType, address: string): Promise<void> {
    // Assuming BY_ADDRESS endpoint needs chainType and address
    const endpoint = API_ENDPOINTS.TOKENS.BY_ADDRESS(chainType, address)
    // Use generic delete which doesn't expect a specific return type
    await this.client.delete(endpoint)
  }
} 