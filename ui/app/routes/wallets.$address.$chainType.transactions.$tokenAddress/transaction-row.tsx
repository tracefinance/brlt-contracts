import { formatDistanceToNow } from 'date-fns';
import { ArrowDownLeft, ArrowUpRight, CheckCircle, HelpCircle, Loader, XCircle } from 'lucide-react';
import { TokenIcon } from "~/components/token-icon";
import { Badge } from "~/components/ui/badge";
import { TableCell, TableRow } from "~/components/ui/table";
import { formatCurrency, shortenAddress } from "~/lib/utils";
import type { Transaction } from "~/models/transaction";

// --- Helper Component for Status Icon ---
interface TransactionStatusIconProps {
    status?: string | null;
}

function TransactionStatusIcon({ status }: TransactionStatusIconProps) {
    const lowerStatus = status?.toLowerCase();
    switch (lowerStatus) {
        case 'success':
            return <CheckCircle className="mr-1 h-4 w-4 text-green-600" />;
        case 'pending':
            return <Loader className="mr-1 h-4 w-4 animate-spin text-muted-foreground" />; // Added animation
        case 'failed':
            return <XCircle className="mr-1 h-4 w-4 text-destructive" />;
        default:
            return <HelpCircle className="mr-1 h-4 w-4 text-muted-foreground" />;
    }
}

// --- Main Transaction Row Component ---
interface TransactionRowProps {
  tx: Transaction;
  walletAddress?: string;
  explorerBaseUrl: string; // Pass explorer URL as prop
}

export function TransactionRow({ tx, walletAddress, explorerBaseUrl }: TransactionRowProps) {
    const isOutbound = walletAddress ? tx.fromAddress.toLowerCase() === walletAddress.toLowerCase() : false;
    const timestamp = new Date(tx.timestamp * 1000);

    return (
        <TableRow key={tx.hash}>
            <TableCell>
                <a href={`${explorerBaseUrl}/tx/${tx.hash}`} target="_blank" rel="noopener noreferrer" className="text-blue-600 hover:underline">
                    {shortenAddress(tx.hash, 6, 6)}
                </a>
            </TableCell>
            <TableCell>
                <Badge variant="outline" className="flex items-center"> {/* Ensure flex alignment */}
                    {isOutbound ? <ArrowUpRight className="mr-1 h-4 w-4" /> : <ArrowDownLeft className="mr-1 h-4 w-4" />}
                    {isOutbound ? 'Send' : 'Receive'}
                </Badge>
            </TableCell>
            <TableCell>
                <a href={`${explorerBaseUrl}/address/${tx.fromAddress}`} target="_blank" rel="noopener noreferrer" className="hover:underline">
                    {shortenAddress(tx.fromAddress)}
                </a>
            </TableCell>
            <TableCell>
                <a href={`${explorerBaseUrl}/address/${tx.toAddress}`} target="_blank" rel="noopener noreferrer" className="hover:underline">
                    {shortenAddress(tx.toAddress)}
                </a>
            </TableCell>
            <TableCell className="flex items-center">
                {tx.tokenSymbol ? (
                    <>
                        <TokenIcon symbol={tx.tokenSymbol} className="mr-2 h-5 w-5" variant="branded"/>
                        {tx.tokenSymbol}
                    </>
                ) : (
                    <>
                        {/* Use a placeholder or specific icon for non-token transfers if needed */}
                        <HelpCircle className="mr-2 h-5 w-5 text-muted-foreground" /> 
                        <span className="text-muted-foreground">N/A</span>
                    </>
                )}
            </TableCell>
            <TableCell className="text-right">{formatCurrency(tx.value)}</TableCell>
            <TableCell title={timestamp.toLocaleString()}>{formatDistanceToNow(timestamp, { addSuffix: true })}</TableCell>
            <TableCell>
                <Badge variant="outline" className="flex items-center"> {/* Ensure flex alignment */}
                    <TransactionStatusIcon status={tx.status} />
                    {tx.status || 'Unknown'}
                </Badge>
            </TableCell>
        </TableRow>
    );
} 