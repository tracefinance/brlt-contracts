"use client";

import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Button } from "@/components/ui/button";
import { MoreHorizontal, Pencil, Trash2 } from "lucide-react";
import { WalletFrontend } from "@/types/wallet";
import { Card } from "@/components/ui/card";
import { cn, truncateMiddle } from "@/lib/utils";
import { useRouter } from "next/navigation";

interface WalletTableProps {
  wallets: WalletFrontend[];
  onEdit: (wallet: WalletFrontend) => void;
  onDelete: (wallet: WalletFrontend) => void;
}

export default function WalletTable({ wallets = [], onEdit, onDelete }: WalletTableProps) {
  const router = useRouter();
  
  // Add a safety check for wallets being undefined
  const walletsArray = Array.isArray(wallets) ? wallets : [];
  
  // Function to navigate to wallet details page
  const navigateToWalletDetails = (wallet: WalletFrontend) => {
    router.push(`/wallets/${encodeURIComponent(wallet.chainType)}/${encodeURIComponent(wallet.address)}`);
  };
  
  return (
    <Card>
      <div className="rounded-md">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Address</TableHead>
              <TableHead>Chain Type</TableHead>
              <TableHead className="w-[80px]">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {walletsArray.length === 0 ? (
              <TableRow>
                <TableCell colSpan={4} className="text-center h-24 text-muted-foreground">
                  No wallets found
                </TableCell>
              </TableRow>
            ) : (
              walletsArray.map((wallet) => (
                <TableRow 
                  key={wallet.id} 
                  className="cursor-pointer hover:bg-muted/50"
                  onClick={() => navigateToWalletDetails(wallet)}
                >
                  <TableCell className="font-medium">{wallet.name}</TableCell>
                  <TableCell className={cn("font-mono text-sm max-w-[250px]", wallet.address && "text-muted-foreground")}>
                    {wallet.address ? truncateMiddle(wallet.address, 10, 8) : "Generating..."}
                  </TableCell>
                  <TableCell>{wallet.chainType}</TableCell>
                  <TableCell onClick={(e) => e.stopPropagation()}>
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button variant="ghost" size="icon">
                          <MoreHorizontal className="h-4 w-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                        <DropdownMenuItem onClick={() => onEdit(wallet)}>
                          <Pencil className="h-4 w-4 mr-2" /> Edit
                        </DropdownMenuItem>
                        <DropdownMenuItem onClick={() => onDelete(wallet)}>
                          <Trash2 className="h-4 w-4 mr-2" /> Delete
                        </DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>
    </Card>
  );
} 