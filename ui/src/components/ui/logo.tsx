"use client";

import Image from "next/image";
import { useTheme } from "@/providers/ThemeProvider";
import { useEffect, useState } from "react";

interface LogoProps {
  width?: number;
  height?: number;
  className?: string;
}

export function Logo({ 
  width = 150, 
  height = 50, 
  className = "h-14 w-auto" 
}: LogoProps) {
  const { theme } = useTheme();
  const [logoSrc, setLogoSrc] = useState("/logo-black.svg");
  
  useEffect(() => {
    // For explicit light/dark themes
    if (theme === "light") {
      setLogoSrc("/logo-black.svg");
      return;
    }
    
    if (theme === "dark") {
      setLogoSrc("/logo-white.svg");
      return;
    }
    
    // For system theme, check OS preference
    if (theme === "system") {
      const isDarkMode = window.matchMedia("(prefers-color-scheme: dark)").matches;
      setLogoSrc(isDarkMode ? "/logo-white.svg" : "/logo-black.svg");
      
      // Listen for changes in OS theme preference
      const mediaQuery = window.matchMedia("(prefers-color-scheme: dark)");
      const handleChange = (e: MediaQueryListEvent) => {
        setLogoSrc(e.matches ? "/logo-white.svg" : "/logo-black.svg");
      };
      
      mediaQuery.addEventListener("change", handleChange);
      return () => mediaQuery.removeEventListener("change", handleChange);
    }
  }, [theme]);

  return (
    <Image 
      src={logoSrc} 
      alt="Vault0 Logo" 
      width={width} 
      height={height} 
      priority
      className={className}
    />
  );
} 