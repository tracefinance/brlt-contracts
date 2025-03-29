import { useContext } from "react";
import { ThemeContext, type Theme } from "~/providers/theme";

interface ThemeContextType {
  theme: Theme;
  setTheme: (theme: Theme) => void;
}

/**
 * Hook for accessing and manipulating the current theme
 * @returns Object containing the current theme and a function to change it
 */
export function useTheme(): ThemeContextType {
  const context = useContext(ThemeContext);
  
  if (context === undefined) {
    throw new Error("useTheme must be used within a ThemeProvider");
  }
  
  return context;
} 