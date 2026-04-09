"use client";

import { useEffect, useState, useCallback } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Progress } from "@/components/ui/progress";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { api, ApiError } from "@/lib/api";
import type { CurrentCardCredits, CreditPeriod } from "@/types/api";

function formatDollars(cents: number): string {
  return `$${(cents / 100).toFixed(2)}`;
}

function creditStatus(cp: CreditPeriod): {
  variant: "success" | "warning" | "danger" | "default";
  label: string;
} {
  if (cp.used) return { variant: "success", label: "Used" };
  const end = new Date(cp.period_end);
  const now = new Date();
  const daysLeft = Math.ceil((end.getTime() - now.getTime()) / (1000 * 60 * 60 * 24));
  if (daysLeft < 0) return { variant: "danger", label: "Expired" };
  if (daysLeft <= 7) return { variant: "warning", label: `${daysLeft}d left` };
  return { variant: "default", label: "Unused" };
}

function formatDateRange(start: string, end: string): string {
  const s = new Date(start);
  const e = new Date(end);
  return `${s.toLocaleDateString("en-US", { month: "short", day: "numeric" })} – ${e.toLocaleDateString("en-US", { month: "short", day: "numeric" })}`;
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
        err instanceof ApiError ? `Error: ${err.status}` : "Failed to load credits",
      );
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchCredits();
  }, [fetchCredits]);

  const toggleCredit = async (creditPeriodId: string, currentlyUsed: boolean) => {
    setToggling((prev) => new Set(prev).add(creditPeriodId));
    try {
      const action = currentlyUsed ? "mark-unused" : "mark-used";
      const updated = await api.post<CreditPeriod>(
        `/api/v1/me/credits/${creditPeriodId}/${action}`,
      );
      setCardCredits((prev) =>
        prev.map((cc) => ({
          ...cc,
          credits: cc.credits.map((cp) =>
            cp.id === creditPeriodId ? updated : cp,
          ),
        })),
      );
    } catch {
      setError("Failed to update credit");
    } finally {
      setToggling((prev) => {
        const next = new Set(prev);
        next.delete(creditPeriodId);
        return next;
      });
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center py-12">
        <div className="h-8 w-8 animate-spin rounded-full border-2 border-primary border-t-transparent" />
      </div>
    );
  }

  const hasCredits = cardCredits.some((cc) => cc.credits.length > 0);

  return (
    <>
      <h1 className="text-2xl font-bold text-foreground">Credits</h1>
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
          <CardContent className="py-8 text-center">
            <p className="text-muted-foreground">
              No credits to track yet. Add a card first from the{" "}
              <a href="/cards" className="text-primary hover:underline">
                Cards page
              </a>.
            </p>
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
              {cardCredits.map((cc) => {
                const filtered = cc.credits.filter((cp) => {
                  if (tab === "used") return cp.used;
                  if (tab === "unused") return !cp.used;
                  return true;
                });
                if (filtered.length === 0) return null;

                const usedCount = cc.credits.filter((cp) => cp.used).length;
                const totalCount = cc.credits.length;
                const pct = totalCount > 0 ? (usedCount / totalCount) * 100 : 0;

                return (
                  <Card key={cc.user_card.id}>
                    <CardHeader className="pb-3">
                      <div className="flex items-center justify-between">
                        <div>
                          <CardTitle className="text-base">
                            {cc.user_card.card.name}
                          </CardTitle>
                          <p className="text-sm text-muted-foreground">
                            {usedCount}/{totalCount} credits used
                          </p>
                        </div>
                        <span className="text-sm font-medium text-primary">
                          {Math.round(pct)}%
                        </span>
                      </div>
                      <Progress value={pct} className="mt-2" />
                    </CardHeader>
                    <CardContent>
                      <div className="divide-y divide-border">
                        {filtered.map((cp) => {
                          const status = creditStatus(cp);
                          const isToggling = toggling.has(cp.id);
                          return (
                            <div
                              key={cp.id}
                              className="flex items-center justify-between py-3"
                            >
                              <div className="min-w-0 flex-1">
                                <p className="text-sm font-medium text-foreground">
                                  {cp.credit_definition.name}
                                </p>
                                <p className="text-xs text-muted-foreground">
                                  {formatDollars(cp.credit_definition.amount_cents)} ·{" "}
                                  {formatDateRange(cp.period_start, cp.period_end)}
                                </p>
                              </div>
                              <div className="flex items-center gap-3">
                                <Badge variant={status.variant}>{status.label}</Badge>
                                <Button
                                  size="sm"
                                  variant={cp.used ? "secondary" : "default"}
                                  disabled={isToggling}
                                  onClick={() => toggleCredit(cp.id, cp.used)}
                                >
                                  {isToggling
                                    ? "..."
                                    : cp.used
                                      ? "Undo"
                                      : "Mark Used"}
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
