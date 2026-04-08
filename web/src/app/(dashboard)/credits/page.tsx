import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

export default function CreditsPage() {
  return (
    <>
      <h1 className="text-2xl font-bold text-foreground">Credits</h1>

      <Card className="mt-6">
        <CardHeader>
          <CardTitle>Credit History</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-muted-foreground">
            No credit activity yet. Credits will appear here once you start
            tracking cards.
          </p>
        </CardContent>
      </Card>
    </>
  );
}
