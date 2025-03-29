import { Sidebar, SidebarHeader } from "~/components/ui/sidebar";
import { WalletSelector } from "./wallet-selector";
import { Wallet } from "./types";

interface WalletSidebarProps {
    wallets: Wallet[];
    currentWallet: Wallet;
    onWalletChange: (wallet: Wallet) => void;
    isLoading?: boolean;
}

export default function WalletSidebar({ 
    wallets,
    currentWallet,
    onWalletChange,
    isLoading,
    ...props 
}: WalletSidebarProps & React.ComponentProps<typeof Sidebar>) {
    return (
        <Sidebar {...props}>
            <SidebarHeader className="mt-14">
                <WalletSelector 
                    wallets={wallets} 
                    currentWallet={currentWallet} 
                    onWalletChange={onWalletChange}
                />
            </SidebarHeader>
        </Sidebar>
    )
}
