"use client";

import { useEffect, useState, useCallback } from "react";
import Link from "next/link";
import { Receipt, CreditCard, Check, Undo2 } from "lucide-react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Progress } from "@/components/ui/progress";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Dollars } from "@/components/dollars";
import { CreditsSkeleton } from "@/components/skeletons";
import { api, ApiError } from "@/lib/api";
import { toast } from "sonner";
import type { CurrentCardCredits, CreditPeriod } from "@/types/api";

function creditStatus(cp: CreditPeriod): {
  variant: "success" | "warning" | "danger" | "default";
  label: string;
  urgency: number; // lower = more urgent
} {
  if (cp.used) return { variant: "success", label: "Used", urgency: 100 };
  const end = new Date(cp.period_end);
  const now = new Date();
  const daysLeft = Math.ceil(
    (end.getTime() - now.getTime()) / (1000 * 60 * 60 * 24),
  );
  if (daysLeft < 0) return { variant: "danger", label: "Expired", urgency: 0 };
  if (daysLeft <= 3)
    return { variant: "danger", label: `${daysLeft}d left`, urgency: 1 };
  if (daysLeft <= 7)
    return { variant: "warning", label: `${daysLeft}d left`, urgency: 2 };
  return { variant: "default", label: "Unused", urgency: 50 };
}

function formatDateRange(start: string, end: string): string {
  const s = new Date(start);
  const e = new Date(end);
  return `${s.toLocaleDateString("en-US", { month: "short", day: "numeric" })} – ${e.toLocaleDateString("en-US", { month: "short", day: "numeric", year: "numeric" })}`;
}

function sortByUrgency(credits: CreditPeriod[]): CreditPeriod[] {
  return [...credits].sort((a, b) => {
    const aStatus = creditStatus(a);
    const bStatus = creditStatus(b);
    return aStatus.urgency - bStatus.urgency;
  });
}

