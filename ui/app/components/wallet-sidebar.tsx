import { Sidebar, SidebarContent, SidebarGroup, SidebarGroupLabel, SidebarHeader, SidebarMenu, SidebarMenuButton, SidebarMenuItem } from "~/components/ui/sidebar";
import { WalletSelector } from "./wallet-selector";
import { Link } from "@remix-run/react";
import { TokenIcon } from "./token-icon";
import { formatCurrency } from "~/lib/utils";
import { Wallet, TokenBalanceResponse } from "~/models/wallet";
import { Token } from "~/models/token";
import { ZERO_ADDRESS } from "~/lib/constants";

interface WalletSidebarProps {
    wallets: Wallet[];
    selectedWallet: Wallet;
    onWalletChange: (wallet: Wallet) => void;
    balances: TokenBalanceResponse[];
    activeTokenAddress?: string;
}

export default function WalletSidebar({ 
    wallets,
    selectedWallet,
    onWalletChange,
    balances,
    activeTokenAddress,
    ...props 
}: WalletSidebarProps & React.ComponentProps<typeof Sidebar>) {
    const comparisonAddress = activeTokenAddress?.toLowerCase();

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
                        {balances.map((balance) => {
                            const tokenAddr = (balance.token?.address || ZERO_ADDRESS).toLowerCase();
                            const tokenSymbol = balance.token?.symbol || 'N/A';
                            
                            const isActive = comparisonAddress === tokenAddr;

                            const addressForUrlPath = tokenAddr;
                            
                            const targetUrl = `/wallets/${selectedWallet.address}/${selectedWallet.chainType}/transactions/${addressForUrlPath}`;

                            const itemKey = tokenAddr;

                            return (
                                <SidebarMenuItem key={itemKey}>
                                    <SidebarMenuButton isActive={isActive} asChild>
                                        <Link className="flex items-center w-full" to={targetUrl}>
                                            <TokenIcon symbol={tokenSymbol} />
                                            <span>{tokenSymbol}</span>
                                            <span className="ml-auto text-sm text-gray-500">
                                                {formatCurrency(balance.balance)}
                                            </span>
                                        </Link>
                                    </SidebarMenuButton>
                                </SidebarMenuItem>
                            );
                        })}
                    </SidebarMenu>
                </SidebarGroup>
            </SidebarContent>
        </Sidebar>
    )
}
