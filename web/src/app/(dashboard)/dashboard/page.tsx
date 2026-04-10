"use client";

import { useEffect, useState, useCallback } from "react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Progress } from "@/components/ui/progress";
import { api, ApiError } from "@/lib/api";
import type { DashboardSummary, CurrentCardCredits, CreditPeriod } from "@/types/api";

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
  const daysLeft = Math.ceil(
    (end.getTime() - now.getTime()) / (1000 * 60 * 60 * 24),
  );
  if (daysLeft < 0) return { variant: "danger", label: "Expired" };
  if (daysLeft <= 7) return { variant: "warning", label: "Expiring" };
  return { variant: "default", label: "Upcoming" };
}

export default function DashboardPage() {
  const [dashboard, setDashboard] = useState<DashboardSummary | null>(null);
  const [cardCredits, setCardCredits] = useState<CurrentCardCredits[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const fetchData = useCallback(async () => {
    try {
      const [dashRes, creditsRes] = await Promise.all([
        api.get<DashboardSummary>("/api/v1/me/dashboard/summary"),
        api.get<{ data: CurrentCardCredits[] }>("/api/v1/me/credits/current"),
      ]);
      setDashboard(dashRes);
      setCardCredits(creditsRes.data);
    } catch (err) {
      setError(
        err instanceof ApiError
          ? `Error: ${err.status}`
          : "Failed to load dashboard",
      );
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  if (loading) {
    return (
      <div className="flex items-center justify-center py-12">
        <div className="h-8 w-8 animate-spin rounded-full border-2 border-primary border-t-transparent" />
      </div>
    );
  }

  if (error) {
    return (
      <>
        <h1 className="text-2xl font-bold text-foreground">Dashboard</h1>
        <div className="mt-4 rounded-lg bg-destructive/15 px-4 py-3 text-sm text-destructive">
          {error}
        </div>
      </>
    );
  }

  if (!dashboard) return null;

  const d = dashboard;

  // Flatten all credit periods for the "Recent Credits" section
  const recentCredits = cardCredits.flatMap((cc) =>
    cc.credits.map((cp) => ({
      ...cp,
      card_name: cc.user_card.card.name,
    })),
  );

  return (
    <>
      <h1 className="text-2xl font-bold text-foreground">Dashboard</h1>

      <div className="mt-6 grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
        <Card>
          <CardHeader>
            <CardTitle>Active Cards</CardTitle>
            <CardDescription>Cards being tracked</CardDescription>
          </CardHeader>
          <CardContent>
            <p className="text-3xl font-bold text-foreground">
              {d.cards.length}
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Credits Used</CardTitle>
            <CardDescription>
              {formatDollars(d.total_used_cents)} of{" "}
              {formatDollars(d.total_available_cents)}
            </CardDescription>
          </CardHeader>
          <CardContent>
            <p className="text-3xl font-bold text-primary">
              {d.credits_used}/{d.credits_total}
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Unused Value</CardTitle>
            <CardDescription>
              Don&apos;t leave money on the table
            </CardDescription>
          </CardHeader>
          <CardContent>
            <p className="text-3xl font-bold text-warning">
              {formatDollars(d.total_unused_cents)}
            </p>
          </CardContent>
        </Card>
      </div>

      <h2 className="mt-10 text-lg font-semibold text-foreground">Cards</h2>
      <div className="mt-4 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {d.cards.map((card) => (
          <Card key={card.user_card_id}>
            <CardHeader>
              <CardTitle>{card.card_name}</CardTitle>
              <CardDescription>{card.issuer}</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="flex items-baseline justify-between">
                <span className="text-sm text-muted-foreground">
                  {card.credits_used}/{card.credits_total} credits
                </span>
                <span className="text-sm font-medium text-foreground">
                  {formatDollars(card.used_cents)}/
                  {formatDollars(card.available_cents)}
                </span>
              </div>
              <Progress
                value={
                  card.available_cents
                    ? (card.used_cents / card.available_cents) * 100
                    : 0
                }
                className="mt-2"
              />
            </CardContent>
          </Card>
        ))}
      </div>

      <h2 className="mt-10 text-lg font-semibold text-foreground">
        Recent Credits
      </h2>
      <Card className="mt-4">
        <CardContent>
          {recentCredits.length === 0 ? (
            <p className="py-4 text-center text-sm text-muted-foreground">
              No credits to show yet.
            </p>
          ) : (
            <div className="divide-y divide-border">
              {recentCredits.map((credit) => {
                const status = creditStatus(credit);
                return (
                  <div
                    key={credit.id}
                    className="flex items-center justify-between py-3"
                  >
                    <div>
                      <p className="text-sm font-medium text-foreground">
                        {credit.credit_definition.name}
                      </p>
                      <p className="text-xs text-muted-foreground">
                        {credit.card_name}
                      </p>
                    </div>
                    <div className="flex items-center gap-3">
                      <span className="text-sm font-medium text-foreground">
                        {formatDollars(credit.credit_definition.amount_cents)}
                      </span>
                      <Badge variant={status.variant}>{status.label}</Badge>
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </CardContent>
      </Card>
    </>
  );
}
