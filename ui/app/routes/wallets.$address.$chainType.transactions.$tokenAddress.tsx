import { type LoaderFunctionArgs } from "@remix-run/node";
import { Link, useLoaderData, useParams } from "@remix-run/react";
import { getToken, requireUserId } from "~/server/session.server";
import { TransactionClient } from "~/server/api";
import type { PagedTransactions, Transaction } from "~/models/transaction";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "~/components/ui/table";
import { Card, CardContent, CardHeader, CardTitle } from "~/components/ui/card";
import { formatDistanceToNow } from 'date-fns';
import { shortenAddress, formatCurrency } from "~/lib/utils";
import { TokenIcon } from "~/components/token-icon";
import { HelpCircle, ArrowUpRight, ArrowDownLeft, CheckCircle, Loader, XCircle, ChevronLeft, ChevronRight } from 'lucide-react';
import { Badge } from "~/components/ui/badge";
import { Button } from "~/components/ui/button";

// Define the zero address constant
const ZERO_ADDRESS = "0x0000000000000000000000000000000000000000";
const DEFAULT_LIMIT = 20; // Define default limit

// LoaderData now expects the full PagedTransactions object and pagination params
type LoaderData = {
  transactions: PagedTransactions;
  tokenAddress: string;
  offset: number;
  limit: number;
};

export const loader = async ({ request, params }: LoaderFunctionArgs) => {
  //await requireUserId(request);
  //const token = await getToken(request);
  const token = "123"; // Placeholder
  const { address, chainType, tokenAddress } = params;

  if (!address || !chainType || !tokenAddress) {
    throw new Response("Missing required parameters", { status: 400 });
  }
  
  if (!token) {
      throw new Response("Unauthorized", { status: 401 });
  }

  // Get offset and limit from URL search params
  const url = new URL(request.url);
  const offsetParam = url.searchParams.get("offset");
  const limitParam = url.searchParams.get("limit");

  const offset = parseInt(offsetParam || "0", 10);
  const limit = parseInt(limitParam || String(DEFAULT_LIMIT), 10);

  // Validate offset and limit
  const safeOffset = Math.max(0, offset);
  const safeLimit = Math.max(1, limit); 

  try {
    const transactionClient = new TransactionClient(token);
    const apiTokenAddress = tokenAddress.toLowerCase() === ZERO_ADDRESS ? undefined : tokenAddress;

    // Use safeOffset and safeLimit in the API call
    const transactions = await transactionClient.getTransactionsByAddress(
      chainType,
      address,
      safeLimit,    // Pass parsed limit
      safeOffset,   // Pass parsed offset
      apiTokenAddress 
    );

    // Return the full pagination object and parameters
    return Response.json({ transactions, tokenAddress, offset: safeOffset, limit: safeLimit }); 
  } catch (error) {
    console.error("Error fetching transactions:", error);
    if (error instanceof Response) {
        throw error;
    }
    throw new Response("Error fetching transactions", { status: 500 });
  }
};

export default function TokenTransactionsRoute() {
  // Get the full data including pagination info
  const { transactions, offset, limit } = useLoaderData<LoaderData>();
  const params = useParams();
  const walletAddress = params.address;

  const renderTransactionRow = (tx: Transaction) => {
      const isOutbound = walletAddress ? tx.fromAddress.toLowerCase() === walletAddress.toLowerCase() : false;
      const timestamp = new Date(tx.timestamp * 1000);

      // Define base URL - needs to be dynamic based on chainType
      const explorerBaseUrl = "https://etherscan.io"; // Example

      // Determine the address to display/link for the token, using ZERO_ADDRESS for native
      const displayTokenAddress = tx.tokenAddress || ZERO_ADDRESS;

      return (
        <TableRow key={tx.hash}>
            <TableCell>
              <a href={`${explorerBaseUrl}/tx/${tx.hash}`} target="_blank" rel="noopener noreferrer" className="text-blue-600 hover:underline">
                {shortenAddress(tx.hash)}
              </a>
            </TableCell>
            <TableCell>
              <Badge variant="outline">
                {isOutbound ? <ArrowUpRight /> : <ArrowDownLeft />}
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
                        <TokenIcon symbol={tx.tokenSymbol} className="mr-2 h-5 w-5" />
                        {tx.tokenSymbol}
                    </>
                ) : (
                    <>
                        <HelpCircle className="mr-2 h-5 w-5 text-muted-foreground" />
                        <span className="text-muted-foreground">N/A</span>
                    </>
                )}
            </TableCell>
            <TableCell className="text-right">{formatCurrency(tx.value)}</TableCell>
            <TableCell title={timestamp.toLocaleString()}>{formatDistanceToNow(timestamp, { addSuffix: true })}</TableCell>
            <TableCell>
              <Badge 
                variant="outline" 
              >
                {(() => {
                  const status = tx.status?.toLowerCase();
                  if (status === 'success') return <CheckCircle className="text-green-600" />;
                  if (status === 'pending') return <Loader className="text-muted-foreground" />;
                  if (status === 'failed') return <XCircle className="text-destructive" />;
                  return <HelpCircle className="text-muted-foreground" />;
                })()}
                {tx.status || 'Unknown'}
              </Badge>
            </TableCell>
        </TableRow>
      );
  }

  return (
    // Add a wrapper div with border, rounded corners, and padding
    <div className="border rounded-lg overflow-hidden">
      <Table>
        <TableHeader>
          <TableRow className="bg-muted hover:bg-muted">
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
          {transactions.items.length > 0 ? (
            transactions.items.map(renderTransactionRow)
          ) : (
            <TableRow>
              <TableCell colSpan={8} className="text-center">
                No transactions found for this token.
              </TableCell>
            </TableRow>
          )}
        </TableBody>
      </Table>

      {(transactions.items.length > 0 || offset > 0) && (
        <div className="flex items-center justify-end m-2 space-x-2">
          <Link
            to={`?offset=${Math.max(0, offset - limit)}&limit=${limit}`}
            preventScrollReset
            aria-disabled={offset === 0}
            tabIndex={offset === 0 ? -1 : undefined}
            className={offset === 0 ? "pointer-events-none" : ""}
          >
            <Button variant="outline" size="icon" disabled={offset === 0}>
              <span className="sr-only">Previous page</span>
              <ChevronLeft className="h-4 w-4" />
            </Button>
          </Link>
          <Link
            to={`?offset=${offset + limit}&limit=${limit}`}
            preventScrollReset
            aria-disabled={!transactions.hasMore}
            tabIndex={!transactions.hasMore ? -1 : undefined}
            className={!transactions.hasMore ? "pointer-events-none" : ""}
          >
            <Button variant="outline" size="icon" disabled={!transactions.hasMore}>
              <span className="sr-only">Next page</span>
              <ChevronRight className="h-4 w-4" />
            </Button>
          </Link>
        </div>
      )}
    </div>
  );
} 