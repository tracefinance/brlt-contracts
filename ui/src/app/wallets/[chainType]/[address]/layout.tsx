export default function WalletDetailsLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div className="container py-6">
      {children}
    </div>
  );
} 