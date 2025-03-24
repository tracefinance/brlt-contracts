'use client';

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '@/lib/api/client';
import { API_ENDPOINTS } from '@/lib/api/endpoints';
import { Token, AddTokenRequest, TokenListResponse } from '@/types/models/token.model';

/**
 * Hook to fetch a list of tokens with filtering and pagination
 */
export function useTokens(
  chainType?: string, 
  tokenType?: string,
  limit: number = 10, 
  offset: number = 0
) {
  return useQuery({
    queryKey: ['tokens', chainType, tokenType, limit, offset],
    queryFn: async (): Promise<TokenListResponse> => {
      const params: Record<string, string | number> = { limit, offset };
      
      if (chainType) {
        params.chain_type = chainType;
      }
      
      if (tokenType) {
        params.token_type = tokenType;
      }
      
      const { data } = await apiClient.get(API_ENDPOINTS.TOKENS.BASE, { params });
      return TokenListResponse.fromJson(data);
    }
  });
}

/**
 * Hook to verify a token by address
 */
export function useVerifyToken(address: string) {
  return useQuery({
    queryKey: ['token', address],
    queryFn: async (): Promise<Token> => {
      const { data } = await apiClient.get(API_ENDPOINTS.TOKENS.VERIFY(address));
      return Token.fromJson(data);
    },
    enabled: !!address,
  });
}

/**
 * Hook to add a new token
 */
export function useAddToken() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async (request: AddTokenRequest): Promise<Token> => {
      const { data } = await apiClient.post(API_ENDPOINTS.TOKENS.BASE, request.toJson());
      return Token.fromJson(data);
    },
    onSuccess: () => {
      // Invalidate all token-related queries
      queryClient.invalidateQueries({ queryKey: ['tokens'] });
    }
  });
}

/**
 * Hook to delete a token by address
 */
export function useDeleteToken() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async (address: string): Promise<void> => {
      await apiClient.delete(API_ENDPOINTS.TOKENS.BY_ADDRESS(address));
    },
    onSuccess: () => {
      // Invalidate all token-related queries
      queryClient.invalidateQueries({ queryKey: ['tokens'] });
    }
  });
} 