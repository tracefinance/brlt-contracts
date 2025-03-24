'use client';

import { ReactNode } from 'react';
import { ThemeProvider } from '@/components/core/providers/ThemeProvider';
import { WalletProvider } from '@/providers/WalletProvider';
import { TokenProvider } from '@/providers/TokenProvider';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { Toaster } from 'sonner';

// Create a client
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 5 * 60 * 1000, // 5 minutes
      retry: 1,
    },
  },
});

export function Providers({ children }: { children: ReactNode }) {
  return (
    <QueryClientProvider client={queryClient}>
      <ThemeProvider attribute="class" defaultTheme="system" enableSystem>
        <WalletProvider>
          <TokenProvider>
            <Toaster position="top-right" />
            {children}
          </TokenProvider>
        </WalletProvider>
      </ThemeProvider>
    </QueryClientProvider>
  );
} 