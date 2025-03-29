import { Logo } from '~/components/ui/logo';
import { NavLink } from '~/components/ui/nav-link';
import { Vault, Wallet, Settings, Repeat, Shuffle } from 'lucide-react';

export function Header() {
  return (
    <header className="fixed top-0 left-0 right-0 z-50 w-full border-b bg-background h-16">
      <div className="flex w-full h-full items-center gap-4 px-4 py-2">
        <Logo className="h-10 w-auto"/>

        <nav className="flex items-center space-x-6 text-sm font-medium ml-4">
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