import { cn } from "@/lib/utils";

export function Dollars({
  cents,
  className,
  showCents = true,
}: {
  cents: number;
  className?: string;
  showCents?: boolean;
}) {
  const formatted = showCents
    ? `$${(cents / 100).toFixed(2)}`
    : `$${(cents / 100).toFixed(0)}`;

  return (
    <span className={cn("font-mono tabular-nums", className)}>{formatted}</span>
  );
}
