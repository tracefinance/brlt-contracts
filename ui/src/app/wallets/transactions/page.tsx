import { Metadata } from "next";

export const metadata: Metadata = {
  title: "Transactions | Vault0",
  description: "View your wallet transactions",
};

export default function TransactionsPage() {
  return (
    <div>
      <h1 className="text-3xl font-bold mb-6">Transactions</h1>
      <p className="mb-6">View and manage your wallet transactions.</p>
      
      <div className="rounded-lg border shadow-sm">
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b bg-muted/50">
                <th className="px-4 py-3 text-left font-medium">Date</th>
                <th className="px-4 py-3 text-left font-medium">Type</th>
                <th className="px-4 py-3 text-left font-medium">Amount</th>
                <th className="px-4 py-3 text-left font-medium">Wallet</th>
                <th className="px-4 py-3 text-left font-medium">Status</th>
              </tr>
            </thead>
            <tbody>
              <tr className="border-b">
                <td className="px-4 py-3">Mar 24, 2023</td>
                <td className="px-4 py-3">Transfer</td>
                <td className="px-4 py-3">0.5 ETH</td>
                <td className="px-4 py-3">My ETH Wallet</td>
                <td className="px-4 py-3">
                  <span className="inline-flex items-center rounded-full bg-green-100 px-2.5 py-0.5 text-xs font-medium text-green-800">
                    Completed
                  </span>
                </td>
              </tr>
              <tr className="border-b">
                <td className="px-4 py-3">Mar 22, 2023</td>
                <td className="px-4 py-3">Receive</td>
                <td className="px-4 py-3">1.2 ETH</td>
                <td className="px-4 py-3">My ETH Wallet</td>
                <td className="px-4 py-3">
                  <span className="inline-flex items-center rounded-full bg-green-100 px-2.5 py-0.5 text-xs font-medium text-green-800">
                    Completed
                  </span>
                </td>
              </tr>
              <tr className="border-b">
                <td className="px-4 py-3">Mar 20, 2023</td>
                <td className="px-4 py-3">Transfer</td>
                <td className="px-4 py-3">0.1 BTC</td>
                <td className="px-4 py-3">Bitcoin Wallet</td>
                <td className="px-4 py-3">
                  <span className="inline-flex items-center rounded-full bg-yellow-100 px-2.5 py-0.5 text-xs font-medium text-yellow-800">
                    Pending
                  </span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
} 