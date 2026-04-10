"use client";

import { useEffect, useState, useCallback } from "react";
import Link from "next/link";
import {
  CreditCard,
  CheckCircle2,
  AlertTriangle,
  TrendingUp,
  Clock,
  Plus,
} from "lucide-react";
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
import { Dollars } from "@/components/dollars";
import { DashboardSkeleton } from "@/components/skeletons";
import { MonthlyChart } from "@/components/monthly-chart";
import { api, ApiError } from "@/lib/api";
import type {
  DashboardSummary,
  CurrentCardCredits,
  CreditPeriod,
  MonthlyDashboard,
} from "@/types/api";

function creditStatus(cp: CreditPeriod): {
  variant: "success" | "warning" | "danger" | "default";
  label: string;
  daysLeft: number;
} {
  if (cp.used) return { variant: "success", label: "Used", daysLeft: -1 };
  const end = new Date(cp.period_end);
  const now = new Date();
  const daysLeft = Math.ceil(
    (end.getTime() - now.getTime()) / (1000 * 60 * 60 * 24),
  );
  if (daysLeft < 0) return { variant: "danger", label: "Expired", daysLeft };
  if (daysLeft <= 7)
    return { variant: "warning", label: `${daysLeft}d left`, daysLeft };
  return { variant: "default", label: "Upcoming", daysLeft };
}

