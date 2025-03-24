'use client';

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '@/lib/api/client';
import { API_ENDPOINTS } from '@/lib/api/endpoints';
import { Wallet, CreateWalletRequest, UpdateWalletRequest, PagedWallets } from '@/types/models/wallet.model';

/**
 * Hook to fetch a wallet by address and chain type
 */
export function useWallet(address: string, chainType: string) {
  return useQuery({
    queryKey: ['wallet', chainType, address],
    queryFn: async (): Promise<Wallet> => {
      const { data } = await apiClient.get(API_ENDPOINTS.WALLETS.BY_ADDRESS(address, chainType));
      return Wallet.fromJson(data);
    },
    enabled: !!address && !!chainType,
  });
}

/**
 * Hook to fetch a list of wallets with pagination
 */
export function useWallets(limit: number = 10, offset: number = 0) {
  return useQuery({
    queryKey: ['wallets', limit, offset],
    queryFn: async (): Promise<PagedWallets> => {
      const { data } = await apiClient.get(API_ENDPOINTS.WALLETS.BASE, {
        params: { limit, offset }
      });
      return PagedWallets.fromJson(data);
    }
  });
}

/**
 * Hook to create a new wallet
 */
export function useCreateWallet() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async (request: CreateWalletRequest): Promise<Wallet> => {
      const { data } = await apiClient.post(API_ENDPOINTS.WALLETS.BASE, request.toJson());
      return Wallet.fromJson(data);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['wallets'] });
    }
  });
}

/**
 * Hook to update an existing wallet
 */
export function useUpdateWallet(address: string, chainType: string) {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async (request: UpdateWalletRequest): Promise<Wallet> => {
      const { data } = await apiClient.put(
        API_ENDPOINTS.WALLETS.BY_ADDRESS(address, chainType), 
        request.toJson()
      );
      return Wallet.fromJson(data);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['wallets'] });
      queryClient.invalidateQueries({ queryKey: ['wallet', chainType, address] });
    }
  });
}

/**
 * Hook to delete a wallet
 */
export function useDeleteWallet() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async ({ address, chainType }: { address: string, chainType: string }): Promise<void> => {
      await apiClient.delete(API_ENDPOINTS.WALLETS.BY_ADDRESS(address, chainType));
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['wallets'] });
    }
  });
}

/**
 * Hook to fetch a wallet's balance
 */
export function useWalletBalance(address: string, chainType: string) {
  return useQuery({
    queryKey: ['wallet', chainType, address, 'balance'],
    queryFn: async () => {
      const { data } = await apiClient.get(API_ENDPOINTS.WALLETS.BALANCE(address, chainType));
      return data; // Return the response directly as it's already parsed by Axios
    },
    enabled: !!address && !!chainType,
  });
} 