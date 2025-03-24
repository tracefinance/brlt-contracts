'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { LucideIcon } from 'lucide-react';
import { ReactNode } from 'react';

interface NavLinkProps {
  href: string;
  icon?: LucideIcon;
  children: ReactNode;
  className?: string;
}

export function NavLink({ href, icon: Icon, children, className = '' }: NavLinkProps) {
  const pathname = usePathname();
  const isActive = pathname?.startsWith(href);
  
  return (
    <Link 
      href={href} 
      className={`flex items-center gap-2 transition-colors hover:text-foreground/80 ${
        isActive 
          ? 'text-foreground font-semibold' 
          : 'text-foreground/60'
      } ${className}`}
    >
      {Icon && <Icon strokeWidth={1.5} />}
      {children}
    </Link>
  );
} 