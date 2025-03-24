'use client';

import { createContext, useContext, useState, ReactNode, useCallback } from 'react';
import { useCreateWallet, useWallets, useDeleteWallet } from '@/hooks/api/useWallet';
import { Wallet, CreateWalletRequest } from '@/types/models/wallet.model';
import { toast } from 'sonner';

interface WalletContextType {
  isLoading: boolean;
  wallets: Wallet[];
  selectedWallet: Wallet | null;
  createWallet: (chainType: string, name: string, tags?: Record<string, string>) => Promise<void>;
  deleteWallet: (address: string, chainType: string) => Promise<void>;
  selectWallet: (wallet: Wallet | null) => void;
}

const WalletContext = createContext<WalletContextType | null>(null);

export function WalletProvider({ children }: { children: ReactNode }) {
  const [selectedWallet, setSelectedWallet] = useState<Wallet | null>(null);
  
  // Fetch wallets using the custom hook
  const { data: walletPage, isLoading } = useWallets();
  const createWalletMutation = useCreateWallet();
  const deleteWalletMutation = useDeleteWallet();
  
  // Create a new wallet
  const createWallet = useCallback(async (chainType: string, name: string, tags?: Record<string, string>) => {
    try {
      const request = new CreateWalletRequest(chainType, name, tags);
      const newWallet = await createWalletMutation.mutateAsync(request);
      toast.success(`Wallet "${name}" created successfully`);
      
      // Select the newly created wallet
      setSelectedWallet(newWallet);
    } catch (error) {
      toast.error('Failed to create wallet');
      throw error;
    }
  }, [createWalletMutation]);
  
  // Delete a wallet
  const deleteWallet = useCallback(async (address: string, chainType: string) => {
    try {
      await deleteWalletMutation.mutateAsync({ address, chainType });
      
      // If the deleted wallet was selected, clear the selection
      if (selectedWallet && selectedWallet.address === address && selectedWallet.chainType === chainType) {
        setSelectedWallet(null);
      }
      
      toast.success('Wallet deleted successfully');
    } catch (error) {
      toast.error('Failed to delete wallet');
      throw error;
    }
  }, [deleteWalletMutation, selectedWallet]);
  
  // Select a wallet
  const selectWallet = useCallback((wallet: Wallet | null) => {
    setSelectedWallet(wallet);
  }, []);
  
  return (
    <WalletContext.Provider
      value={{
        isLoading,
        wallets: walletPage?.items || [],
        selectedWallet,
        createWallet,
        deleteWallet,
        selectWallet,
      }}
    >
      {children}
    </WalletContext.Provider>
  );
}

// Custom hook to use the wallet context
export function useWalletContext() {
  const context = useContext(WalletContext);
  
  if (!context) {
    throw new Error('useWalletContext must be used within a WalletProvider');
  }
  
  return context;
} 