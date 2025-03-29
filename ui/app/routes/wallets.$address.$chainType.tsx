import { Link, useLoaderData, useNavigate } from "@remix-run/react";
import { SidebarInset, SidebarProvider, SidebarTrigger } from "~/components/ui/sidebar";
import WalletSidebar from "~/components/wallet-sidebar";
import { Breadcrumb, BreadcrumbItem, BreadcrumbList, BreadcrumbSeparator } from "~/components/ui/breadcrumb";
import { WalletClient } from "~/server/api";
import { LoaderFunctionArgs, redirect } from "@remix-run/node";
import { Wallet, TokenBalance } from "~/components/types";

export async function loader({ params }: LoaderFunctionArgs) {
  const address = params.address;
  const chainType = params.chainType;
  
  // Ensure address and chainType are defined
  if (!address || !chainType) {
    return redirect("/wallets");
  }
  
  const client = new WalletClient("123");
  
  // Get all wallets for the sidebar
  const walletsResponse = await client.listWallets(100, 0);
  const wallets = walletsResponse.items;
  
  // Directly get the wallet using the API instead of finding it in the list
  const currentWallet = await client.getWallet(chainType, address);
  
  // Get balances for the current wallet
  const balances = await client.getWalletBalance(chainType, address);
  
  return {
    wallets,
    currentWallet,
    balances,
  };
}

type LoaderData = {
  wallets: Wallet[];
  currentWallet: Wallet;
  balances: TokenBalance[];
};

export default function WalletDetails() {
  const { wallets, currentWallet, balances } = useLoaderData<typeof loader>();
  const navigate = useNavigate();
  
  // Handle wallet change by redirecting to the selected wallet's URL
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
      />
      <SidebarInset className="mt-16">
        <header className="flex h-16 shrink-0 items-center gap-2 border-b px-4">
          <SidebarTrigger className="size-12 -ml-2" />
          <Breadcrumb>
            <BreadcrumbList>
              <BreadcrumbItem>
                <Link to={`/wallets/${currentWallet.address}/${currentWallet.chainType}`}>{currentWallet.name}</Link>
              </BreadcrumbItem>
              <BreadcrumbSeparator />
              <BreadcrumbItem>
                <span>DAI</span>
              </BreadcrumbItem>
              <BreadcrumbSeparator />
              <BreadcrumbItem>
                <span>Transactions</span>
              </BreadcrumbItem>                            
            </BreadcrumbList>
          </Breadcrumb>
        </header>
        <div className="flex flex-1 flex-col gap-4 p-4">
          <div>
            <h1 className="text-2xl font-bold">Wallet Details</h1>
            <div className="mt-4">
              <div className="text-sm text-muted-foreground">Address</div>
              <div className="font-mono">{currentWallet.address}</div>
            </div>
            <div className="mt-4">
              <div className="text-sm text-muted-foreground">Chain</div>
              <div>{currentWallet.chainType}</div>
            </div>
            {balances.length > 0 && (
              <div className="mt-6">
                <h2 className="text-xl font-semibold mb-2">Token Balances</h2>
                <div className="space-y-2">
                  {balances.map((balance, i) => (
                    <div key={i} className="flex justify-between p-2 bg-muted rounded-md">
                      <div>{balance.token.symbol}</div>
                      <div>{balance.balance}</div>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>
        </div>        
      </SidebarInset>
    </SidebarProvider>
  );
}
