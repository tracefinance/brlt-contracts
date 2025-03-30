import { redirect } from "@remix-run/node";
import { WalletClient } from "~/server/api";

export async function loader() {
  const client = new WalletClient("123");
  const wallets = await client.listWallets(100, 0);
  
  // If wallets exist, redirect to the first wallet's details page with address and chainType
  if (wallets.items.length > 0) {
    const firstWallet = wallets.items[0];
    return redirect(`/wallets/${firstWallet.address}/${firstWallet.chainType}`);
  }
  
  // If no wallets, show empty state
  return { wallets: [] };
}

export default function WalletsIndex() {
    return (
        <div className="flex flex-col items-center justify-center h-full">
            <h1 className="text-xl font-semibold mb-4">No Wallets Found</h1>
            <p className="text-muted-foreground">Create a wallet to get started</p>
        </div>
    );
}