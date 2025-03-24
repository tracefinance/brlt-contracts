"use client";

import React from "react";
import { ChevronsUpDown } from "lucide-react";
import {
  SidebarMenu,
  SidebarMenuItem,
  SidebarMenuButton,
} from "@/components/ui/sidebar";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";

type Wallet = {
  id: string;
  name: string;
};

interface WalletSelectorProps {
  currentWallet: Wallet;
  wallets: Wallet[];
  onWalletChange: (wallet: Wallet) => void;
}

export function WalletSelector({
  currentWallet = { id: "1", name: "Wallet1" },
  wallets = [],
  onWalletChange = () => {},
}: Partial<WalletSelectorProps>) {
  // Use default wallets if none provided
  const allWallets = wallets.length > 0 
    ? wallets 
    : [
        { id: "1", name: "Wallet1" },
        { id: "2", name: "Wallet2" },
        { id: "3", name: "Wallet3" },
      ];

  return (
    <SidebarMenu>
      <SidebarMenuItem>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <SidebarMenuButton
              size="lg"
              className="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
            >
              <div className="flex aspect-square size-8 items-center justify-center rounded-lg bg-sidebar-primary text-sidebar-primary-foreground">
                {currentWallet.name.charAt(0)}
              </div>
              <div className="grid flex-1 text-left text-sm leading-tight">
                <span className="truncate font-semibold">
                  {currentWallet.name}
                </span>
              </div>
              <ChevronsUpDown className="ml-auto" />
            </SidebarMenuButton>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="start" className="w-[200px]">
            {allWallets.map((wallet) => (
              <DropdownMenuItem
                key={wallet.id}
                onClick={() => onWalletChange(wallet)}
              >
                {wallet.name}
              </DropdownMenuItem>
            ))}
          </DropdownMenuContent>
        </DropdownMenu>
      </SidebarMenuItem>
    </SidebarMenu>
  );
} 