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
import { cn } from "@/lib/utils";

interface WalletTableProps {
  wallets: WalletFrontend[];
  onEdit: (wallet: WalletFrontend) => void;
  onDelete: (wallet: WalletFrontend) => void;
}

export default function WalletTable({ wallets = [], onEdit, onDelete }: WalletTableProps) {
  // Add a safety check for wallets being undefined
  const walletsArray = Array.isArray(wallets) ? wallets : [];
  
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
                <TableRow key={wallet.id}>
                  <TableCell className="font-medium">{wallet.name}</TableCell>
                  <TableCell className={cn("font-mono text-sm max-w-[250px] truncate", wallet.address && "text-muted-foreground")}>
                    {wallet.address || "Generating..."}
                  </TableCell>
                  <TableCell>{wallet.chainType}</TableCell>
                  <TableCell>
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