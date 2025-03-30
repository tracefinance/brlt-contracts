import { useNavigate, useLocation } from "@remix-run/react";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "~/components/ui/select";

interface PageSizeSelectProps {
  currentLimit: number;
  options?: number[];
}

const DEFAULT_OPTIONS = [10, 20, 50, 100];

export function PageSizeSelect({
  currentLimit,
  options = DEFAULT_OPTIONS
}: PageSizeSelectProps) {
  const navigate = useNavigate();
  const location = useLocation();

  const handleValueChange = (newLimit: string) => {
    const params = new URLSearchParams(location.search);
    params.set('limit', newLimit);
    params.set('offset', '0');
    navigate(`${location.pathname}?${params.toString()}`, { preventScrollReset: true });
  };

  return (
    <div className="flex items-center space-x-2">
        <span className="text-sm text-muted-foreground">Show:</span>
        <Select
          value={String(currentLimit)}
          onValueChange={handleValueChange}
        >
          <SelectTrigger className="h-8 w-[90px]">
            <SelectValue placeholder="Limit" />
          </SelectTrigger>
          <SelectContent>
              {options.map((option) => (
                <SelectItem key={option} value={String(option)}>
                  {option}
                </SelectItem>
              ))}
          </SelectContent>
        </Select>
    </div>
  );
} 