import DashboardLayout from "@/components/core/layout/DashboardLayout";
import { ReactNode } from "react";

export default function WalletsLayout({ children }: { children: ReactNode }) {
  return <DashboardLayout>{children}</DashboardLayout>;
} 