export default function DashboardPage() {
  const [dashboard, setDashboard] = useState<DashboardSummary | null>(null);
  const [cardCredits, setCardCredits] = useState<CurrentCardCredits[]>([]);
  const [monthly, setMonthly] = useState<MonthlyDashboard[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const fetchData = useCallback(async () => {
    try {
      const [dashRes, creditsRes, monthlyRes] = await Promise.all([
        api.get<DashboardSummary>("/api/v1/me/dashboard/summary"),
        api.get<{ data: CurrentCardCredits[] }>("/api/v1/me/credits/current"),
        api
          .get<{ data: MonthlyDashboard[] }>("/api/v1/me/dashboard/monthly")
          .catch(() => ({ data: [] })),
      ]);
      setDashboard(dashRes);
      setCardCredits(creditsRes.data);
      setMonthly(monthlyRes.data);
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

  if (loading) return <DashboardSkeleton />;

  if (error) {
    return (
      <>
        <h1 className="text-2xl font-bold">Dashboard</h1>
        <div className="mt-4 rounded-lg bg-destructive/15 px-4 py-3 text-sm text-destructive">
          {error}
        </div>
      </>
    );
  }

  if (!dashboard) return null;

  const d = dashboard;

  // Flatten and find expiring credits
  const allCredits = cardCredits.flatMap((cc) =>
    cc.credits.map((cp) => ({
      ...cp,
      card_name: cc.user_card.card.name,
    })),
  );

  const expiringCredits = allCredits
    .filter((cp) => {
      if (cp.used) return false;
      const end = new Date(cp.period_end);
      const now = new Date();
      const daysLeft = Math.ceil(
        (end.getTime() - now.getTime()) / (1000 * 60 * 60 * 24),
      );
      return daysLeft >= 0 && daysLeft <= 7;
    })
    .sort((a, b) => {
      const aEnd = new Date(a.period_end).getTime();
      const bEnd = new Date(b.period_end).getTime();
      return aEnd - bEnd;
    });

  const isEmpty = d.cards.length === 0;

  return (
    <>
      <h1 className="text-2xl font-bold tracking-tight">Dashboard</h1>

      {isEmpty ? (
        <Card className="mt-6">
          <CardContent className="flex flex-col items-center py-12 text-center">
            <CreditCard className="size-12 text-muted-foreground" />
            <h2 className="mt-4 text-lg font-semibold">No cards yet</h2>
            <p className="mt-1 max-w-sm text-sm text-muted-foreground">
              Add a credit card to start tracking your rewards and never miss a
              credit again.
            </p>
            <Button render={<Link href="/cards" />} className="mt-6">
              <Plus className="mr-2 size-4" />
              Add Your First Card
            </Button>
          </CardContent>
        </Card>
      ) : (
        <>
          {/* KPI Cards */}
          <div className="mt-6 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
            <Card>
              <CardHeader className="flex flex-row items-center justify-between pb-2">
                <CardDescription>Active Cards</CardDescription>
                <CreditCard className="size-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <p className="text-2xl font-bold">{d.cards.length}</p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="flex flex-row items-center justify-between pb-2">
                <CardDescription>Credits Used</CardDescription>
                <CheckCircle2 className="size-4 text-primary" />
              </CardHeader>
              <CardContent>
                <p className="text-2xl font-bold text-primary">
                  {d.credits_used}
                  <span className="text-base font-normal text-muted-foreground">
                    /{d.credits_total}
                  </span>
                </p>
                <p className="mt-1 text-xs text-muted-foreground">
                  <Dollars cents={d.total_used_cents} className="text-xs" /> of{" "}
                  <Dollars cents={d.total_available_cents} className="text-xs" />
                </p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="flex flex-row items-center justify-between pb-2">
                <CardDescription>Unused Value</CardDescription>
                <AlertTriangle className="size-4 text-warning" />
              </CardHeader>
              <CardContent>
                <Dollars
                  cents={d.total_unused_cents}
                  className="text-2xl font-bold text-warning"
                />
                <p className="mt-1 text-xs text-muted-foreground">
                  Don&apos;t leave money on the table
                </p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="flex flex-row items-center justify-between pb-2">
                <CardDescription>Capture Rate</CardDescription>
                <TrendingUp className="size-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <p className="text-2xl font-bold">
                  {d.total_available_cents > 0
                    ? Math.round(
                        (d.total_used_cents / d.total_available_cents) * 100,
                      )
                    : 0}
                  %
                </p>
                <Progress
                  value={
                    d.total_available_cents > 0
                      ? (d.total_used_cents / d.total_available_cents) * 100
                      : 0
                  }
                  className="mt-2"
                />
              </CardContent>
            </Card>
          </div>

          {/* Expiring Soon Alert */}
          {expiringCredits.length > 0 && (
            <Card className="mt-6 border-warning/30 bg-warning/5">
              <CardHeader className="pb-3">
                <CardTitle className="flex items-center gap-2 text-warning">
                  <Clock className="size-4" />
                  Expiring Soon
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="divide-y divide-border">
                  {expiringCredits.map((credit) => {
                    const status = creditStatus(credit);
                    return (
                      <div
                        key={credit.id}
                        className="flex items-center justify-between py-2.5"
                      >
                        <div>
                          <p className="text-sm font-medium">
                            {credit.credit_definition.name}
                          </p>
                          <p className="text-xs text-muted-foreground">
                            {credit.card_name}
                          </p>
                        </div>
                        <div className="flex items-center gap-3">
                          <Dollars
                            cents={credit.credit_definition.amount_cents}
                            className="text-sm font-medium"
                          />
                          <Badge variant={status.variant}>{status.label}</Badge>
                        </div>
                      </div>
                    );
                  })}
                </div>
                <Button
                  render={<Link href="/credits" />}
                  variant="outline"
                  size="sm"
                  className="mt-3"
                >
                  View all credits
                </Button>
              </CardContent>
            </Card>
          )}

          {/* Monthly Chart */}
          {monthly.length > 0 && (
            <Card className="mt-6">
              <CardHeader>
                <CardTitle>Monthly Usage</CardTitle>
                <CardDescription>Credits used vs available</CardDescription>
              </CardHeader>
              <CardContent>
                <MonthlyChart data={monthly} />
              </CardContent>
            </Card>
          )}

          {/* Card Grid */}
          <h2 className="mt-10 text-lg font-semibold">Cards</h2>
          <div className="mt-4 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {d.cards.map((card) => {
              const pct = card.available_cents
                ? (card.used_cents / card.available_cents) * 100
                : 0;
              return (
                <Card key={card.user_card_id}>
                  <CardHeader>
                    <CardTitle className="text-base">
                      {card.card_name}
                    </CardTitle>
                    <CardDescription>{card.issuer}</CardDescription>
                  </CardHeader>
                  <CardContent>
                    <div className="flex items-baseline justify-between">
                      <span className="text-xs text-muted-foreground">
                        {card.credits_used}/{card.credits_total} credits
                      </span>
                      <span className="text-sm font-medium">
                        <Dollars cents={card.used_cents} className="text-sm" />
                        <span className="text-muted-foreground">/</span>
                        <Dollars
                          cents={card.available_cents}
                          className="text-sm"
                        />
                      </span>
                    </div>
                    <Progress value={pct} className="mt-2" />
                  </CardContent>
                </Card>
              );
            })}
          </div>

          {/* Recent Credits */}
          <h2 className="mt-10 text-lg font-semibold">Recent Credits</h2>
          <Card className="mt-4">
            <CardContent>
              {allCredits.length === 0 ? (
                <p className="py-4 text-center text-sm text-muted-foreground">
                  No credits to show yet.
                </p>
              ) : (
                <div className="divide-y divide-border">
                  {allCredits.slice(0, 8).map((credit) => {
                    const status = creditStatus(credit);
                    return (
                      <div
                        key={credit.id}
                        className="flex items-center justify-between py-3"
                      >
                        <div>
                          <p className="text-sm font-medium">
                            {credit.credit_definition.name}
                          </p>
                          <p className="text-xs text-muted-foreground">
                            {credit.card_name}
                          </p>
                        </div>
                        <div className="flex items-center gap-3">
                          <Dollars
                            cents={credit.credit_definition.amount_cents}
                            className="text-sm font-medium"
                          />
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
      )}
    </>
  );
}
