"use client";

import React from "react";
import { WalletSidebar } from "@/components/wallets/wallet-sidebar";

export default function WalletsLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <div className="flex w-full h-full">
      <WalletSidebar className="pt-14" />
      
      <main className="flex-1 p-6 overflow-auto">
        {children}
      </main>
    </div>
  );
} 