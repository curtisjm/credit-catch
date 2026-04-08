import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

export default function CardsPage() {
  return (
    <>
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-gray-900">Cards</h1>
      </div>

      <Card className="mt-6">
        <CardHeader>
          <CardTitle>Your Credit Cards</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-gray-500">
            No cards added yet. Add a credit card to start tracking rewards.
          </p>
        </CardContent>
      </Card>
    </>
  );
}
