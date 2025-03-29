import { Sidebar, SidebarContent, SidebarGroup, SidebarGroupLabel, SidebarHeader, SidebarMenu, SidebarMenuButton, SidebarMenuItem } from "~/components/ui/sidebar";
import { WalletSelector } from "./wallet-selector";
import { Token, Wallet } from "./types";
import { Link } from "@remix-run/react";
import { TokenIcon } from "@web3icons/react";

interface WalletSidebarProps {
    wallets: Wallet[];
    selectedWallet: Wallet;
    onWalletChange: (wallet: Wallet) => void;
    tokens: Token[];
}

export default function WalletSidebar({ 
    wallets,
    selectedWallet,
    onWalletChange,
    tokens,
    ...props 
}: WalletSidebarProps & React.ComponentProps<typeof Sidebar>) {
    return (
        <Sidebar {...props}>
            <SidebarHeader className="mt-16">
                <WalletSelector 
                    wallets={wallets} 
                    selectedWallet={selectedWallet} 
                    onWalletChange={onWalletChange}
                />
            </SidebarHeader>
            <SidebarContent>
                <SidebarGroup>
                    <SidebarGroupLabel>Tokens</SidebarGroupLabel>
                    <SidebarMenu>
                        {tokens.map((token) => (
                            <SidebarMenuItem key={token.address}>
                                <SidebarMenuButton asChild>
                                    <Link to={`/wallets/${selectedWallet.address}/${selectedWallet.chainType}/transactions/${token.address}`}>
                                        <TokenIcon symbol={token.symbol.startsWith('W') ? token.symbol.slice(1) : token.symbol} />
                                        <span>{token.symbol}</span>
                                    </Link>
                                </SidebarMenuButton>
                            </SidebarMenuItem>
                        ))}
                    </SidebarMenu>
                </SidebarGroup>
            </SidebarContent>
        </Sidebar>
    )
}
