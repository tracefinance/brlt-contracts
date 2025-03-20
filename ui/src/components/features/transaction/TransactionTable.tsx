"use client";

import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Card } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Transaction } from "@/types/models/transaction.model";
import { cn, truncateMiddle } from "@/lib/utils";
import { format } from "date-fns";
import { TokenIcon, NetworkIcon } from "@web3icons/react";
import { Hexagon } from "lucide-react";

interface TransactionTableProps {
  transactions: Transaction[];
  isLoading?: boolean;
}

export default function TransactionTable({ transactions = [], isLoading = false }: TransactionTableProps) {
  // Add a safety check for transactions being undefined
  const transactionsArray = Array.isArray(transactions) ? transactions : [];
  
  // Function to format timestamp
  const formatTimestamp = (timestamp: number) => {
    return format(new Date(timestamp * 1000), 'MMM d, yyyy HH:mm:ss');
  };

  // Function to map transaction status to badge variant
  const getStatusVariant = (status: string) => {
    switch (status.toLowerCase()) {
      case 'confirmed':
      case 'success':
        return 'success';
      case 'pending':
        return 'pending';
      case 'failed':
        return 'failed';
      default:
        return 'outline';
    }
  };
  
  return (
    <Card className="p-0 overflow-hidden">
      <div className="rounded-md">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Hash</TableHead>
              <TableHead>Block Number</TableHead>
              <TableHead>Type</TableHead>
              <TableHead>Token</TableHead>
              <TableHead>From Address</TableHead>
              <TableHead>To Address</TableHead>
              <TableHead>Status</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {isLoading ? (
              <TableRow>
                <TableCell colSpan={7} className="text-center h-24">
                  <div className="flex justify-center">
                    <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
                  </div>
                </TableCell>
              </TableRow>
            ) : transactionsArray.length === 0 ? (
              <TableRow>
                <TableCell colSpan={7} className="text-center h-24 text-muted-foreground">
                  No transactions found
                </TableCell>
              </TableRow>
            ) : (
              transactionsArray.map((tx, index) => (
                <TableRow key={`${tx.hash}-${index}`}>
                  <TableCell>
                    {truncateMiddle(tx.hash, 8, 8)}
                  </TableCell>
                  <TableCell>{tx.timestamp}</TableCell>
                  <TableCell>
                    <Badge variant="outline">                      
                      {tx.type.toUpperCase()}
                    </Badge>
                  </TableCell>
                  <TableCell>
                    {tx.tokenSymbol ? (
                      <div className="flex items-center gap-1">
                        <TokenIcon symbol={tx.tokenSymbol.toLowerCase()} size={20} variant="branded" />
                        <span>{tx.tokenSymbol}</span>
                      </div>
                    ) : (
                      <div className="flex items-center gap-1 text-gray-400">
                        <Hexagon strokeWidth={1.5} size={20} />
                        <span>{tx.type.toUpperCase()}</span>
                      </div>
                    )}
                  </TableCell>
                  <TableCell>
                    {truncateMiddle(tx.fromAddress, 6, 4)}
                  </TableCell>
                  <TableCell>
                    {truncateMiddle(tx.toAddress, 6, 4)}
                  </TableCell>
                  <TableCell>
                    <Badge variant={getStatusVariant(tx.status)}>
                      {tx.status}
                    </Badge>
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