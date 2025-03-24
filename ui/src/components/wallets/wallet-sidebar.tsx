"use client";

import React, { useState } from "react";
import {
  Sidebar,
  SidebarContent,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuItem,
  SidebarMenuButton,
  SidebarSeparator,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupLabel,
  SidebarGroupContent,
} from "@/components/ui/sidebar";
import { WalletIcon, PlusIcon, CreditCardIcon, SettingsIcon } from "lucide-react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { WalletSelector } from "@/components/wallets/wallet-selector";

type Wallet = {
  id: string;
  name: string;
};

interface WalletSidebarProps {
  className?: string;
}

export function WalletSidebar({ className }: WalletSidebarProps) {
  const pathname = usePathname();
  const [currentWallet, setCurrentWallet] = useState<Wallet>({ id: "1", name: "Wallet1" });

  const handleWalletChange = (wallet: Wallet) => {
    setCurrentWallet(wallet);
    // Additional logic here if needed (e.g., API calls, navigation)
  };

  return (
    <Sidebar className={className}>
      <SidebarHeader>
        <WalletSelector 
          currentWallet={currentWallet} 
          onWalletChange={handleWalletChange} 
        />
      </SidebarHeader>
      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupLabel>Manage</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              <SidebarMenuItem>
                <SidebarMenuButton
                  asChild
                  isActive={pathname === "/wallets"}
                  tooltip="All Wallets"
                >
                  <Link href="/wallets">
                    <WalletIcon className="mr-2" />
                    <span>All Wallets</span>
                  </Link>
                </SidebarMenuButton>
              </SidebarMenuItem>
              
              <SidebarMenuItem>
                <SidebarMenuButton
                  asChild
                  isActive={pathname === "/wallets/new"}
                  tooltip="Add Wallet"
                >
                  <Link href="/wallets/new">
                    <PlusIcon className="mr-2" />
                    <span>Add Wallet</span>
                  </Link>
                </SidebarMenuButton>
              </SidebarMenuItem>
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
        
        <SidebarSeparator />
        
        <SidebarGroup>
          <SidebarGroupLabel>Finance</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              <SidebarMenuItem>
                <SidebarMenuButton
                  asChild
                  isActive={pathname === "/wallets/transactions"}
                  tooltip="Transactions"
                >
                  <Link href="/wallets/transactions">
                    <CreditCardIcon className="mr-2" />
                    <span>Transactions</span>
                  </Link>
                </SidebarMenuButton>
              </SidebarMenuItem>
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>
      
      <SidebarFooter>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton
              asChild
              isActive={pathname === "/wallets/settings"}
              tooltip="Settings"
            >
              <Link href="/wallets/settings">
                <SettingsIcon className="mr-2" />
                <span>Settings</span>
              </Link>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarFooter>
    </Sidebar>
  );
} 