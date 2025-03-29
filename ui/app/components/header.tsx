import { Logo } from '~/components/ui/logo';
import { NavLink } from '~/components/ui/nav-link';
import { Vault, Wallet, Settings, Repeat, Shuffle } from 'lucide-react';

export function Header() {
  return (
    <header className="flex sticky top-0 z-50 w-full items-center border-b bg-background">
      <div className="flex h-[--header-height] w-full items-center gap-4 px-4 py-2">
        <Logo className="h-10 w-auto"/>

        <nav className="flex items-center space-x-6 text-sm font-medium">
          <NavLink href="/wallets" icon={Wallet}>
            Wallets
          </NavLink>

          <NavLink href="/vaults" icon={Vault}>
            Vaults
          </NavLink>
          
          <NavLink href="/swaps" icon={Repeat}>
            Swap
          </NavLink>

          <NavLink href="/bridges" icon={Shuffle}>
            Bridge
          </NavLink>

          <NavLink href="/settings" icon={Settings}>
            Settings
          </NavLink>
        </nav>
      </div>
    </header>
  );
} 