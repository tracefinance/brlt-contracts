"use client";

import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import WalletTable from "@/components/features/wallet/WalletTable";
import WalletForm from "@/components/features/wallet/WalletForm";
import { WalletFrontend, CreateWalletRequest, UpdateWalletRequest } from "@/types/wallet";
import { walletApi } from "@/lib/api";
import { useToast } from "@/components/ui/use-toast";
import { PlusCircle } from "lucide-react";

export default function WalletsPage() {
  const { toast } = useToast();
  const [wallets, setWallets] = useState<WalletFrontend[]>([]);
  const [loading, setLoading] = useState(true);
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [editingWallet, setEditingWallet] = useState<WalletFrontend | null>(null);
  
  // Pagination state
  const [limit] = useState(10);
  const [offset, setOffset] = useState(0);
  const [hasMore, setHasMore] = useState(false);
  
  // Fetch wallets on component mount
  useEffect(() => {
    fetchWallets();
  }, [offset, limit]);

  async function fetchWallets() {
    try {
      setLoading(true);
      const result = await walletApi.getWallets(limit, offset);
      // Update wallets and pagination state
      setWallets(result.wallets || []);
      setHasMore(result.hasMore);
    } catch (error) {
      console.error("Failed to fetch wallets:", error);
      toast({
        title: "Error",
        description: "Failed to load wallets",
        variant: "destructive",
      });
      // Set empty array on error
      setWallets([]);
      setHasMore(false);
    } finally {
      setLoading(false);
    }
  }

  async function handleCreateWallet(data: CreateWalletRequest) {
    try {
      const newWallet = await walletApi.createWallet(data);
      setWallets([...wallets, newWallet]);
      setIsCreateOpen(false);
      toast({
        title: "Success",
        description: "Wallet created successfully",
      });
    } catch (error) {
      console.error("Failed to create wallet:", error);
      toast({
        title: "Error",
        description: "Failed to create wallet",
        variant: "destructive",
      });
      throw error;
    }
  }

  async function handleUpdateWallet(data: UpdateWalletRequest) {
    if (!editingWallet) return;
    
    try {
      const updated = await walletApi.updateWallet(editingWallet.chainType, editingWallet.address, data);
      setWallets(wallets.map(w => w.id === updated.id ? updated : w));
      setEditingWallet(null);
      toast({
        title: "Success",
        description: "Wallet updated successfully",
      });
    } catch (error) {
      console.error("Failed to update wallet:", error);
      toast({
        title: "Error",
        description: "Failed to update wallet",
        variant: "destructive",
      });
      throw error;
    }
  }

  async function handleDeleteWallet(wallet: WalletFrontend) {
    try {
      await walletApi.deleteWallet(wallet.chainType, wallet.address);
      setWallets(wallets.filter(w => w.id !== wallet.id));
      toast({
        title: "Success",
        description: "Wallet deleted successfully",
      });
    } catch (error) {
      console.error("Failed to delete wallet:", error);
      toast({
        title: "Error",
        description: "Failed to delete wallet",
        variant: "destructive",
      });
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold tracking-tight">Wallets</h1>
        <Button onClick={() => setIsCreateOpen(true)}>
          <PlusCircle className="h-4 w-4 mr-2" />
          Add Wallet
        </Button>
      </div>
      
      {loading ? (
        <div className="flex justify-center p-8">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
        </div>
      ) : (
        <WalletTable 
          wallets={wallets}
          onEdit={setEditingWallet}
          onDelete={handleDeleteWallet}
        />
      )}
      
      {/* Pagination Controls */}
      <div className="flex justify-between">
        <Button 
          variant="outline" 
          onClick={() => setOffset(Math.max(0, offset - limit))}
          disabled={offset === 0}
        >
          Previous
        </Button>
        
        <Button 
          variant="outline" 
          onClick={() => setOffset(offset + limit)}
          disabled={!hasMore}
        >
          Next
        </Button>
      </div>
      
      {/* Create Wallet Dialog */}
      <Dialog open={isCreateOpen} onOpenChange={setIsCreateOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Add New Wallet</DialogTitle>
          </DialogHeader>
          <WalletForm 
            onSubmit={handleCreateWallet}
            onCancel={() => setIsCreateOpen(false)}
          />
        </DialogContent>
      </Dialog>
      
      {/* Edit Wallet Dialog */}
      <Dialog open={!!editingWallet} onOpenChange={(open) => !open && setEditingWallet(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Edit Wallet</DialogTitle>
          </DialogHeader>
          {editingWallet && (
            <WalletForm 
              wallet={editingWallet}
              onSubmit={handleUpdateWallet}
              onCancel={() => setEditingWallet(null)}
            />
          )}
        </DialogContent>
      </Dialog>
    </div>
  );
} 