import { Metadata } from "next";

export const metadata: Metadata = {
  title: "Add Wallet | Vault0",
  description: "Add a new wallet to your Vault0 account",
};

export default function AddWalletPage() {
  return (
    <div>
      <h1 className="text-3xl font-bold mb-6">Add Wallet</h1>
      <p className="mb-6">Use the form below to add a new wallet to your account.</p>
      
      <div className="bg-card rounded-lg border p-6 shadow-sm">
        <form className="space-y-4">
          <div className="space-y-2">
            <label htmlFor="name" className="text-sm font-medium">
              Wallet Name
            </label>
            <input
              id="name"
              type="text"
              className="w-full rounded-md border px-3 py-2"
              placeholder="My ETH Wallet"
            />
          </div>
          
          <div className="space-y-2">
            <label htmlFor="address" className="text-sm font-medium">
              Wallet Address
            </label>
            <input
              id="address"
              type="text"
              className="w-full rounded-md border px-3 py-2"
              placeholder="0x..."
            />
          </div>
          
          <div className="space-y-2">
            <label htmlFor="type" className="text-sm font-medium">
              Wallet Type
            </label>
            <select
              id="type"
              className="w-full rounded-md border px-3 py-2"
            >
              <option value="ethereum">Ethereum</option>
              <option value="bitcoin">Bitcoin</option>
              <option value="solana">Solana</option>
            </select>
          </div>
          
          <button
            type="submit"
            className="rounded-md bg-primary px-4 py-2 text-primary-foreground hover:bg-primary/90"
          >
            Add Wallet
          </button>
        </form>
      </div>
    </div>
  );
} 