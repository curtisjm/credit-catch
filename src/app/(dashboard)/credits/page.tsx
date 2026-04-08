import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

export default function CreditsPage() {
  return (
    <>
      <h1 className="text-2xl font-bold text-gray-900">Credits</h1>

      <Card className="mt-6">
        <CardHeader>
          <CardTitle>Credit History</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-gray-500">
            No credit activity yet. Credits will appear here once you start
            tracking cards.
          </p>
        </CardContent>
      </Card>
    </>
  );
}
