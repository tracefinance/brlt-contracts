import React from "react";
import { ChevronsUpDown } from "lucide-react";
import {
  SidebarMenu,
  SidebarMenuItem,
  SidebarMenuButton,
  useSidebar,
} from "~/components/ui/sidebar";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "~/components/ui/dropdown-menu";
import { NetworkIcon } from "@web3icons/react";
import { Wallet } from "./types";

interface WalletSelectorProps {
  selectedWallet: Wallet;
  wallets: Wallet[];
  onWalletChange: (wallet: Wallet) => void;
}

export function WalletSelector({
  selectedWallet,
  wallets = [],
  onWalletChange = () => {}
}: Partial<WalletSelectorProps>) {
  const { isMobile } = useSidebar();
  
  return (
    <SidebarMenu>
      <SidebarMenuItem>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <div>
              <SidebarMenuButton
                size="lg"
                className="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
              >
                <div className="flex aspect-square size-8 items-center justify-center rounded-md bg-sidebar-primary text-sidebar-primary-foreground">
                  <NetworkIcon id={selectedWallet?.chainType || "ethereum"} size={24} variant="mono" />
                </div>
                <div className="grid flex-1 text-left text-sm leading-tight">
                  <span className="truncate font-semibold">
                    {selectedWallet?.name || "Wallet"}
                  </span>
                </div>
                <ChevronsUpDown className="ml-auto" />
              </SidebarMenuButton>
            </div>
          </DropdownMenuTrigger>
          <DropdownMenuContent className="w-[--radix-dropdown-menu-trigger-width] min-w-[200px]"
            align="start"
            side={isMobile ? "bottom" : "right"}
            sideOffset={4}>
            {wallets && wallets.length > 0 ? (
              wallets.map((wallet) => (
                <DropdownMenuItem key={wallet.address} onClick={() => onWalletChange(wallet)}>
                  <div className="flex size-6 items-center justify-center rounded-md border bg-background">
                    <NetworkIcon id={wallet.chainType} size={24} variant="mono" />
                  </div>
                  <span className="ml-2">{wallet.name}</span>
                </DropdownMenuItem>
              ))
            ) : (
              <div className="p-3 text-sm text-muted-foreground">No wallets found</div>
            )}
          </DropdownMenuContent>
        </DropdownMenu>
      </SidebarMenuItem>
    </SidebarMenu>
  );
} 