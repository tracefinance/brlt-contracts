"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import { walletApi, transactionApi } from "@/lib/api";
import { useToast } from "@/components/ui/use-toast";
import { WalletFrontend } from "@/types/wallet";
import { TransactionFrontend } from "@/types/transaction";
import { ChevronLeft, RefreshCw } from "lucide-react";
import WalletDetail from "@/components/features/wallet/WalletDetail";
import TransactionTable from "@/components/features/transaction/TransactionTable";

export default function WalletDetailsPage() {
  const params = useParams();
  const router = useRouter();
  const { toast } = useToast();
  
  const chainType = decodeURIComponent(params.chainType as string);
  const address = decodeURIComponent(params.address as string);
  
  const [wallet, setWallet] = useState<WalletFrontend | null>(null);
  const [transactions, setTransactions] = useState<TransactionFrontend[]>([]);
  const [isWalletLoading, setIsWalletLoading] = useState(true);
  const [isTransactionsLoading, setIsTransactionsLoading] = useState(true);
  const [isSyncing, setIsSyncing] = useState(false);
  
  // Pagination state
  const [limit] = useState(10);
  const [offset, setOffset] = useState(0);
  const [hasMore, setHasMore] = useState(false);
  
  // Fetch wallet details on component mount
  useEffect(() => {
    fetchWallet();
  }, [chainType, address]);
  
  // Fetch transactions on component mount or when pagination changes
  useEffect(() => {
    fetchTransactions();
  }, [chainType, address, offset, limit]);
  
  async function fetchWallet() {
    try {
      setIsWalletLoading(true);
      const result = await walletApi.getWallet(chainType, address);
      setWallet(result);
    } catch (error) {
      console.error("Failed to fetch wallet:", error);
      toast({
        title: "Error",
        description: "Failed to load wallet details",
        variant: "destructive",
      });
      // Navigate back to wallets list on error
      router.push("/wallets");
    } finally {
      setIsWalletLoading(false);
    }
  }
  
  async function fetchTransactions() {
    try {
      setIsTransactionsLoading(true);
      const result = await transactionApi.getTransactionsByWallet(chainType, address, limit, offset);
      setTransactions(result.transactions);
      setHasMore(result.hasMore);
    } catch (error) {
      console.error("Failed to fetch transactions:", error);
      toast({
        title: "Error",
        description: "Failed to load transactions",
        variant: "destructive",
      });
      setTransactions([]);
      setHasMore(false);
    } finally {
      setIsTransactionsLoading(false);
    }
  }
  
  async function handleSyncTransactions() {
    try {
      setIsSyncing(true);
      const count = await transactionApi.syncTransactions(chainType, address);
      
      // Refresh transactions after sync
      await fetchTransactions();
      
      toast({
        title: "Success",
        description: `Synced ${count} new transactions`,
      });
    } catch (error) {
      console.error("Failed to sync transactions:", error);
      toast({
        title: "Error",
        description: "Failed to sync transactions",
        variant: "destructive",
      });
    } finally {
      setIsSyncing(false);
    }
  }
  
  return (
    <div className="space-y-6">
      <div className="flex items-center gap-2">
        <Button variant="outline" size="icon" onClick={() => router.push("/wallets")}>
          <ChevronLeft className="h-4 w-4" />
        </Button>
        <h1 className="text-2xl font-bold tracking-tight">Wallet Details</h1>
      </div>
      
      {isWalletLoading ? (
        <div className="flex justify-center p-8">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
        </div>
      ) : wallet ? (
        <WalletDetail wallet={wallet} />
      ) : null}
      
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-semibold">Transactions</h2>
        <Button 
          variant="outline" 
          onClick={handleSyncTransactions} 
          disabled={isSyncing}
        >
          <RefreshCw className={`h-4 w-4 mr-2 ${isSyncing ? "animate-spin" : ""}`} />
          {isSyncing ? "Syncing..." : "Sync Transactions"}
        </Button>
      </div>
      
      <TransactionTable 
        transactions={transactions} 
        isLoading={isTransactionsLoading} 
      />
      
      {/* Pagination Controls */}
      <div className="flex justify-between">
        <Button 
          variant="outline" 
          onClick={() => setOffset(Math.max(0, offset - limit))}
          disabled={offset === 0 || isTransactionsLoading}
        >
          Previous
        </Button>
        
        <Button 
          variant="outline" 
          onClick={() => setOffset(offset + limit)}
          disabled={!hasMore || isTransactionsLoading}
        >
          Next
        </Button>
      </div>
    </div>
  );
} 