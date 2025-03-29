import { 
  apiGet, 
  apiPost, 
  apiPut, 
  apiDelete 
} from './client';
import { 
  Token, 
  AddTokenRequest, 
  TokenListResponse 
} from '~/models/token';
import { API_ENDPOINTS } from './endpoints';

/**
 * Lists all tokens with optional filtering and pagination
 * @param token Authentication token
 * @param chainType Optional chain type to filter by
 * @param tokenType Optional token type to filter by
 * @param limit Maximum number of tokens to return (default: 10)
 * @param offset Number of tokens to skip for pagination (default: 0)
 * @returns Paginated list of tokens
 */
export async function listTokens(
  token: string,
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
  
  const data = await apiGet<any>(API_ENDPOINTS.TOKENS.BASE, params, token);
  return TokenListResponse.fromJson(data);
}

/**
 * Adds a new token
 * @param request Token creation request data
 * @param token Authentication token
 * @returns Created token
 */
export async function addToken(
  request: AddTokenRequest,
  token: string
): Promise<Token> {
  const data = await apiPost<any>(API_ENDPOINTS.TOKENS.BASE, request.toJson(), token);
  return Token.fromJson(data);
}

/**
 * Verifies a token by its address and chain type
 * @param address Token address
 * @param chainType Blockchain network type
 * @param token Authentication token
 * @returns Token details
 */
export async function verifyToken(
  address: string,
  chainType: string,
  token: string
): Promise<Token> {
  const endpoint = API_ENDPOINTS.TOKENS.VERIFY(address, chainType);
  const data = await apiGet<any>(endpoint, undefined, token);
  return Token.fromJson(data);
}

/**
 * Gets a token by its address and chain type
 * @param address Token address
 * @param chainType Blockchain network type
 * @param token Authentication token
 * @returns Token details
 */
export async function getToken(
  address: string,
  chainType: string,
  token: string
): Promise<Token> {
  const endpoint = API_ENDPOINTS.TOKENS.BY_ADDRESS(address, chainType);
  const data = await apiGet<any>(endpoint, undefined, token);
  return Token.fromJson(data);
}

/**
 * Deletes a token
 * @param address Token address
 * @param chainType Blockchain network type
 * @param token Authentication token
 */
export async function deleteToken(
  address: string,
  chainType: string,
  token: string
): Promise<void> {
  const endpoint = API_ENDPOINTS.TOKENS.BY_ADDRESS(address, chainType);
  await apiDelete<void>(endpoint, token);
}

/**
 * Gets a token by its ID
 * @param id Token ID 
 * @param token Authentication token
 * @returns Token details
 */
export async function getTokenById(
  id: string,
  token: string
): Promise<Token> {
  const endpoint = API_ENDPOINTS.TOKENS.BY_ID(id);
  const data = await apiGet<any>(endpoint, undefined, token);
  return Token.fromJson(data);
} 