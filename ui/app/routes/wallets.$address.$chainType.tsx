import { Link, Outlet, useLoaderData, useNavigate, useParams } from "@remix-run/react";
import { useEffect } from "react";
import { SidebarInset, SidebarProvider, SidebarTrigger } from "~/components/ui/sidebar";
import WalletSidebar from "~/components/wallet-sidebar";
import { Breadcrumb, BreadcrumbItem, BreadcrumbList, BreadcrumbSeparator } from "~/components/ui/breadcrumb";
import { WalletClient } from "~/server/api";
import { LoaderFunctionArgs, json, redirect } from "@remix-run/node";
import { Wallet, TokenBalanceResponse } from "~/models/wallet";
import { ZERO_ADDRESS } from "~/lib/constants";

// Define the ID for this route loader data, used by child routes
export const walletDetailsRouteId = "routes/wallets.$address.$chainType";

// Define the structure of the data loaded by this route
export type WalletDetailsLoaderData = {
  wallets: Wallet[];
  currentWallet: Wallet;
  balances: TokenBalanceResponse[];
};

export async function loader({ params }: LoaderFunctionArgs) {
  const address = params.address;
  const chainType = params.chainType;
  
  // Ensure address and chainType are defined
  if (!address || !chainType) {
    return redirect("/wallets");
  }
  
  // TODO: Replace hardcoded token with actual session token
  const token = "123";
  if (!token) {
      // This redirect should go to login, not throw an error that the layout tries to handle
      return redirect("/login"); 
  }

  const walletClient = new WalletClient(token);

  try {
    // Fetch data needed for the layout/sidebar
    const [walletsResponse, currentWallet, balances] = await Promise.all([
      walletClient.listWallets(100, 0),
      walletClient.getWallet(chainType, address),
      walletClient.getWalletBalance(chainType, address),
    ]);

    // Use json helper
    return json<WalletDetailsLoaderData>({ 
        wallets: walletsResponse.items,
        currentWallet,
        balances,
    });

  } catch (error) {
    console.error("Error loading wallet layout data:", error);
    if (error instanceof Response && error.status === 404) {
        // Wallet not found, redirect to base wallets page
        return redirect("/wallets?error=notfound"); 
    }
    // Throw a generic error response
    throw new Response("Error loading wallet data", { status: 500 }); 
  }
}

export default function WalletDetailsLayout() {
  const loaderData = useLoaderData<WalletDetailsLoaderData>();
  const navigate = useNavigate();
  const params = useParams();

  // Client-side redirect effect
  useEffect(() => {
    // Redirect if we are on the base wallet route (no tokenAddress param)
    if (params.address && params.chainType && !params.tokenAddress) {
      navigate(`/wallets/${params.address}/${params.chainType}/transactions/${ZERO_ADDRESS}`, { 
          replace: true 
      });
    }
  }, [params.address, params.chainType, params.tokenAddress, navigate]);

  // Defensive check is still good practice
  if (!loaderData || !loaderData.currentWallet) { 
      return (
          <SidebarProvider>
              <SidebarInset className="mt-16">
                  <div className="p-4">Loading...</div>
              </SidebarInset>
          </SidebarProvider>
      );
  }

  const { wallets, currentWallet, balances } = loaderData;

  const handleWalletChange = (wallet: Wallet) => {
    navigate(`/wallets/${wallet.address}/${wallet.chainType}`);
  };

  return (
    <SidebarProvider>
      <WalletSidebar
        wallets={wallets}
        selectedWallet={currentWallet}
        balances={balances}
        onWalletChange={handleWalletChange}
        activeTokenAddress={params.tokenAddress}
      />
      <SidebarInset className="mt-16">
        <header className="flex h-16 shrink-0 items-center gap-2 border-b px-4">
           <SidebarTrigger className="size-12 -ml-2" />
           {/* TODO: Breadcrumb logic needs update */} 
           <Breadcrumb>
            <BreadcrumbList>
              <BreadcrumbItem>
                {currentWallet.name || 'Wallet'}
              </BreadcrumbItem>
              <BreadcrumbSeparator/>
              <BreadcrumbItem>Transactions</BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
        </header>
        <div className="flex flex-1 flex-col gap-4 p-4">
          {/* Render Outlet for child routes (transactions) */}
          <Outlet />
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}
