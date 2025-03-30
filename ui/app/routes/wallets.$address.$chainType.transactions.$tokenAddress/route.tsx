import { type LoaderFunctionArgs } from "@remix-run/node";
import { useLoaderData, useParams } from "@remix-run/react";
import { ZERO_ADDRESS } from "~/lib/constants";
import type { PagedTransactions } from "~/models/transaction";
import { TransactionClient } from "~/server/api";
import { TransactionTable } from "./transaction-table";
import { PageControls } from "~/components/page-controls";
import { PageSizeSelect } from "~/components/page-size-select";

// Define the zero address constant
const DEFAULT_LIMIT = 10; // Define default limit

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

// Function to get explorer URL based on chainType (placeholder)
// TODO: Implement proper logic to determine explorer URL based on chainType
const getExplorerBaseUrl = (chainType?: string) => {
    // Simple example, replace with actual logic
    if (chainType?.toLowerCase() === 'ethereum') {
        return "https://etherscan.io";
    }
    // Add other chains as needed (e.g., polygon, arbitrum)
    // if (chainType?.toLowerCase() === 'polygon') {
    //     return "https://polygonscan.com";
    // }
    return "https://etherscan.io"; // Default fallback
};

export default function TokenTransactionsRoute() {
  // Get the full data including pagination info
  const { transactions, offset, limit } = useLoaderData<LoaderData>();
  const params = useParams();
  const walletAddress = params.address;
  const chainType = params.chainType;

  // Determine the explorer URL
  const explorerBaseUrl = getExplorerBaseUrl(chainType);

  return (
    <div className="border rounded-lg overflow-hidden">
      <TransactionTable 
        transactions={transactions.items}
        walletAddress={walletAddress}
        explorerBaseUrl={explorerBaseUrl} 
      />

      {/* Container for pagination and page size selector */}
      {(transactions.items.length > 0 || offset > 0) && (
        <div className="flex items-center space-x-4 p-2 border-t"> {/* Adjusted spacing */} 
          <PageSizeSelect 
            currentLimit={limit} 
          />
          <PageControls 
            offset={offset} 
            limit={limit} 
            hasMore={transactions.hasMore} 
          />
        </div>
      )}
    </div>
  );
}