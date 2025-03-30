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
    const comparisonAddress = activeTokenAddress === 'native' ? undefined : activeTokenAddress?.toLowerCase();

    const ZERO_ADDRESS = "0x0000000000000000000000000000000000000000";

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
                            const isNativeToken = !balance.token.address || balance.token.address === ZERO_ADDRESS;
                            
                            const tokenAddressForComparison = isNativeToken ? undefined : balance.token.address.toLowerCase();

                            const isActive = comparisonAddress === tokenAddressForComparison;

                            const addressForUrlPath = isNativeToken ? 'native' : balance.token.address;
                            
                            const targetUrl = `/wallets/${selectedWallet.address}/${selectedWallet.chainType}/transactions/${addressForUrlPath}`;

                            const itemKey = isNativeToken ? 'native' : balance.token.address;

                            return (
                                <SidebarMenuItem 
                                    key={itemKey} 
                                    className={isActive ? "bg-accent" : ""}
                                >
                                    <SidebarMenuButton asChild>
                                        <Link className="flex items-center w-full" to={targetUrl}>
                                            <TokenIcon symbol={balance.token.symbol} />
                                            <span>{balance.token.symbol}</span>
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
