import { Link } from "@remix-run/react";
import { ChevronLeft, ChevronRight } from "lucide-react";
import { Button } from "~/components/ui/button";

interface PageControlsProps { // Renamed props interface
  offset: number;
  limit: number;
  hasMore: boolean;
}

// Renamed component
export function PageControls({ offset, limit, hasMore }: PageControlsProps) {
  // Removed the showPagination check and the early return null

  // Removed outer div with margin (m-2) as spacing should be handled by the parent
  return (
    <div className="flex items-center space-x-2">
      <Link
        to={`?offset=${Math.max(0, offset - limit)}&limit=${limit}`}
        preventScrollReset
        aria-disabled={offset === 0}
        tabIndex={offset === 0 ? -1 : undefined}
        className={offset === 0 ? "pointer-events-none opacity-50" : ""} // Added opacity for disabled visual
      >
        <Button variant="outline" size="icon" disabled={offset === 0}>
          <span className="sr-only">Previous page</span>
          <ChevronLeft className="h-4 w-4" />
        </Button>
      </Link>
      <Link
        to={`?offset=${offset + limit}&limit=${limit}`}
        preventScrollReset
        aria-disabled={!hasMore}
        tabIndex={!hasMore ? -1 : undefined}
        className={!hasMore ? "pointer-events-none opacity-50" : ""} // Added opacity for disabled visual
      >
        <Button variant="outline" size="icon" disabled={!hasMore}>
          <span className="sr-only">Next page</span>
          <ChevronRight className="h-4 w-4" />
        </Button>
      </Link>
    </div>
  );
} 