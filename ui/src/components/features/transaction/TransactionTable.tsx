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
import { TransactionFrontend } from "@/types/transaction";
import { cn, truncateMiddle } from "@/lib/utils";
import { format } from "date-fns";
import { TokenIcon } from "@web3icons/react";
import { CircleDollarSign } from "lucide-react";

interface TransactionTableProps {
  transactions: TransactionFrontend[];
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
    <Card>
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
                  <TableCell className="font-mono text-sm">
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
                        <TokenIcon symbol={tx.tokenSymbol.toLowerCase()} size={20} variant="mono" />
                        <span>{tx.tokenSymbol}</span>
                      </div>
                    ) : (
                      <div className="flex items-center gap-1">
                        <CircleDollarSign className="h-5 w-5 text-muted-foreground" />
                        <span className="text-muted-foreground">-</span>
                      </div>
                    )}
                  </TableCell>
                  <TableCell className="font-mono text-sm">
                    {truncateMiddle(tx.fromAddress, 6, 4)}
                  </TableCell>
                  <TableCell className="font-mono text-sm">
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