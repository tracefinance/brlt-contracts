"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { cn } from "@/lib/utils";
import { Wallet } from "lucide-react";

interface SidebarItem {
  name: string;
  href: string;
  icon: React.ElementType;
}

const sidebarItems: SidebarItem[] = [
  {
    name: "Wallets",
    href: "/wallets",
    icon: Wallet,
  },
];

export default function Sidebar() {
  const pathname = usePathname();

  return (
    <div className="w-64 h-full bg-card border-r flex flex-col">
      <div className="p-6 border-b">
        <h1 className="text-2xl font-bold">Vault0</h1>
      </div>
      <nav className="flex-1 p-4 space-y-2">
        {sidebarItems.map((item) => (
          <Link
            key={item.name}
            href={item.href}
            className={cn(
              "flex items-center px-3 py-2 rounded-md text-sm font-medium transition-colors",
              pathname === item.href
                ? "bg-primary/10 text-primary"
                : "text-muted-foreground hover:bg-accent hover:text-accent-foreground"
            )}
          >
            <item.icon className="h-5 w-5 mr-3" />
            {item.name}
          </Link>
        ))}
      </nav>
      <div className="p-4 border-t mt-auto">
        <div className="text-xs text-muted-foreground">
          &copy; {new Date().getFullYear()} Vault0
        </div>
      </div>
    </div>
  );
} 