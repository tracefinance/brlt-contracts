'use client';

import { createContext, useContext, useState, ReactNode, useCallback } from 'react';
import { useTokens, useAddToken, useDeleteToken } from '@/hooks/api/useToken';
import { Token, AddTokenRequest } from '@/types/models/token.model';
import { toast } from 'sonner';

interface TokenContextType {
  isLoading: boolean;
  tokens: Token[];
  selectedToken: Token | null;
  addToken: (address: string, chainType: string, symbol: string, decimals: number, type: string) => Promise<void>;
  deleteToken: (address: string) => Promise<void>;
  selectToken: (token: Token | null) => void;
  filterTokens: (chainType?: string, tokenType?: string) => void;
}

const TokenContext = createContext<TokenContextType | null>(null);

export function TokenProvider({ children }: { children: ReactNode }) {
  const [selectedToken, setSelectedToken] = useState<Token | null>(null);
  const [chainTypeFilter, setChainTypeFilter] = useState<string | undefined>(undefined);
  const [tokenTypeFilter, setTokenTypeFilter] = useState<string | undefined>(undefined);
  
  // Fetch tokens using the custom hook
  const { data: tokenList, isLoading } = useTokens(chainTypeFilter, tokenTypeFilter);
  const addTokenMutation = useAddToken();
  const deleteTokenMutation = useDeleteToken();
  
  // Add a new token
  const addToken = useCallback(async (
    address: string, 
    chainType: string, 
    symbol: string, 
    decimals: number, 
    type: string
  ) => {
    try {
      const request = new AddTokenRequest(address, chainType, symbol, decimals, type);
      const newToken = await addTokenMutation.mutateAsync(request);
      toast.success(`Token ${symbol} added successfully`);
      
      // Select the newly added token
      setSelectedToken(newToken);
    } catch (error) {
      toast.error('Failed to add token');
      throw error;
    }
  }, [addTokenMutation]);
  
  // Delete a token
  const deleteToken = useCallback(async (address: string) => {
    try {
      await deleteTokenMutation.mutateAsync(address);
      
      // If the deleted token was selected, clear the selection
      if (selectedToken && selectedToken.address === address) {
        setSelectedToken(null);
      }
      
      toast.success('Token deleted successfully');
    } catch (error) {
      toast.error('Failed to delete token');
      throw error;
    }
  }, [deleteTokenMutation, selectedToken]);
  
  // Select a token
  const selectToken = useCallback((token: Token | null) => {
    setSelectedToken(token);
  }, []);
  
  // Filter tokens
  const filterTokens = useCallback((chainType?: string, tokenType?: string) => {
    setChainTypeFilter(chainType);
    setTokenTypeFilter(tokenType);
  }, []);
  
  return (
    <TokenContext.Provider
      value={{
        isLoading,
        tokens: tokenList?.items || [],
        selectedToken,
        addToken,
        deleteToken,
        selectToken,
        filterTokens,
      }}
    >
      {children}
    </TokenContext.Provider>
  );
}

// Custom hook to use the token context
export function useTokenContext() {
  const context = useContext(TokenContext);
  
  if (!context) {
    throw new Error('useTokenContext must be used within a TokenProvider');
  }
  
  return context;
} 