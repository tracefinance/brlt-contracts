import { Outlet, useLoaderData } from "@remix-run/react";
import { SidebarInset, SidebarProvider, SidebarTrigger } from "~/components/ui/sidebar";
import { listWallets } from "../server/api/wallet.server";
import WalletSidebar from "~/components/wallet-sidebar";
import { Separator } from "~/components/ui/separator";

export async function loader() {
  const wallets = await listWallets(10, 0, "");
  
  return {
    wallets: wallets.items,
    currentWallet: wallets.items[0]
  };
}

export default function Wallets() {
  const { wallets, currentWallet } = useLoaderData<typeof loader>();

  return (
    <SidebarProvider>
      <WalletSidebar
        wallets={wallets}
        currentWallet={currentWallet}
        onWalletChange={() => {}}
      />
      <SidebarInset>
        <header className="flex h-16 shrink-0 items-center gap-2 border-b px-4">
            <SidebarTrigger className="-ml-1" />
            <Separator orientation="vertical" className="mr-2 h-4" />
            <h2 className="text-lg font-semibold">Wallets</h2>
        </header>
        <div className="flex flex-1 flex-col gap-4 p-4">
          <Outlet />  
        </div>        
      </SidebarInset>
    </SidebarProvider>
  );
}