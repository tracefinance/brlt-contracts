import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "~/components/ui/table";
import type { Transaction } from "~/models/transaction";
import { TransactionRow } from "./transaction-row"; // Import the newly created component

interface TransactionTableProps {
  transactions: Transaction[];
  walletAddress?: string;
  explorerBaseUrl: string;
}

export function TransactionTable({ transactions, walletAddress, explorerBaseUrl }: TransactionTableProps) {
  return (
    <Table>
      <TableHeader>
        <TableRow className="bg-muted hover:bg-muted"> {/* Removed border/rounded here, apply on wrapper */}
          <TableHead>Hash</TableHead>
          <TableHead>Type</TableHead>
          <TableHead>From</TableHead>
          <TableHead>To</TableHead>
          <TableHead>Token</TableHead>
          <TableHead className="text-right">Value</TableHead>
          <TableHead>Age</TableHead>
          <TableHead>Status</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {transactions.length > 0 ? (
          transactions.map((tx) => (
            <TransactionRow
              key={tx.hash}
              tx={tx}
              walletAddress={walletAddress}
              explorerBaseUrl={explorerBaseUrl}
            />
          ))
        ) : (
          <TableRow>
            <TableCell colSpan={8} className="text-center h-24"> {/* Added height for empty state */}
              No transactions found for this token.
            </TableCell>
          </TableRow>
        )}
      </TableBody>
    </Table>
  );
} 