export default function CreditsPage() {
  const [cardCredits, setCardCredits] = useState<CurrentCardCredits[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [toggling, setToggling] = useState<Set<string>>(new Set());

  const fetchCredits = useCallback(async () => {
    try {
      const res = await api.get<{ data: CurrentCardCredits[] }>(
        "/api/v1/me/credits/current",
      );
      setCardCredits(res.data);
    } catch (err) {
      setError(
        err instanceof ApiError
          ? `Error: ${err.status}`
          : "Failed to load credits",
      );
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchCredits();
  }, [fetchCredits]);

  const toggleCredit = async (
    creditPeriod: CreditPeriod,
    cardName: string,
  ) => {
    setToggling((prev) => new Set(prev).add(creditPeriod.id));
    try {
      const action = creditPeriod.used ? "mark-unused" : "mark-used";
      const updated = await api.post<CreditPeriod>(
        `/api/v1/me/credits/${creditPeriod.id}/${action}`,
      );
      setCardCredits((prev) =>
        prev.map((cc) => ({
          ...cc,
          credits: cc.credits.map((cp) =>
            cp.id === creditPeriod.id ? updated : cp,
          ),
        })),
      );
      if (updated.used) {
        toast.success("Credit marked as used", {
          description: `${creditPeriod.credit_definition.name} — ${cardName}`,
        });
      } else {
        toast("Credit marked as unused", {
          description: `${creditPeriod.credit_definition.name} — ${cardName}`,
        });
      }
    } catch {
      toast.error("Failed to update credit");
    } finally {
      setToggling((prev) => {
        const next = new Set(prev);
        next.delete(creditPeriod.id);
        return next;
      });
    }
  };

  if (loading) return <CreditsSkeleton />;

  const hasCredits = cardCredits.some((cc) => cc.credits.length > 0);

  // Sort card groups by most urgent first
  const sortedCardCredits = [...cardCredits].sort((a, b) => {
    const aMin = Math.min(
      ...a.credits.map((cp) => creditStatus(cp).urgency),
      99,
    );
    const bMin = Math.min(
      ...b.credits.map((cp) => creditStatus(cp).urgency),
      99,
    );
    return aMin - bMin;
  });

  return (
    <>
      <h1 className="text-2xl font-bold tracking-tight">Credits</h1>
      <p className="mt-1 text-sm text-muted-foreground">
        Track which card credits you&apos;ve used this period
      </p>

      {error && (
        <div className="mt-4 rounded-lg bg-destructive/15 px-4 py-3 text-sm text-destructive">
          {error}
        </div>
      )}

      {!hasCredits ? (
        <Card className="mt-6">
          <CardContent className="flex flex-col items-center py-12 text-center">
            <Receipt className="size-12 text-muted-foreground" />
            <h2 className="mt-4 text-lg font-semibold">No credits to track</h2>
            <p className="mt-1 max-w-sm text-sm text-muted-foreground">
              Add a card first to start tracking credits.
            </p>
            <Button render={<Link href="/cards" />} className="mt-6">
              <CreditCard className="mr-2 size-4" />
              Go to Cards
            </Button>
          </CardContent>
        </Card>
      ) : (
        <Tabs defaultValue="all" className="mt-6">
          <TabsList>
            <TabsTrigger value="all">All</TabsTrigger>
            <TabsTrigger value="unused">Unused</TabsTrigger>
            <TabsTrigger value="used">Used</TabsTrigger>
          </TabsList>

          {["all", "unused", "used"].map((tab) => (
            <TabsContent key={tab} value={tab} className="space-y-6">
              {sortedCardCredits.map((cc) => {
                const filtered = (
                  tab === "all"
                    ? cc.credits
                    : cc.credits.filter((cp) =>
                        tab === "used" ? cp.used : !cp.used,
                      )
                );
                const sorted = sortByUrgency(filtered);
                if (sorted.length === 0) return null;

                const usedCount = cc.credits.filter((cp) => cp.used).length;
                const totalCount = cc.credits.length;
                const pct =
                  totalCount > 0 ? (usedCount / totalCount) * 100 : 0;

                return (
                  <Card key={cc.user_card.id}>
                    <CardHeader className="pb-3">
                      <div className="flex items-center justify-between">
                        <div>
                          <CardTitle className="text-base">
                            {cc.user_card.card.name}
                          </CardTitle>
                          <CardDescription>
                            {usedCount}/{totalCount} credits used
                          </CardDescription>
                        </div>
                        <span className="font-mono text-sm font-medium text-primary tabular-nums">
                          {Math.round(pct)}%
                        </span>
                      </div>
                      <Progress value={pct} className="mt-2" />
                    </CardHeader>
                    <CardContent>
                      <div className="divide-y divide-border">
                        {sorted.map((cp) => {
                          const status = creditStatus(cp);
                          const isToggling = toggling.has(cp.id);
                          return (
                            <div
                              key={cp.id}
                              className="flex items-center justify-between gap-4 py-3"
                            >
                              <div className="min-w-0 flex-1">
                                <p className="text-sm font-medium">
                                  {cp.credit_definition.name}
                                </p>
                                <p className="text-xs text-muted-foreground">
                                  <Dollars
                                    cents={cp.credit_definition.amount_cents}
                                    className="text-xs"
                                  />{" "}
                                  ·{" "}
                                  {formatDateRange(
                                    cp.period_start,
                                    cp.period_end,
                                  )}
                                </p>
                              </div>
                              <div className="flex items-center gap-2">
                                <Badge variant={status.variant}>
                                  {status.label}
                                </Badge>
                                <Button
                                  size="sm"
                                  variant={cp.used ? "outline" : "default"}
                                  disabled={isToggling}
                                  onClick={() =>
                                    toggleCredit(cp, cc.user_card.card.name)
                                  }
                                  className="gap-1.5"
                                >
                                  {isToggling ? (
                                    <span className="size-3.5 animate-spin rounded-full border-2 border-current border-t-transparent" />
                                  ) : cp.used ? (
                                    <Undo2 className="size-3.5" />
                                  ) : (
                                    <Check className="size-3.5" />
                                  )}
                                  {cp.used ? "Undo" : "Mark Used"}
                                </Button>
                              </div>
                            </div>
                          );
                        })}
                      </div>
                    </CardContent>
                  </Card>
                );
              })}
            </TabsContent>
          ))}
        </Tabs>
      )}
    </>
  );
}
