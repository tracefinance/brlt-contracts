import { Metadata } from "next";

export const metadata: Metadata = {
  title: "Wallets | Vault0",
  description: "Manage your wallets",
};

export default function WalletsPage() {
  return (
    <div className="container py-10">
      <h1 className="text-3xl font-bold mb-6">Wallets</h1>
      <p>Your wallets will appear here.</p>
    </div>
  );
} 