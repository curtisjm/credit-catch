import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Progress } from "@/components/ui/progress";
import { mockDashboard, mockCredits } from "@/lib/mock-data";

function formatDollars(cents: number): string {
  return `$${(cents / 100).toFixed(2)}`;
}

const statusBadge = {
  used: { variant: "success" as const, label: "Used" },
  expiring: { variant: "warning" as const, label: "Expiring" },
  expired: { variant: "danger" as const, label: "Expired" },
  upcoming: { variant: "default" as const, label: "Upcoming" },
};

export default function DashboardPage() {
  const d = mockDashboard;

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
          <div className="divide-y divide-border">
            {mockCredits.map((credit) => {
              const badge = statusBadge[credit.status];
              return (
                <div
                  key={credit.id}
                  className="flex items-center justify-between py-3"
                >
                  <div>
                    <p className="text-sm font-medium text-foreground">
                      {credit.credit_name}
                    </p>
                    <p className="text-xs text-muted-foreground">
                      {credit.card_name}
                    </p>
                  </div>
                  <div className="flex items-center gap-3">
                    <span className="text-sm font-medium text-foreground">
                      {formatDollars(credit.amount_cents)}
                    </span>
                    <Badge variant={badge.variant}>{badge.label}</Badge>
                  </div>
                </div>
              );
            })}
          </div>
        </CardContent>
      </Card>
    </>
  );
}
