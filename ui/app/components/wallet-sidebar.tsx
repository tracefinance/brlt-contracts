import { Sidebar, SidebarContent, SidebarGroup, SidebarGroupLabel, SidebarHeader, SidebarMenu, SidebarMenuButton, SidebarMenuItem } from "~/components/ui/sidebar";
import { WalletSelector } from "./wallet-selector";
import { Token, TokenBalance, Wallet } from "./types";
import { Link } from "@remix-run/react";
import { TokenIcon } from "./token-icon";
import { formatCurrency } from "~/lib/utils";

interface WalletSidebarProps {
    wallets: Wallet[];
    selectedWallet: Wallet;
    onWalletChange: (wallet: Wallet) => void;
    balances: TokenBalance[];
}

export default function WalletSidebar({ 
    wallets,
    selectedWallet,
    onWalletChange,
    balances,
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
                        {balances.map((balance) => (
                            <SidebarMenuItem key={balance.token.address}>
                                <SidebarMenuButton asChild>
                                    <Link className="flex items-center w-full" to={`/wallets/${selectedWallet.address}/${selectedWallet.chainType}/transactions/${balance.token.address}`}>
                                        <TokenIcon symbol={balance.token.symbol} />
                                        <span>{balance.token.symbol}</span>
                                        <span className="ml-auto text-sm text-gray-500">
                                            {formatCurrency(balance.balance)}
                                        </span>
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
