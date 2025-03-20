"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { format } from "date-fns";
import { truncateMiddle } from "@/lib/utils";
import { Wallet } from "@/types/models/wallet.model";
import { NetworkIcon } from "@web3icons/react";

interface WalletDetailProps {
  wallet: Wallet;
}

export default function WalletDetail({ wallet }: WalletDetailProps) {
  // Function to format date
  const formatDate = (dateString: string) => {
    return format(new Date(dateString), 'MMM d, yyyy HH:mm:ss');
  };
  
  return (
    <Card>
      <CardHeader>
        <CardTitle>{wallet.name}</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <h3 className="text-sm font-medium text-muted-foreground">Chain Type</h3>
            <p className="flex items-center gap-1 capitalize">
              <NetworkIcon id={wallet.chainType.toLowerCase()} size={20} variant="branded" />
              <span>{wallet.chainType}</span>
            </p>
          </div>
          
          <div>
            <h3 className="text-sm font-medium text-muted-foreground">Key ID</h3>
            <p className="font-mono text-sm">{truncateMiddle(wallet.keyId, 10, 10)}</p>
          </div>
          
          <div className="col-span-1 md:col-span-2">
            <h3 className="text-sm font-medium text-muted-foreground">Address</h3>
            <p className="font-mono text-sm break-all">{wallet.address}</p>
          </div>
          
          {wallet.tags && Object.keys(wallet.tags).length > 0 && (
            <div className="col-span-1 md:col-span-2">
              <h3 className="text-sm font-medium text-muted-foreground">Tags</h3>
              <div className="flex flex-wrap gap-2 mt-1">
                {Object.entries(wallet.tags).map(([key, value]) => (
                  <Badge key={key} variant="secondary">
                    {key}: {value}
                  </Badge>
                ))}
              </div>
            </div>
          )}
          
          <div>
            <h3 className="text-sm font-medium text-muted-foreground">Created At</h3>
            <p>{formatDate(wallet.createdAt)}</p>
          </div>
          
          <div>
            <h3 className="text-sm font-medium text-muted-foreground">Updated At</h3>
            <p>{formatDate(wallet.updatedAt)}</p>
          </div>
        </div>
      </CardContent>
    </Card>
  );
} 