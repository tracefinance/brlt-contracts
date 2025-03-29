import { Link, Outlet, useLoaderData } from "@remix-run/react";
import { SidebarInset, SidebarProvider, SidebarTrigger } from "~/components/ui/sidebar";
import { listWallets } from "../server/api/wallet.server";
import WalletSidebar from "~/components/wallet-sidebar";
import { listTokens } from "~/server/api/token.server";
import { useState } from "react";
import { Breadcrumb, BreadcrumbItem, BreadcrumbList, BreadcrumbSeparator } from "~/components/ui/breadcrumb";

export async function loader() {
  const wallets = await listWallets(10, 0, "");
  const defaultWallet = wallets.items[0];
  const tokens = await listTokens("", defaultWallet.chainType, "erc20", 10, 0);

  return {
    wallets: wallets.items,
    defaultWallet: defaultWallet,
    tokens: tokens.items,
  };
}

export default function Wallets() {
  const { wallets, defaultWallet: defaultWallet, tokens } = useLoaderData<typeof loader>();
  
  const [selectedWallet, setSelectedWallet] = useState({ 
    name: defaultWallet.name,
    address: defaultWallet.address,
    chainType: defaultWallet.chainType,
  });

  return (
    <SidebarProvider>
      <WalletSidebar
        wallets={wallets}
        selectedWallet={selectedWallet}
        tokens={tokens}
        onWalletChange={setSelectedWallet}
      />
      <SidebarInset className="mt-16">
        <header className="flex h-16 shrink-0 items-center gap-2 border-b px-4">
            <SidebarTrigger className="size-12 -ml-2" />
            <Breadcrumb>
            <BreadcrumbList>
                <BreadcrumbItem>
                    <Link to="/wallets">{selectedWallet.name}</Link>
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
          <Outlet />  
        </div>        
      </SidebarInset>
    </SidebarProvider>
  );
}